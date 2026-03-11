package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ─── iMessage channel_type poller (BlueBubbles) ───────────────────────────────
//
// Polls the BlueBubbles REST API for new iMessages every 3s.
// BlueBubbles must be running on a Mac: https://bluebubbles.app
//
// Required config fields:
//   bluebubbles_url      — base URL of the BlueBubbles server (e.g. http://192.168.1.10:1234)
//   bluebubbles_password — server password configured in BlueBubbles settings
//
// Authentication: password passed as ?password= query parameter on every request.

// ─── BlueBubbles API response structures ─────────────────────────────────────

type bbHandle struct {
	Address string `json:"address"`
}

type bbChat struct {
	GUID        string `json:"guid"`
	DisplayName string `json:"displayName"`
}

type bbMessage struct {
	GUID        string    `json:"guid"`
	Text        string    `json:"text"`
	IsFromMe    bool      `json:"isFromMe"`
	DateCreated int64     `json:"dateCreated"`
	Handle      *bbHandle `json:"handle"`
	Chats       []bbChat  `json:"chats"`
}

type bbMessagesResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    []bbMessage `json:"data"`
}

// ─── imessagePollLoop ─────────────────────────────────────────────────────────

func (s *AgentMonitorService) imessagePollLoop(ctx context.Context, entry *channelMonitorEntry) {
	baseURL := strings.TrimRight(entry.config["bluebubbles_url"], "/")
	password := entry.config["bluebubbles_password"]

	if baseURL == "" || password == "" {
		s.iLog.Error(fmt.Sprintf(
			"imessagePollLoop: bluebubbles_url/bluebubbles_password not configured for channel=%s — poller disabled",
			entry.channelName,
		))
		return
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}

	s.iLog.Info(fmt.Sprintf(
		"imessagePollLoop: started for channel=%s agent=%s server=%s",
		entry.channelName, entry.agentName, baseURL,
	))

	// activeChats tracks chatGUID for each sender address that has interacted; used
	// to broadcast the offline notification on shutdown.
	activeChats := make(map[string]string) // senderAddr -> chatGUID

	defer func() {
		if len(activeChats) > 0 {
			stopMsg := entry.stopMessage
			for _, chatG := range activeChats {
				s.sendBlueBubblesMessage(context.Background(), httpClient, baseURL, password, chatG, stopMsg)
			}
		}
		s.iLog.Info(fmt.Sprintf("imessagePollLoop: stopped for channel=%s", entry.channelName))
	}()

	// Bootstrap: record all current message GUIDs so we only process NEW messages.
	seenGUIDs := make(map[string]bool)
	if initMsgs, err := s.bbGetRecentMessages(ctx, httpClient, baseURL, password, 50); err == nil {
		for _, m := range initMsgs {
			seenGUIDs[m.GUID] = true
		}
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		msgs, err := s.bbGetRecentMessages(ctx, httpClient, baseURL, password, 25)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			s.iLog.Error(fmt.Sprintf("imessagePollLoop: poll error channel=%s: %v", entry.channelName, err))
			time.Sleep(5 * time.Second)
			continue
		}

		for _, msg := range msgs {
			if seenGUIDs[msg.GUID] {
				continue
			}
			seenGUIDs[msg.GUID] = true

			if msg.IsFromMe {
				continue // skip messages we sent
			}

			text := strings.TrimSpace(msg.Text)
			if text == "" {
				continue
			}

			senderAddr := ""
			if msg.Handle != nil {
				senderAddr = msg.Handle.Address
			}
			if senderAddr == "" {
				continue
			}

			// Determine chat GUID for reply delivery.
			chatGUID := ""
			for _, ch := range msg.Chats {
				chatGUID = ch.GUID
				break
			}
			if chatGUID == "" {
				chatGUID = "iMessage;-;" + senderAddr
			}

			s.iLog.Info(fmt.Sprintf(
				"imessagePollLoop: message from sender=%s chat=%s channel=%s text=%q",
				senderAddr, chatGUID, entry.channelName, text,
			))

			// Remember this chat so we can notify on shutdown.
			activeChats[senderAddr] = chatGUID

			go func(sAddr, chatG, msgText string) {
				response, err := s.channelSvc.HandleInbound(ctx, entry.channelID, sAddr, sAddr, msgText)
				if err != nil {
					s.iLog.Error(fmt.Sprintf(
						"imessagePollLoop: HandleInbound error channel=%s sender=%s: %v",
						entry.channelName, sAddr, err,
					))
					response = "Sorry, an error occurred processing your message. Please try again."
				}
				if response != "" {
					s.sendBlueBubblesMessage(context.Background(), httpClient, baseURL, password, chatG, response)
				}
			}(senderAddr, chatGUID, text)
		}
	}
}

// bbGetRecentMessages fetches the most recent messages from BlueBubbles.
func (s *AgentMonitorService) bbGetRecentMessages(ctx context.Context, client *http.Client, baseURL, password string, limit int) ([]bbMessage, error) {
	apiURL := fmt.Sprintf(
		"%s/api/v1/message?limit=%d&sort=DESC&password=%s",
		baseURL, limit, url.QueryEscape(password),
	)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("BlueBubbles API %d: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var result bbMessagesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// sendBlueBubblesMessage sends a text reply via BlueBubbles REST API.
func (s *AgentMonitorService) sendBlueBubblesMessage(ctx context.Context, client *http.Client, baseURL, password, chatGUID, text string) {
	apiURL := fmt.Sprintf("%s/api/v1/message/text?password=%s", baseURL, url.QueryEscape(password))
	payload, _ := json.Marshal(map[string]interface{}{
		"chatGuid": chatGUID,
		"message":  text,
		"method":   "apple-script",
	})

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(payload)))
	if err != nil {
		s.iLog.Error(fmt.Sprintf("sendBlueBubblesMessage: build request error chatGuid=%s: %v", chatGUID, err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("sendBlueBubblesMessage: send error chatGuid=%s: %v", chatGUID, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		s.iLog.Error(fmt.Sprintf(
			"sendBlueBubblesMessage: API error %d chatGuid=%s: %s",
			resp.StatusCode, chatGUID, string(body),
		))
		return
	}
	s.iLog.Info(fmt.Sprintf("sendBlueBubblesMessage: delivered to chatGuid=%s", chatGUID))
}
