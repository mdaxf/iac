package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ─── Mattermost channel_type poller ──────────────────────────────────────────
//
// Short-polls the Mattermost REST API for new posts every 2s.
//
// Required config fields:
//   server_url  — Mattermost server base URL (e.g. https://mattermost.example.com)
//   bot_token   — Mattermost bot or personal access token
//   channel_ids — comma-separated Mattermost channel IDs

// ─── Mattermost API response structures ──────────────────────────────────────

type mmPost struct {
	ID        string `json:"id"`
	CreateAt  int64  `json:"create_at"` // milliseconds since epoch
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
	Message   string `json:"message"`
	Type      string `json:"type"` // "" for regular posts, "system_*" for system messages
}

type mmPostsResponse struct {
	Order []string          `json:"order"`
	Posts map[string]mmPost `json:"posts"`
}

type mmUserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// ─── mattermostPollLoop ───────────────────────────────────────────────────────

func (s *AgentMonitorService) mattermostPollLoop(ctx context.Context, entry *channelMonitorEntry) {
	serverURL := strings.TrimRight(entry.config["server_url"], "/")
	token := entry.config["bot_token"]
	channelIDsRaw := entry.config["channel_ids"]

	if serverURL == "" || token == "" || channelIDsRaw == "" {
		s.iLog.Error(fmt.Sprintf(
			"mattermostPollLoop: server_url/bot_token/channel_ids not configured for channel=%s — poller disabled",
			entry.channelName,
		))
		return
	}

	mmChannels := []string{}
	for _, cid := range strings.Split(channelIDsRaw, ",") {
		cid = strings.TrimSpace(cid)
		if cid != "" {
			mmChannels = append(mmChannels, cid)
		}
	}
	if len(mmChannels) == 0 {
		s.iLog.Error(fmt.Sprintf("mattermostPollLoop: no valid channel_ids for channel=%s", entry.channelName))
		return
	}

	authHeader := "Bearer " + token
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// Get our own user ID to filter out the bot's own messages.
	botUserID := s.mattermostGetSelfID(ctx, httpClient, serverURL, authHeader, entry)

	if botUserID != "" {
		startMsg := entry.startMessage
		s.notifyMattermostChannels(ctx, httpClient, serverURL, authHeader, mmChannels, startMsg)
	}

	s.iLog.Info(fmt.Sprintf(
		"mattermostPollLoop: started for channel=%s agent=%s monitoring %d Mattermost channel(s)",
		entry.channelName, entry.agentName, len(mmChannels),
	))

	defer func() {
		stopMsg := entry.stopMessage
		s.notifyMattermostChannels(context.Background(), httpClient, serverURL, authHeader, mmChannels, stopMsg)
		s.iLog.Info(fmt.Sprintf("mattermostPollLoop: stopped for channel=%s", entry.channelName))
	}()

	// Bootstrap: record current time so we only process NEW messages.
	lastSince := time.Now().UnixMilli() - 5000 // 5-second buffer for clock skew

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		pollSince := lastSince
		lastSince = time.Now().UnixMilli()

		for _, mmChanID := range mmChannels {
			posts, err := s.mattermostGetNewPosts(ctx, httpClient, serverURL, authHeader, mmChanID, pollSince)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				s.iLog.Error(fmt.Sprintf(
					"mattermostPollLoop: poll error mm_channel=%s channel=%s: %v",
					mmChanID, entry.channelName, err,
				))
				continue
			}

			for _, post := range posts {
				// Skip system messages and our own bot posts.
				if post.Type != "" || post.UserID == botUserID {
					continue
				}

				text := strings.TrimSpace(post.Message)
				if text == "" {
					continue
				}

				s.iLog.Info(fmt.Sprintf(
					"mattermostPollLoop: message from user=%s mm_channel=%s channel=%s text=%q",
					post.UserID, mmChanID, entry.channelName, text,
				))

				go func(sID, mmCID, msgText string) {
					response, err := s.channelSvc.HandleInbound(ctx, entry.channelID, sID, sID, msgText)
					if err != nil {
						s.iLog.Error(fmt.Sprintf(
							"mattermostPollLoop: HandleInbound error channel=%s sender=%s: %v",
							entry.channelName, sID, err,
						))
						response = "Sorry, an error occurred processing your message. Please try again."
					}
					if response != "" {
						s.sendMattermostMessage(context.Background(), httpClient, serverURL, authHeader, mmCID, response)
					}
				}(post.UserID, mmChanID, text)
			}
		}
	}
}

// mattermostGetSelfID retrieves the bot's own user ID to filter outbound messages.
func (s *AgentMonitorService) mattermostGetSelfID(ctx context.Context, client *http.Client, serverURL, authHeader string, entry *channelMonitorEntry) string {
	req, _ := http.NewRequestWithContext(ctx, "GET", serverURL+"/api/v4/users/me", nil)
	req.Header.Set("Authorization", authHeader)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		s.iLog.Error(fmt.Sprintf("mattermostGetSelfID: request failed for channel=%s: %v", entry.channelName, err))
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var user mmUserInfo
	json.Unmarshal(body, &user) //nolint:errcheck
	s.iLog.Info(fmt.Sprintf("mattermostGetSelfID: bot user_id=%s channel=%s", user.ID, entry.channelName))
	return user.ID
}

// mattermostGetNewPosts fetches posts created after `since` (milliseconds) in chronological order.
func (s *AgentMonitorService) mattermostGetNewPosts(ctx context.Context, client *http.Client, serverURL, authHeader, mmChanID string, since int64) ([]mmPost, error) {
	apiURL := fmt.Sprintf("%s/api/v4/channels/%s/posts?since=%d", serverURL, mmChanID, since)
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
		time.Sleep(5 * time.Second)
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Mattermost API %d: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var result mmPostsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Build sorted slice (order array lists IDs newest-first; reverse to chronological).
	posts := make([]mmPost, 0, len(result.Order))
	for _, id := range result.Order {
		if p, ok := result.Posts[id]; ok {
			posts = append(posts, p)
		}
	}
	for i, j := 0, len(posts)-1; i < j; i, j = i+1, j-1 {
		posts[i], posts[j] = posts[j], posts[i]
	}
	return posts, nil
}

// sendMattermostMessage posts a reply to a Mattermost channel.
func (s *AgentMonitorService) sendMattermostMessage(ctx context.Context, client *http.Client, serverURL, authHeader, mmChanID, text string) {
	payload, _ := json.Marshal(map[string]interface{}{
		"channel_id": mmChanID,
		"message":    text,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", serverURL+"/api/v4/posts", strings.NewReader(string(payload)))
	if err != nil {
		s.iLog.Error(fmt.Sprintf("sendMattermostMessage: build request error mm_channel=%s: %v", mmChanID, err))
		return
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("sendMattermostMessage: send error mm_channel=%s: %v", mmChanID, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.iLog.Error(fmt.Sprintf(
			"sendMattermostMessage: API error %d mm_channel=%s: %s",
			resp.StatusCode, mmChanID, string(body),
		))
		return
	}
	s.iLog.Info(fmt.Sprintf("sendMattermostMessage: delivered to mm_channel=%s", mmChanID))
}

// notifyMattermostChannels sends a notification to each configured Mattermost channel.
func (s *AgentMonitorService) notifyMattermostChannels(ctx context.Context, client *http.Client, serverURL, authHeader string, channelIDs []string, text string) {
	for _, cid := range channelIDs {
		s.sendMattermostMessage(ctx, client, serverURL, authHeader, cid, text)
	}
}
