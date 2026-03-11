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

// ─── Slack channel_type poller ─────────────────────────────────────────────────
//
// Short-polls conversations.history for each configured Slack channel every 2s.
// No Socket Mode required — uses the standard Web API with bot token.
//
// Required config fields (set in AgentChannel.ChannelConfig):
//   bot_token   — Slack bot token (xoxb-...)
//   channel_ids — comma-separated Slack channel IDs (C..., D... for DMs)
//
// The bot must be invited to each channel and granted scopes:
//   channels:history, groups:history, im:history, mpim:history, chat:write

// ─── Slack API response structures ───────────────────────────────────────────

type slkMessage struct {
	Type    string `json:"type"`
	User    string `json:"user"`
	BotID   string `json:"bot_id"`
	Subtype string `json:"subtype"`
	Text    string `json:"text"`
	TS      string `json:"ts"` // "1234567890.123456" — lexicographically sortable
}

type slkHistoryResponse struct {
	OK       bool         `json:"ok"`
	Messages []slkMessage `json:"messages"`
	HasMore  bool         `json:"has_more"`
	Error    string       `json:"error"`
}

// ─── slackPollLoop ────────────────────────────────────────────────────────────

func (s *AgentMonitorService) slackPollLoop(ctx context.Context, entry *channelMonitorEntry) {
	token := entry.config["bot_token"]
	if token == "" {
		s.iLog.Error(fmt.Sprintf(
			"slackPollLoop: bot_token not configured for channel=%s — poller disabled",
			entry.channelName,
		))
		return
	}

	channelIDsRaw := entry.config["channel_ids"]
	if channelIDsRaw == "" {
		s.iLog.Error(fmt.Sprintf(
			"slackPollLoop: channel_ids not configured for channel=%s — poller disabled",
			entry.channelName,
		))
		return
	}

	slackChannels := []string{}
	for _, cid := range strings.Split(channelIDsRaw, ",") {
		cid = strings.TrimSpace(cid)
		if cid != "" {
			slackChannels = append(slackChannels, cid)
		}
	}
	if len(slackChannels) == 0 {
		s.iLog.Error(fmt.Sprintf("slackPollLoop: no valid channel_ids for channel=%s", entry.channelName))
		return
	}

	authHeader := "Bearer " + token
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// Verify bot token; send startup notification if successful.
	if s.slackVerifyBot(ctx, httpClient, authHeader, entry) {
		startMsg := entry.startMessage
		s.notifySlackChannels(ctx, httpClient, authHeader, slackChannels, startMsg)
	}

	s.iLog.Info(fmt.Sprintf(
		"slackPollLoop: started for channel=%s agent=%s monitoring %d Slack channel(s)",
		entry.channelName, entry.agentName, len(slackChannels),
	))

	defer func() {
		stopMsg := entry.stopMessage
		s.notifySlackChannels(context.Background(), httpClient, authHeader, slackChannels, stopMsg)
		s.iLog.Info(fmt.Sprintf("slackPollLoop: stopped for channel=%s", entry.channelName))
	}()

	// lastTS tracks the highest seen message timestamp per Slack channel.
	// Slack timestamps are strings like "1234567890.123456" — lexicographically sortable.
	lastTS := make(map[string]string, len(slackChannels))

	// Bootstrap: record current latest TS so we only process NEW messages after startup.
	for _, slkChanID := range slackChannels {
		if ts := s.slackLatestTS(ctx, httpClient, authHeader, slkChanID); ts != "" {
			lastTS[slkChanID] = ts
		}
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		for _, slkChanID := range slackChannels {
			msgs, err := s.slackGetNewMessages(ctx, httpClient, authHeader, slkChanID, lastTS[slkChanID])
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				s.iLog.Error(fmt.Sprintf(
					"slackPollLoop: poll error slack_channel=%s channel=%s: %v",
					slkChanID, entry.channelName, err,
				))
				continue
			}

			for _, msg := range msgs {
				// Track latest TS (Slack strings compare lexicographically correctly).
				if msg.TS > lastTS[slkChanID] {
					lastTS[slkChanID] = msg.TS
				}

				// Skip bot/system messages and messages without a user.
				if msg.BotID != "" || msg.Subtype != "" || msg.User == "" {
					continue
				}

				text := strings.TrimSpace(msg.Text)
				if text == "" {
					continue
				}

				s.iLog.Info(fmt.Sprintf(
					"slackPollLoop: message from user=%s slack_channel=%s channel=%s text=%q",
					msg.User, slkChanID, entry.channelName, text,
				))

				go func(sID, slkCID, msgText string) {
					response, err := s.channelSvc.HandleInbound(ctx, entry.channelID, sID, sID, msgText)
					if err != nil {
						s.iLog.Error(fmt.Sprintf(
							"slackPollLoop: HandleInbound error channel=%s sender=%s: %v",
							entry.channelName, sID, err,
						))
						response = "Sorry, an error occurred processing your message. Please try again."
					}
					if response != "" {
						s.sendSlackMessage(context.Background(), httpClient, authHeader, slkCID, response)
					}
				}(msg.User, slkChanID, text)
			}
		}
	}
}

// slackVerifyBot calls auth.test to validate the bot token.
func (s *AgentMonitorService) slackVerifyBot(ctx context.Context, client *http.Client, authHeader string, entry *channelMonitorEntry) bool {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://slack.com/api/auth.test", nil)
	req.Header.Set("Authorization", authHeader)
	resp, err := client.Do(req)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("slackVerifyBot: request failed for channel=%s: %v", entry.channelName, err))
		return false
	}
	defer resp.Body.Close()
	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result) //nolint:errcheck
	if !result.OK {
		s.iLog.Error(fmt.Sprintf("slackVerifyBot: auth.test failed for channel=%s: %s", entry.channelName, result.Error))
		return false
	}
	s.iLog.Info(fmt.Sprintf("slackVerifyBot: bot token verified for channel=%s", entry.channelName))
	return true
}

// slackLatestTS fetches the timestamp of the most recent message in a Slack channel
// so the poller only processes new messages after startup.
func (s *AgentMonitorService) slackLatestTS(ctx context.Context, client *http.Client, authHeader, slkChanID string) string {
	apiURL := "https://slack.com/api/conversations.history?channel=" + url.QueryEscape(slkChanID) + "&limit=1"
	req, _ := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	req.Header.Set("Authorization", authHeader)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result slkHistoryResponse
	if err := json.Unmarshal(body, &result); err != nil || !result.OK || len(result.Messages) == 0 {
		return ""
	}
	return result.Messages[0].TS
}

// slackGetNewMessages fetches messages newer than oldestTS in chronological order.
func (s *AgentMonitorService) slackGetNewMessages(ctx context.Context, client *http.Client, authHeader, slkChanID, oldestTS string) ([]slkMessage, error) {
	apiURL := "https://slack.com/api/conversations.history?channel=" + url.QueryEscape(slkChanID) + "&limit=100"
	if oldestTS != "" {
		apiURL += "&oldest=" + url.QueryEscape(oldestTS) + "&inclusive=false"
	}
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result slkHistoryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if !result.OK {
		if result.Error == "ratelimited" {
			time.Sleep(5 * time.Second)
			return nil, nil
		}
		return nil, fmt.Errorf("Slack API error: %s", result.Error)
	}

	// Slack returns newest-first; reverse to chronological order.
	msgs := result.Messages
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

// sendSlackMessage posts a text message to a Slack channel.
func (s *AgentMonitorService) sendSlackMessage(ctx context.Context, client *http.Client, authHeader, slkChanID, text string) {
	// Slack has a 4000-character limit per message; truncate if needed.
	const slackMaxLen = 4000
	if len(text) > slackMaxLen {
		text = text[:slackMaxLen-1] + "…"
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"channel": slkChanID,
		"text":    text,
	})
	req, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postMessage", strings.NewReader(string(payload)))
	if err != nil {
		s.iLog.Error(fmt.Sprintf("sendSlackMessage: build request error slack_channel=%s: %v", slkChanID, err))
		return
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("sendSlackMessage: send error slack_channel=%s: %v", slkChanID, err))
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	json.Unmarshal(body, &result) //nolint:errcheck
	if !result.OK {
		s.iLog.Error(fmt.Sprintf("sendSlackMessage: Slack API error slack_channel=%s: %s", slkChanID, result.Error))
		return
	}
	s.iLog.Info(fmt.Sprintf("sendSlackMessage: delivered to slack_channel=%s", slkChanID))
}

// notifySlackChannels sends a notification message to each configured Slack channel.
func (s *AgentMonitorService) notifySlackChannels(ctx context.Context, client *http.Client, authHeader string, channelIDs []string, text string) {
	for _, cid := range channelIDs {
		s.sendSlackMessage(ctx, client, authHeader, cid, text)
	}
}
