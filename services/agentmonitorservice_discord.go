package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ─── Discord channel_type poller ───────────────────────────────────────────────
//
// For agent_channels with channel_type="discord", AgentMonitorService.startPoller()
// launches discordPollLoop in a goroutine.
//
// Discord has no long-polling endpoint like Telegram's getUpdates; instead we
// short-poll GET /channels/{id}/messages?after={last_id} for each configured
// channel_id every 2 seconds. This is sufficient for interactive chat and stays
// within Discord's rate limits (5 req/5 s per route).
//
// Required config fields (set in AgentChannel.ChannelConfig):
//   channel_ids — comma-separated Discord channel IDs to monitor (text/DM channels)
//   bot_token   — Discord bot token (from Developer Portal)

// ─── Discord API response structures ─────────────────────────────────────────

type dcMessage struct {
	ID      string   `json:"id"`
	Content string   `json:"content"`
	Author  dcAuthor `json:"author"`
}

type dcAuthor struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Bot      bool   `json:"bot"`
}

// ─── discordPollLoop ──────────────────────────────────────────────────────────

// discordPollLoop polls each configured Discord channel for new messages and
// calls AgentChannelService.HandleInbound for each human-authored text message.
// It sends the agent response back to the same Discord channel via the Messages API.
func (s *AgentMonitorService) discordPollLoop(ctx context.Context, entry *channelMonitorEntry) {
	token := entry.config["bot_token"]
	if token == "" {
		s.iLog.Error(fmt.Sprintf(
			"discordPollLoop: bot_token not configured for channel=%s — poller disabled",
			entry.channelName,
		))
		return
	}

	channelIDsRaw := entry.config["channel_ids"]
	if channelIDsRaw == "" {
		s.iLog.Error(fmt.Sprintf(
			"discordPollLoop: channel_ids not configured for channel=%s — poller disabled",
			entry.channelName,
		))
		return
	}

	discordChannels := []string{}
	for _, cid := range strings.Split(channelIDsRaw, ",") {
		cid = strings.TrimSpace(cid)
		if cid != "" {
			discordChannels = append(discordChannels, cid)
		}
	}
	if len(discordChannels) == 0 {
		s.iLog.Error(fmt.Sprintf("discordPollLoop: no valid channel_ids for channel=%s", entry.channelName))
		return
	}

	authHeader := "Bot " + token
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// lastMessageID tracks the highest seen message ID per Discord channel (snowflake).
	lastMessageID := make(map[string]string, len(discordChannels))

	// Verify bot token and send startup notification
	if s.discordVerifyBot(ctx, httpClient, authHeader, entry) {
		startMsg := entry.startMessage
		s.notifyDiscordChannels(ctx, httpClient, authHeader, discordChannels, startMsg)
	}

	s.iLog.Info(fmt.Sprintf(
		"discordPollLoop: started for channel=%s agent=%s monitoring %d Discord channel(s)",
		entry.channelName, entry.agentName, len(discordChannels),
	))

	defer func() {
		stopMsg := entry.stopMessage
		s.notifyDiscordChannels(context.Background(), httpClient, authHeader, discordChannels, stopMsg)
		s.iLog.Info(fmt.Sprintf("discordPollLoop: stopped for channel=%s", entry.channelName))
	}()

	// Use a WaitGroup to fan-out per-Discord-channel polling goroutines,
	// all controlled by the same ctx so they stop together.
	var wg sync.WaitGroup
	for _, dcChannelID := range discordChannels {
		wg.Add(1)
		go func(dcChanID string) {
			defer wg.Done()
			s.discordChannelPollLoop(ctx, entry, httpClient, authHeader, dcChanID, lastMessageID)
		}(dcChannelID)
	}
	wg.Wait()
}

// discordChannelPollLoop polls a single Discord channel for new messages.
func (s *AgentMonitorService) discordChannelPollLoop(
	ctx context.Context,
	entry *channelMonitorEntry,
	httpClient *http.Client,
	authHeader string,
	dcChannelID string,
	lastIDs map[string]string, // shared map — only this goroutine writes key dcChannelID
) {
	s.iLog.Info(fmt.Sprintf(
		"discordChannelPollLoop: monitoring discord_channel=%s channel=%s",
		dcChannelID, entry.channelName,
	))

	// Bootstrap: fetch the latest message ID so we only process NEW messages.
	if bootID := s.discordLatestMessageID(ctx, httpClient, authHeader, dcChannelID); bootID != "" {
		lastIDs[dcChannelID] = bootID
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		afterID := lastIDs[dcChannelID]
		msgs, err := s.discordGetNewMessages(ctx, httpClient, authHeader, dcChannelID, afterID)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			s.iLog.Error(fmt.Sprintf(
				"discordChannelPollLoop: poll error dc_channel=%s: %v",
				dcChannelID, err,
			))
			time.Sleep(5 * time.Second)
			continue
		}

		for _, msg := range msgs {
			// Track highest seen message ID (Discord snowflakes are lexicographically sortable).
			if msg.ID > lastIDs[dcChannelID] {
				lastIDs[dcChannelID] = msg.ID
			}

			// Skip bot messages (avoid infinite loops) and empty content.
			if msg.Author.Bot || strings.TrimSpace(msg.Content) == "" {
				continue
			}

			senderID := msg.Author.ID
			senderName := msg.Author.Username
			text := strings.TrimSpace(msg.Content)

			s.iLog.Info(fmt.Sprintf(
				"discordChannelPollLoop: message from sender=%s name=%s dc_channel=%s channel=%s text=%q",
				senderID, senderName, dcChannelID, entry.channelName, text,
			))

			// Process concurrently per sender, but serialize same-sender messages
			// (serialisation is handled inside HandleInbound via senderMu).
			go func(sID, dcCID, sName, msgText string) {
				response, err := s.channelSvc.HandleInbound(ctx, entry.channelID, sID, sName, msgText)
				if err != nil {
					s.iLog.Error(fmt.Sprintf(
						"discordChannelPollLoop: HandleInbound error channel=%s sender=%s: %v",
						entry.channelName, sID, err,
					))
					response = "Sorry, an error occurred processing your message. Please try again."
				}
				if response != "" {
					s.sendDiscordMessage(context.Background(), httpClient, authHeader, dcCID, response)
				}
			}(senderID, dcChannelID, senderName, text)
		}
	}
}

// discordLatestMessageID fetches the ID of the most recent message in a channel
// so the poller only processes new messages after startup.
func (s *AgentMonitorService) discordLatestMessageID(ctx context.Context, client *http.Client, authHeader, dcChannelID string) string {
	apiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages?limit=1", dcChannelID)
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
	var msgs []dcMessage
	if err := json.Unmarshal(body, &msgs); err != nil || len(msgs) == 0 {
		return ""
	}
	return msgs[0].ID
}

// discordGetNewMessages fetches messages after afterID in ascending order.
func (s *AgentMonitorService) discordGetNewMessages(ctx context.Context, client *http.Client, authHeader, dcChannelID, afterID string) ([]dcMessage, error) {
	apiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages?limit=100", dcChannelID)
	if afterID != "" {
		apiURL += "&after=" + afterID
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

	if resp.StatusCode == http.StatusTooManyRequests {
		// Back off on rate limit
		time.Sleep(5 * time.Second)
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Discord API %d: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var msgs []dcMessage
	if err := json.Unmarshal(body, &msgs); err != nil {
		return nil, err
	}

	// Discord returns newest-first when using `after`; reverse to chronological order.
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

// sendDiscordMessage sends a text reply to a Discord channel.
func (s *AgentMonitorService) sendDiscordMessage(ctx context.Context, client *http.Client, authHeader, dcChannelID, text string) {
	apiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", dcChannelID)

	// Discord messages have a 2000-character limit; truncate if needed.
	const discordMaxLen = 2000
	if len(text) > discordMaxLen {
		text = text[:discordMaxLen-1] + "…"
	}

	payload, _ := json.Marshal(map[string]interface{}{"content": text})
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(payload)))
	if err != nil {
		s.iLog.Error(fmt.Sprintf("sendDiscordMessage: build request error dc_channel=%s: %v", dcChannelID, err))
		return
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("sendDiscordMessage: send error dc_channel=%s: %v", dcChannelID, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		s.iLog.Error(fmt.Sprintf("sendDiscordMessage: Discord API error %d dc_channel=%s: %s", resp.StatusCode, dcChannelID, string(body)))
		return
	}
	s.iLog.Info(fmt.Sprintf("sendDiscordMessage: delivered to dc_channel=%s", dcChannelID))
}

// discordVerifyBot calls GET /users/@me to verify the bot token.
func (s *AgentMonitorService) discordVerifyBot(ctx context.Context, client *http.Client, authHeader string, entry *channelMonitorEntry) bool {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://discord.com/api/v10/users/@me", nil)
	req.Header.Set("Authorization", authHeader)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		s.iLog.Error(fmt.Sprintf("discordVerifyBot: token verification failed for channel=%s: %v", entry.channelName, err))
		return false
	}
	resp.Body.Close()
	s.iLog.Info(fmt.Sprintf("discordVerifyBot: bot token verified for channel=%s", entry.channelName))
	return true
}

// notifyDiscordChannels sends a notification to each configured Discord channel.
func (s *AgentMonitorService) notifyDiscordChannels(ctx context.Context, client *http.Client, authHeader string, channelIDs []string, text string) {
	for _, cid := range channelIDs {
		s.sendDiscordMessage(ctx, client, authHeader, cid, text)
	}
}
