package agentchannel

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/services"
)

// ─── LINE Webhook ─────────────────────────────────────────────────────────────
//
// Configure LINE Developers console to POST events to:
//   POST /api/agentchannels/:id/line-webhook
//
// Required channel config: channel_secret, channel_access_token

// LINEWebhookPost handles inbound LINE Messaging API webhook events.
// Signature verified with X-Line-Signature (HMAC-SHA256 of body with channel_secret).
func (cc *AgentChannelController) LINEWebhookPost(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.LINEWebhookPost", time.Since(startTime)) }()

	id := c.Param("id")
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	ch, err := svc.GetByID(id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	cfg := svc.ResolveChannelConfig(ch)
	channelSecret := cfg["channel_secret"]
	accessToken := cfg["channel_access_token"]

	// Verify LINE signature: base64(HMAC-SHA256(channel_secret, body)).
	if channelSecret != "" {
		sig := c.GetHeader("X-Line-Signature")
		if !verifyLINESignature(rawBody, sig, channelSecret) {
			iLog.Error(fmt.Sprintf("LINEWebhookPost: signature verification failed for channel %s", id))
			c.Status(http.StatusUnauthorized)
			return
		}
	}

	// Parse LINE webhook payload.
	var payload struct {
		Events []struct {
			Type string `json:"type"`
			Source struct {
				Type    string `json:"type"`
				UserID  string `json:"userId"`
				GroupID string `json:"groupId"`
				RoomID  string `json:"roomId"`
			} `json:"source"`
			Message *struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"message"`
		} `json:"events"`
	}
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Acknowledge immediately — LINE requires a 200 response within 1s.
	c.Status(http.StatusOK)

	// Process each text message event asynchronously.
	for _, event := range payload.Events {
		if event.Type != "message" || event.Message == nil || event.Message.Type != "text" {
			continue
		}
		userID := event.Source.UserID
		text := strings.TrimSpace(event.Message.Text)
		if userID == "" || text == "" {
			continue
		}

		go func(uid, msgText string) {
			response, err := svc.HandleInbound(context.Background(), id, uid, uid, msgText)
			if err != nil {
				iLog.Error(fmt.Sprintf("LINEWebhookPost: HandleInbound error channel=%s sender=%s: %v", id, uid, err))
				response = "Sorry, an error occurred. Please try again."
			}
			if response != "" {
				sendLINEPushMessage(accessToken, uid, response, &iLog)
			}
		}(userID, text)
	}
}

// verifyLINESignature validates the X-Line-Signature header.
func verifyLINESignature(body []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// sendLINEPushMessage delivers a text message to a LINE user via the push API.
func sendLINEPushMessage(accessToken, userID, text string, iLog *logger.Log) {
	payload, _ := json.Marshal(map[string]interface{}{
		"to": userID,
		"messages": []map[string]interface{}{
			{"type": "text", "text": text},
		},
	})
	req, err := http.NewRequest("POST", "https://api.line.me/v2/bot/message/push", bytes.NewReader(payload))
	if err != nil {
		iLog.Error(fmt.Sprintf("sendLINEPushMessage: build request error user=%s: %v", userID, err))
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		iLog.Error(fmt.Sprintf("sendLINEPushMessage: send error user=%s: %v", userID, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		iLog.Error(fmt.Sprintf("sendLINEPushMessage: LINE API error %d user=%s: %s", resp.StatusCode, userID, string(body)))
		return
	}
	iLog.Info(fmt.Sprintf("sendLINEPushMessage: delivered to user=%s", userID))
}

// ─── WhatsApp Webhook ─────────────────────────────────────────────────────────
//
// Configure Meta webhook to:
//   GET  /api/agentchannels/:id/whatsapp-webhook  (verification challenge)
//   POST /api/agentchannels/:id/whatsapp-webhook  (inbound messages)
//
// Required channel config: phone_number_id, access_token, webhook_verify_token

// WhatsAppWebhookGet handles Meta webhook verification (hub.challenge response).
func (cc *AgentChannelController) WhatsAppWebhookGet(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.WhatsAppWebhookGet", time.Since(startTime)) }()

	id := c.Param("id")
	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	ch, err := svc.GetByID(id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	cfg := svc.ResolveChannelConfig(ch)
	verifyToken := cfg["webhook_verify_token"]

	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token == verifyToken {
		iLog.Info(fmt.Sprintf("WhatsAppWebhookGet: verification successful for channel=%s", id))
		c.String(http.StatusOK, challenge)
		return
	}

	iLog.Error(fmt.Sprintf("WhatsAppWebhookGet: verification failed for channel=%s (token mismatch)", id))
	c.Status(http.StatusForbidden)
}

// WhatsAppWebhookPost handles inbound WhatsApp Cloud API webhook events.
func (cc *AgentChannelController) WhatsAppWebhookPost(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.WhatsAppWebhookPost", time.Since(startTime)) }()

	id := c.Param("id")
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	ch, err := svc.GetByID(id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	cfg := svc.ResolveChannelConfig(ch)
	accessToken := cfg["access_token"]
	phoneNumberID := cfg["phone_number_id"]

	// Parse Meta Cloud API payload.
	var payload struct {
		Entry []struct {
			Changes []struct {
				Value struct {
					Messages []struct {
						From string `json:"from"`
						Type string `json:"type"`
						Text *struct {
							Body string `json:"body"`
						} `json:"text"`
					} `json:"messages"`
					Contacts []struct {
						WaID    string `json:"wa_id"`
						Profile struct {
							Name string `json:"name"`
						} `json:"profile"`
					} `json:"contacts"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Acknowledge immediately — Meta requires HTTP 200 within a few seconds.
	c.Status(http.StatusOK)

	// Build sender name map from contacts array.
	nameMap := map[string]string{}
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, contact := range change.Value.Contacts {
				nameMap[contact.WaID] = contact.Profile.Name
			}
		}
	}

	for _, waEntry := range payload.Entry {
		for _, change := range waEntry.Changes {
			for _, msg := range change.Value.Messages {
				if msg.Type != "text" || msg.Text == nil {
					continue
				}
				senderID := msg.From
				senderName := nameMap[senderID]
				if senderName == "" {
					senderName = senderID
				}
				text := strings.TrimSpace(msg.Text.Body)
				if text == "" {
					continue
				}

				go func(sID, sName, msgText string) {
					response, err := svc.HandleInbound(context.Background(), id, sID, sName, msgText)
					if err != nil {
						iLog.Error(fmt.Sprintf("WhatsAppWebhookPost: HandleInbound error channel=%s sender=%s: %v", id, sID, err))
						response = "Sorry, an error occurred. Please try again."
					}
					if response != "" {
						sendWhatsAppMessage(accessToken, phoneNumberID, sID, response, &iLog)
					}
				}(senderID, senderName, text)
			}
		}
	}
}

// sendWhatsAppMessage sends a text message via Meta Cloud API.
func sendWhatsAppMessage(accessToken, phoneNumberID, to, text string, iLog *logger.Log) {
	apiURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", phoneNumberID)
	payload, _ := json.Marshal(map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text":              map[string]string{"body": text},
	})

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payload))
	if err != nil {
		iLog.Error(fmt.Sprintf("sendWhatsAppMessage: build request error to=%s: %v", to, err))
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		iLog.Error(fmt.Sprintf("sendWhatsAppMessage: send error to=%s: %v", to, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		iLog.Error(fmt.Sprintf("sendWhatsAppMessage: Meta API error %d to=%s: %s", resp.StatusCode, to, string(body)))
		return
	}
	iLog.Info(fmt.Sprintf("sendWhatsAppMessage: delivered to=%s", to))
}

// ─── Microsoft Teams Webhook ──────────────────────────────────────────────────
//
// Configure Bot Framework messaging endpoint to:
//   POST /api/agentchannels/:id/teams-webhook
//
// Required channel config: app_id, app_password, tenant_id

// TeamsWebhookPost handles inbound Microsoft Teams Bot Framework Activity messages.
func (cc *AgentChannelController) TeamsWebhookPost(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.TeamsWebhookPost", time.Since(startTime)) }()

	id := c.Param("id")
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	ch, err := svc.GetByID(id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	cfg := svc.ResolveChannelConfig(ch)
	appID := cfg["app_id"]
	appPassword := cfg["app_password"]
	tenantID := cfg["tenant_id"]

	// Parse Bot Framework Activity object.
	var activity struct {
		Type         string `json:"type"`
		ID           string `json:"id"`
		From         struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"from"`
		Conversation struct {
			ID       string `json:"id"`
			TenantID string `json:"tenantId"`
		} `json:"conversation"`
		Text       string `json:"text"`
		ServiceURL string `json:"serviceUrl"`
		ChannelID  string `json:"channelId"`
	}
	if err := json.Unmarshal(rawBody, &activity); err != nil || activity.Type != "message" {
		// Bot Framework always expects 200 OK, even for non-message activities.
		c.Status(http.StatusOK)
		return
	}

	senderID := activity.From.ID
	senderName := activity.From.Name
	text := strings.TrimSpace(activity.Text)
	serviceURL := strings.TrimRight(activity.ServiceURL, "/")
	convID := activity.Conversation.ID

	if senderID == "" || text == "" {
		c.Status(http.StatusOK)
		return
	}

	// Acknowledge immediately — Bot Framework requires a quick 200 response.
	c.Status(http.StatusOK)

	go func(sID, sName, msgText, svcURL, convId string) {
		response, err := svc.HandleInbound(context.Background(), id, sID, sName, msgText)
		if err != nil {
			iLog.Error(fmt.Sprintf("TeamsWebhookPost: HandleInbound error channel=%s sender=%s: %v", id, sID, err))
			response = "Sorry, an error occurred. Please try again."
		}
		if response != "" && svcURL != "" && convId != "" {
			sendTeamsMessage(appID, appPassword, tenantID, svcURL, convId, response, &iLog)
		}
	}(senderID, senderName, text, serviceURL, convID)
}

// sendTeamsMessage obtains a Bot Framework token and sends a reply activity.
func sendTeamsMessage(appID, appPassword, tenantID, serviceURL, conversationID, text string, iLog *logger.Log) {
	// Step 1: Obtain Bot Framework OAuth2 token.
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)
	tokenBody := fmt.Sprintf(
		"grant_type=client_credentials&client_id=%s&client_secret=%s&scope=%s",
		url.QueryEscape(appID),
		url.QueryEscape(appPassword),
		url.QueryEscape("https://api.botframework.com/.default"),
	)

	tokenResp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(tokenBody))
	if err != nil {
		iLog.Error(fmt.Sprintf("sendTeamsMessage: token request error conv=%s: %v", conversationID, err))
		return
	}
	defer tokenResp.Body.Close()

	var tokenResult struct {
		AccessToken string `json:"access_token"`
	}
	tokenBodyBytes, _ := io.ReadAll(tokenResp.Body)
	if json.Unmarshal(tokenBodyBytes, &tokenResult) != nil || tokenResult.AccessToken == "" {
		iLog.Error(fmt.Sprintf("sendTeamsMessage: failed to obtain Bot Framework token for conv=%s", conversationID))
		return
	}

	// Step 2: Send message activity to the conversation.
	apiURL := fmt.Sprintf("%s/v3/conversations/%s/activities", serviceURL, url.PathEscape(conversationID))
	payload, _ := json.Marshal(map[string]interface{}{
		"type": "message",
		"text": text,
	})

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payload))
	if err != nil {
		iLog.Error(fmt.Sprintf("sendTeamsMessage: build request error conv=%s: %v", conversationID, err))
		return
	}
	req.Header.Set("Authorization", "Bearer "+tokenResult.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		iLog.Error(fmt.Sprintf("sendTeamsMessage: send error conv=%s: %v", conversationID, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		iLog.Error(fmt.Sprintf("sendTeamsMessage: Bot Framework API error %d conv=%s: %s", resp.StatusCode, conversationID, string(body)))
		return
	}
	iLog.Info(fmt.Sprintf("sendTeamsMessage: delivered to conv=%s", conversationID))
}

// ─── Shared webhook helper ────────────────────────────────────────────────────

// verifyHMAC is shared between this file and agentchannel_controller.go via package scope.
// It is already declared in agentchannel_controller.go; no re-declaration needed here.
