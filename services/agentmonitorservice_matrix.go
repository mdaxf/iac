package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

// ─── Matrix channel_type poller ───────────────────────────────────────────────
//
// Long-polls the Matrix Client-Server sync API with a 30s timeout.
// This is the standard Matrix homeserver polling mechanism — no WebSocket required.
//
// Required config fields:
//   homeserver_url — Matrix homeserver URL (e.g. https://matrix.example.com)
//   user_id        — Bot Matrix ID (e.g. @bot:example.com) — used to filter own messages
//   access_token   — Matrix access token for the bot account
//   room_ids       — comma-separated Matrix room IDs to monitor (optional;
//                    if empty all joined rooms are monitored)

// ─── Matrix API response structures ──────────────────────────────────────────

type mxSyncResponse struct {
	NextBatch string `json:"next_batch"`
	Rooms     *struct {
		Join map[string]mxJoinedRoom `json:"join"`
	} `json:"rooms"`
}

type mxJoinedRoom struct {
	Timeline mxTimeline `json:"timeline"`
}

type mxTimeline struct {
	Events []mxEvent `json:"events"`
}

type mxEvent struct {
	EventID        string    `json:"event_id"`
	Type           string    `json:"type"`
	Sender         string    `json:"sender"`
	Content        mxContent `json:"content"`
	OriginServerTS int64     `json:"origin_server_ts"`
}

type mxContent struct {
	MsgType string `json:"msgtype"`
	Body    string `json:"body"`
}

// ─── matrixPollLoop ───────────────────────────────────────────────────────────

func (s *AgentMonitorService) matrixPollLoop(ctx context.Context, entry *channelMonitorEntry) {
	homeserver := strings.TrimRight(entry.config["homeserver_url"], "/")
	userID := entry.config["user_id"]
	accessToken := entry.config["access_token"]
	roomIDsRaw := entry.config["room_ids"]

	if homeserver == "" || accessToken == "" {
		s.iLog.Error(fmt.Sprintf(
			"matrixPollLoop: homeserver_url/access_token not configured for channel=%s — poller disabled",
			entry.channelName,
		))
		return
	}

	// Build room filter map (empty = all joined rooms).
	roomFilter := map[string]bool{}
	if roomIDsRaw != "" {
		for _, rid := range strings.Split(roomIDsRaw, ",") {
			rid = strings.TrimSpace(rid)
			if rid != "" {
				roomFilter[rid] = true
			}
		}
	}

	authHeader := "Bearer " + accessToken
	// HTTP client timeout must exceed the sync long-poll window (30s).
	httpClient := &http.Client{Timeout: 45 * time.Second}

	// Send startup notification to configured rooms.
	if len(roomFilter) > 0 {
		startMsg := entry.startMessage
		for roomID := range roomFilter {
			s.sendMatrixMessage(context.Background(), httpClient, homeserver, authHeader, roomID, startMsg)
		}
	}

	s.iLog.Info(fmt.Sprintf(
		"matrixPollLoop: started for channel=%s agent=%s homeserver=%s rooms=%d",
		entry.channelName, entry.agentName, homeserver, len(roomFilter),
	))

	defer func() {
		if len(roomFilter) > 0 {
			stopMsg := entry.stopMessage
			for roomID := range roomFilter {
				s.sendMatrixMessage(context.Background(), httpClient, homeserver, authHeader, roomID, stopMsg)
			}
		}
		s.iLog.Info(fmt.Sprintf("matrixPollLoop: stopped for channel=%s", entry.channelName))
	}()

	// Bootstrap: perform an initial sync with limit=0 to get a next_batch token
	// without replaying existing history.
	var nextBatch string
	filterJSON, _ := json.Marshal(map[string]interface{}{
		"room": map[string]interface{}{
			"timeline": map[string]int{"limit": 0},
		},
	})
	initialURL := fmt.Sprintf(
		"%s/_matrix/client/v3/sync?access_token=%s&filter=%s",
		homeserver, url.QueryEscape(accessToken), url.QueryEscape(string(filterJSON)),
	)
	if req, err := http.NewRequestWithContext(ctx, "GET", initialURL, nil); err == nil {
		if resp, err := httpClient.Do(req); err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			var initResult mxSyncResponse
			if json.Unmarshal(body, &initResult) == nil && initResult.NextBatch != "" {
				nextBatch = initResult.NextBatch
				s.iLog.Info(fmt.Sprintf("matrixPollLoop: bootstrap next_batch=%s channel=%s", nextBatch, entry.channelName))
			}
		}
	}

	var txnCounter int64

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		syncURL := fmt.Sprintf(
			"%s/_matrix/client/v3/sync?access_token=%s&timeout=30000",
			homeserver, url.QueryEscape(accessToken),
		)
		if nextBatch != "" {
			syncURL += "&since=" + url.QueryEscape(nextBatch)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", syncURL, nil)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("matrixPollLoop: build request error channel=%s: %v", entry.channelName, err))
			time.Sleep(5 * time.Second)
			continue
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			s.iLog.Error(fmt.Sprintf("matrixPollLoop: sync error channel=%s: %v", entry.channelName, err))
			time.Sleep(5 * time.Second)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var result mxSyncResponse
		if err := json.Unmarshal(body, &result); err != nil {
			s.iLog.Error(fmt.Sprintf("matrixPollLoop: JSON parse error channel=%s: %v", entry.channelName, err))
			time.Sleep(2 * time.Second)
			continue
		}

		if result.NextBatch != "" {
			nextBatch = result.NextBatch
		}

		if result.Rooms == nil {
			continue
		}

		for roomID, room := range result.Rooms.Join {
			// Skip rooms not in the filter (if a filter is configured).
			if len(roomFilter) > 0 && !roomFilter[roomID] {
				continue
			}

			for _, event := range room.Timeline.Events {
				// Only process m.room.message events from other users.
				if event.Type != "m.room.message" || event.Sender == userID {
					continue
				}
				if event.Content.MsgType != "m.text" {
					continue
				}

				text := strings.TrimSpace(event.Content.Body)
				if text == "" {
					continue
				}

				s.iLog.Info(fmt.Sprintf(
					"matrixPollLoop: message from sender=%s room=%s channel=%s text=%q",
					event.Sender, roomID, entry.channelName, text,
				))

				txn := atomic.AddInt64(&txnCounter, 1)

				go func(senderID, rID, msgText string, txnNum int64) {
					response, err := s.channelSvc.HandleInbound(ctx, entry.channelID, senderID, senderID, msgText)
					if err != nil {
						s.iLog.Error(fmt.Sprintf(
							"matrixPollLoop: HandleInbound error channel=%s sender=%s: %v",
							entry.channelName, senderID, err,
						))
						response = "Sorry, an error occurred processing your message. Please try again."
					}
					if response != "" {
						txnID := fmt.Sprintf("iac_%d_%d", txnNum, time.Now().UnixMilli())
						s.sendMatrixMessage(context.Background(), httpClient, homeserver, authHeader, rID, response, txnID)
					}
				}(event.Sender, roomID, text, txn)
			}
		}
	}
}

// sendMatrixMessage sends a text message to a Matrix room.
// txnIDOpt is an optional transaction ID for idempotency; one is generated if not provided.
func (s *AgentMonitorService) sendMatrixMessage(ctx context.Context, client *http.Client, homeserver, authHeader, roomID, text string, txnIDOpt ...string) {
	txnID := fmt.Sprintf("iac_%d", time.Now().UnixNano())
	if len(txnIDOpt) > 0 && txnIDOpt[0] != "" {
		txnID = txnIDOpt[0]
	}

	apiURL := fmt.Sprintf(
		"%s/_matrix/client/v3/rooms/%s/send/m.room.message/%s",
		homeserver, url.PathEscape(roomID), url.PathEscape(txnID),
	)
	payload, _ := json.Marshal(map[string]interface{}{
		"msgtype": "m.text",
		"body":    text,
	})

	req, err := http.NewRequestWithContext(ctx, "PUT", apiURL, strings.NewReader(string(payload)))
	if err != nil {
		s.iLog.Error(fmt.Sprintf("sendMatrixMessage: build request error room=%s: %v", roomID, err))
		return
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("sendMatrixMessage: send error room=%s: %v", roomID, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.iLog.Error(fmt.Sprintf(
			"sendMatrixMessage: Matrix API error %d room=%s: %s",
			resp.StatusCode, roomID, string(body),
		))
		return
	}
	s.iLog.Info(fmt.Sprintf("sendMatrixMessage: delivered to room=%s", roomID))
}
