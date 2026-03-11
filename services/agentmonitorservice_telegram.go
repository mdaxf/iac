package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ─── Telegram channel_type poller ─────────────────────────────────────────────
//
// When an agent_channel with channel_type="telegram" is bound to a channel_monitor
// agent, AgentMonitorService.startPoller() launches telegramPollLoop in a goroutine.
//
// It long-polls the Telegram Bot API (getUpdates) — no webhook setup required —
// and delegates each inbound text message to AgentChannelService.HandleInbound.
// The agent's response text is delivered back to the same Telegram chat via sendMessage.

// ─── Telegram API response structures ────────────────────────────────────────

type tgGetUpdatesResponse struct {
	OK     bool       `json:"ok"`
	Result []tgUpdate `json:"result"`
}

type tgUpdate struct {
	UpdateID int64      `json:"update_id"`
	Message  *tgMessage `json:"message"`
}

type tgMessage struct {
	MessageID int64   `json:"message_id"`
	From      tgUser  `json:"from"`
	Chat      tgChat  `json:"chat"`
	Text      string  `json:"text"`
}

type tgUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type tgChat struct {
	ID int64 `json:"id"`
}

// ─── startPoller ─────────────────────────────────────────────────────────────

// startPoller launches a channel-type specific polling goroutine for the given entry.
//
// Native pollers (IAC connects outbound to the platform API):
//   - telegram   — long-polls getUpdates API
//   - discord    — short-polls Messages API every 2 s per channel_id
//   - slack      — short-polls conversations.history per channel_id every 2 s
//   - signal     — polls signal-cli REST API every 2 s
//   - imessage   — polls BlueBubbles REST API every 3 s
//   - mattermost — polls posts?since every 2 s per channel_id
//   - matrix     — long-polls /_matrix/client/v3/sync with 30 s timeout
//   - irc        — persistent TCP/TLS connection; reads PRIVMSG events
//
// Webhook mode (platform pushes events to IAC; no outbound polling needed):
//   - line      — configure LINE console webhook URL to POST /api/agentchannels/:id/line-webhook
//   - whatsapp  — configure Meta webhook URL to GET/POST /api/agentchannels/:id/whatsapp-webhook
//   - msteams   — configure Bot Framework endpoint to POST /api/agentchannels/:id/teams-webhook
//
// OpenClaw bridge (all other channels):
// For channel types without a native poller, OpenClaw pushes inbound messages to
// POST /api/agentchannels/:id/inbound.
func (s *AgentMonitorService) startPoller(ctx context.Context, entry *channelMonitorEntry) {
	// launchPoller wraps a poller goroutine so its lifetime is tracked by pollerWg.
	// StopAll() waits on pollerWg before the process exits, ensuring stop messages
	// are delivered before os.Exit is called.
	launchPoller := func(fn func()) {
		s.pollerWg.Add(1)
		go func() {
			defer s.pollerWg.Done()
			fn()
		}()
	}

	switch entry.channelType {
	case "telegram":
		launchPoller(func() { s.telegramPollLoop(ctx, entry) })
	case "discord":
		launchPoller(func() { s.discordPollLoop(ctx, entry) })
	case "slack":
		launchPoller(func() { s.slackPollLoop(ctx, entry) })
	case "signal":
		launchPoller(func() { s.signalPollLoop(ctx, entry) })
	case "imessage":
		launchPoller(func() { s.imessagePollLoop(ctx, entry) })
	case "mattermost":
		launchPoller(func() { s.mattermostPollLoop(ctx, entry) })
	case "matrix":
		launchPoller(func() { s.matrixPollLoop(ctx, entry) })
	case "irc":
		launchPoller(func() { s.ircPollLoop(ctx, entry) })
	case "line":
		s.iLog.Info(fmt.Sprintf(
			"AgentMonitorService: LINE channel=%s agent=%s — webhook mode: configure LINE Developers console webhook URL to POST /api/agentchannels/%s/line-webhook",
			entry.channelName, entry.agentName, entry.channelID,
		))
	case "whatsapp":
		s.iLog.Info(fmt.Sprintf(
			"AgentMonitorService: WhatsApp channel=%s agent=%s — webhook mode: configure Meta webhook to GET/POST /api/agentchannels/%s/whatsapp-webhook",
			entry.channelName, entry.agentName, entry.channelID,
		))
	case "msteams":
		s.iLog.Info(fmt.Sprintf(
			"AgentMonitorService: Teams channel=%s agent=%s — webhook mode: configure Bot Framework messaging endpoint to POST /api/agentchannels/%s/teams-webhook",
			entry.channelName, entry.agentName, entry.channelID,
		))
	default:
		s.iLog.Info(fmt.Sprintf(
			"AgentMonitorService: channel_type=%s uses OpenClaw bridge — "+
				"inbound messages arrive via POST /api/agentchannels/%s/inbound",
			entry.channelType, entry.channelID,
		))
	}
}

// ─── telegramPollLoop ────────────────────────────────────────────────────────

// telegramPollLoop long-polls the Telegram Bot API and processes each inbound
// text message by calling AgentChannelService.HandleInbound, then sends the
// agent's response back via sendMessage.
//
// On startup: verifies the bot token and notifies all authorized channel users
// that the agent is online.
// On shutdown: sends a disconnect notice to those same users.
func (s *AgentMonitorService) telegramPollLoop(ctx context.Context, entry *channelMonitorEntry) {
	token := entry.config["bot_token"]
	if token == "" {
		s.iLog.Error(fmt.Sprintf(
			"telegramPollLoop: bot_token not configured for channel=%s — poller disabled",
			entry.channelName,
		))
		return
	}

	baseURL := fmt.Sprintf("https://api.telegram.org/bot%s", token)
	var offset int64

	// Use a client with a longer timeout than the getUpdates long-poll window (30 s).
	httpClient := &http.Client{Timeout: 45 * time.Second}

	// Verify bot token with getMe before starting the loop.
	getMeURL := baseURL + "/getMe"
	getMeReq, _ := http.NewRequestWithContext(ctx, "GET", getMeURL, nil)
	if getMeResp, err := httpClient.Do(getMeReq); err == nil {
		getMeResp.Body.Close()
		s.iLog.Info(fmt.Sprintf(
			"telegramPollLoop: bot token verified for channel=%s agent=%s",
			entry.channelName, entry.agentName,
		))
		// Notify authorized users that communication is active.
		startMsg := entry.startMessage
		s.notifyTelegramChannelUsers(token, entry.channelID, startMsg)
	} else {
		s.iLog.Error(fmt.Sprintf(
			"telegramPollLoop: failed to verify bot token for channel=%s: %v",
			entry.channelName, err,
		))
	}

	s.iLog.Info(fmt.Sprintf(
		"telegramPollLoop: started for channel=%s agent=%s",
		entry.channelName, entry.agentName,
	))

	// Ensure shutdown notification is sent when the loop exits.
	defer func() {
		stopMsg := entry.stopMessage
		s.notifyTelegramChannelUsers(token, entry.channelID, stopMsg)
		s.iLog.Info(fmt.Sprintf("telegramPollLoop: stopped for channel=%s", entry.channelName))
	}()

	for {
		// Exit immediately if context was cancelled (channel removed / service stopped).
		select {
		case <-ctx.Done():
			return
		default:
		}

		pollURL := fmt.Sprintf(
			"%s/getUpdates?offset=%d&timeout=30&allowed_updates=message",
			baseURL, offset,
		)
		req, err := http.NewRequestWithContext(ctx, "GET", pollURL, nil)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("telegramPollLoop: failed to build request channel=%s: %v", entry.channelName, err))
			time.Sleep(5 * time.Second)
			continue
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return // context cancelled — defer will send shutdown notice
			}
			s.iLog.Error(fmt.Sprintf("telegramPollLoop: getUpdates error channel=%s: %v", entry.channelName, err))
			time.Sleep(5 * time.Second)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var result tgGetUpdatesResponse
		if err := json.Unmarshal(body, &result); err != nil {
			s.iLog.Error(fmt.Sprintf(
				"telegramPollLoop: JSON parse error channel=%s: %v | body: %s",
				entry.channelName, err, string(body),
			))
			time.Sleep(2 * time.Second)
			continue
		}

		if !result.OK {
			s.iLog.Error(fmt.Sprintf(
				"telegramPollLoop: Telegram API ok=false channel=%s: %s",
				entry.channelName, string(body),
			))
			time.Sleep(5 * time.Second)
			continue
		}

		for _, update := range result.Result {
			// Advance offset unconditionally — never re-process an update.
			offset = update.UpdateID + 1

			if update.Message == nil || update.Message.Text == "" {
				continue // skip non-text updates (photos, stickers, etc.)
			}

			msg := update.Message
			senderID := fmt.Sprintf("%d", msg.From.ID)
			chatID := fmt.Sprintf("%d", msg.Chat.ID)

			// Build a human-readable display name, preferring @username.
			name := msg.From.FirstName
			if msg.From.LastName != "" {
				name += " " + msg.From.LastName
			}
			if msg.From.Username != "" {
				name = "@" + msg.From.Username
			}

			s.iLog.Info(fmt.Sprintf(
				"telegramPollLoop: message from sender=%s name=%s chatID=%s channel=%s text=%q",
				senderID, name, chatID, entry.channelName, msg.Text,
			))

			// Run HandleInbound in a goroutine so the poll loop immediately advances
			// the offset and can process the next update without blocking on agent execution.
			go func(sID, cID, sName, text string) {
				response, err := s.channelSvc.HandleInbound(ctx, entry.channelID, sID, sName, text)
				if err != nil {
					s.iLog.Error(fmt.Sprintf(
						"telegramPollLoop: HandleInbound error channel=%s sender=%s: %v",
						entry.channelName, sID, err,
					))
					response = "Sorry, an error occurred processing your message. Please try again."
				}
				if response != "" {
					s.sendTelegramMessage(token, cID, response)
				}
			}(senderID, chatID, name, msg.Text)
		}
	}
}

// notifyTelegramChannelUsers sends a notification message to all authorized,
// active users of the given channel. For Telegram, the channel_user_id is the
// Telegram user ID which doubles as the private chat ID for bot messages.
func (s *AgentMonitorService) notifyTelegramChannelUsers(token, channelID, text string) {
	if s.channelSvc == nil {
		return
	}
	mappings, err := s.channelSvc.ListMappings(channelID)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("notifyTelegramChannelUsers: failed to list mappings: %v", err))
		return
	}
	for _, m := range mappings {
		if m.Authorized && m.Active && m.ChannelUserID != "" {
			// Use background context so shutdown notifications still send after ctx cancellation.
			s.sendTelegramMessageCtx(context.Background(), token, m.ChannelUserID, text)
		}
	}
}

// sendTelegramMessage delivers a text reply to a Telegram chat using background context.
func (s *AgentMonitorService) sendTelegramMessage(token, chatID, text string) {
	s.sendTelegramMessageCtx(context.Background(), token, chatID, text)
}

// sendTelegramMessageCtx delivers a text reply to a Telegram chat with the given context.
func (s *AgentMonitorService) sendTelegramMessageCtx(ctx context.Context, token, chatID, text string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload, _ := json.Marshal(map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payload))
	if err != nil {
		s.iLog.Error(fmt.Sprintf("sendTelegramMessage: build request error chatID=%s: %v", chatID, err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("sendTelegramMessage: send error chatID=%s: %v", chatID, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.iLog.Error(fmt.Sprintf(
			"sendTelegramMessage: Telegram API error %d chatID=%s: %s",
			resp.StatusCode, chatID, string(body),
		))
		return
	}
	s.iLog.Info(fmt.Sprintf("sendTelegramMessage: delivered to chatID=%s", chatID))
}
