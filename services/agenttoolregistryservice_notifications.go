package services

// Notification/communication tools — allow agents to push results and reports to external channels:
// - send_email:    SMTP email
// - send_telegram: Telegram Bot API
// - send_webhook:  Generic HTTP webhook (Teams, Discord, custom)
// - send_slack:    Slack Incoming Webhook shorthand
//
// Credential resolution priority (highest wins):
//   1. Explicit per-call parameter (supplied by the LLM in tool args)
//   2. Per-agent notification_config (agent_definitions.notification_config JSONB)
//      injected into the request context by AgentRunnerService
//   3. System-level defaults from configuration.json → notifications section
//      accessed via config.GlobalConfiguration.Notifications

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/mdaxf/iac/config"
)

func (s *AgentToolRegistryService) registerNotificationTools() {
	s.registerEmailTool()
	s.registerTelegramTool()
	s.registerWebhookTool()
	s.registerSlackTool()
}

// ─── helpers to resolve credentials from the three-tier hierarchy ──────────────

// notifStr resolves a notification string by checking (in order):
//  1. args[argKey]   — explicit per-call parameter
//  2. agentCfg[cfgKey] — per-agent notification_config from context
//  3. sysFallback    — system-level default from configuration.json
//
// Values from all tiers are trimmed of leading/trailing whitespace to
// prevent configuration errors (e.g. trailing newlines in JSONB or JSON fields).
func notifStr(args map[string]interface{}, argKey string, agentCfg map[string]string, cfgKey string, sysFallback string) string {
	if v := strings.TrimSpace(stringArg(args, argKey)); v != "" {
		return v
	}
	if v, ok := agentCfg[cfgKey]; ok {
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	return strings.TrimSpace(sysFallback)
}

// sysNotif returns the system-level NotificationsConfig, or an empty one if GlobalConfiguration is nil.
func sysNotif() config.NotificationsConfig {
	if config.GlobalConfiguration != nil {
		return config.GlobalConfiguration.Notifications
	}
	return config.NotificationsConfig{}
}

// ─── Email ────────────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerEmailTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "send_email",
			Description: "Sends an email via the system-configured SMTP server. " +
				"SMTP server settings (host, port, credentials, TLS) are managed system-wide in configuration.json — " +
				"do NOT pass SMTP server parameters in the tool call. " +
				"Only provide the message content: 'to', 'subject', 'body' are required. " +
				"The sender address defaults to the system From address; supply 'from' only to override it for this email. " +
				"Supports plain text and HTML bodies.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"to":           {Type: "string", Description: "Comma-separated recipient addresses (required)."},
					"cc":           {Type: "string", Description: "Comma-separated CC addresses (optional)."},
					"subject":      {Type: "string", Description: "Email subject line (required)."},
					"body":         {Type: "string", Description: "Email body text (required)."},
					"from":         {Type: "string", Description: "Sender address override, e.g. 'Reports <reports@example.com>'. Omit to use the system default sender."},
					"content_type": {Type: "string", Description: "Body content type: 'plain' (default) or 'html'."},
				},
				Required: []string{"to", "subject", "body"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		// SMTP server settings come exclusively from system configuration.
		// Agents cannot override host, port, credentials, or TLS — only the sender address (from).
		sys := sysNotif()
		agentCfg := agentNotifConfigFromCtx(ctx)

		host := strings.TrimSpace(sys.SMTP.Host)
		username := strings.TrimSpace(sys.SMTP.Username)
		password := strings.TrimSpace(sys.SMTP.Password)
		apiToken := strings.TrimSpace(sys.SMTP.APIToken)
		authType := strings.TrimSpace(sys.SMTP.AuthType)
		if authType == "" {
			authType = "plain"
		}
		port := sys.SMTP.Port
		if port == 0 {
			port = 587
		}

		// Only the sender address (from) can be overridden at agent or call level.
		from := notifStr(args, "from", agentCfg, "smtp_from", sys.SMTP.From)

		to := stringArg(args, "to")
		subject := stringArg(args, "subject")
		body := stringArg(args, "body")

		if host == "" {
			return "", fmt.Errorf("SMTP host is not configured: set notifications.smtp.host in configuration.json")
		}
		if from == "" {
			return "", fmt.Errorf("sender address is not configured: set notifications.smtp.from in configuration.json or supply 'from' in the tool call")
		}
		if to == "" || subject == "" || body == "" {
			return "", fmt.Errorf("to, subject, and body are required")
		}

		contentType := "text/plain"
		if ct := stringArg(args, "content_type"); ct == "html" {
			contentType = "text/html"
		}

		toList := strings.Split(to, ",")
		ccList := splitCSV(stringArg(args, "cc"))
		allRecipients := append(toList, ccList...)
		for i, r := range allRecipients {
			allRecipients[i] = strings.TrimSpace(r)
		}

		ccHeader := ""
		if len(ccList) > 0 {
			ccHeader = "Cc: " + strings.Join(ccList, ", ") + "\r\n"
		}
		msg := fmt.Sprintf(
			"From: %s\r\nTo: %s\r\n%sSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s; charset=UTF-8\r\nDate: %s\r\n\r\n%s",
			from, strings.Join(toList, ", "), ccHeader, subject, contentType,
			time.Now().UTC().Format(time.RFC1123Z), body,
		)

		addr := fmt.Sprintf("%s:%d", host, port)

		// TLS comes from system config only. Default to true for port 587 (STARTTLS standard).
		useTLS := sys.SMTP.UseTLS || port == 587

		// buildSMTPAuth returns the smtp.Auth to use based on auth_type.
		// "plain"  → PlainAuth with username + password
		// "token"  → PlainAuth with username + apiToken (if username empty, token is used as both)
		// "none"   → nil (no auth)
		buildSMTPAuth := func() smtp.Auth {
			switch authType {
			case "token":
				user := username
				if user == "" {
					user = apiToken // Postmark-style: token as both username and password
				}
				if apiToken == "" {
					return nil
				}
				return smtp.PlainAuth("", user, apiToken, host)
			case "none":
				return nil
			default: // "plain"
				if username == "" {
					return nil
				}
				return smtp.PlainAuth("", username, password, host)
			}
		}

		var sendErr error
		if useTLS {
			c, err := smtp.Dial(addr)
			if err != nil {
				return "", fmt.Errorf("smtp dial error: %w", err)
			}
			defer c.Close()
			tlsCfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
			if err := c.StartTLS(tlsCfg); err != nil {
				return "", fmt.Errorf("starttls error: %w", err)
			}
			if auth := buildSMTPAuth(); auth != nil {
				if err := c.Auth(auth); err != nil {
					return "", fmt.Errorf("smtp auth error: %w", err)
				}
			}
			if err := c.Mail(from); err != nil {
				return "", fmt.Errorf("smtp MAIL FROM error: %w", err)
			}
			for _, r := range allRecipients {
				if err := c.Rcpt(r); err != nil {
					return "", fmt.Errorf("smtp RCPT TO %q error: %w", r, err)
				}
			}
			w, err := c.Data()
			if err != nil {
				return "", fmt.Errorf("smtp DATA error: %w", err)
			}
			if _, err = w.Write([]byte(msg)); err != nil {
				return "", fmt.Errorf("smtp write error: %w", err)
			}
			sendErr = w.Close()
		} else {
			sendErr = smtp.SendMail(addr, buildSMTPAuth(), from, allRecipients, []byte(msg))
		}

		if sendErr != nil {
			return "", fmt.Errorf("failed to send email: %w", sendErr)
		}

		out, _ := json.Marshal(map[string]interface{}{
			"sent":      true,
			"to":        toList,
			"subject":   subject,
			"provider":  sys.SMTP.Provider,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return string(out), nil
	})
}

// ─── Telegram ────────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerTelegramTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "send_telegram",
			Description: "Sends a message to a Telegram chat via the Telegram Bot API. " +
				"Credentials are resolved in order: " +
				"(1) explicit tool parameters, " +
				"(2) agent notification_config (keys: telegram_bot_token, telegram_chat_id), " +
				"(3) system-level default bot_token from configuration.json notifications.telegram. " +
				"If the agent has a telegram_chat_id in its notification_config you may omit chat_id. " +
				"Supports Markdown formatting in the message text.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"bot_token": {
						Type:        "string",
						Description: "Telegram Bot API token. Falls back to agent notification_config telegram_bot_token, then system config notifications.telegram.bot_token.",
					},
					"chat_id": {
						Type:        "string",
						Description: "Target chat ID. Falls back to agent notification_config telegram_chat_id. Required if not set in agent config.",
					},
					"text": {
						Type:        "string",
						Description: "Message text (required). Supports Markdown: *bold*, _italic_, `code`, ```code block```, [link](url)",
					},
					"parse_mode": {
						Type:        "string",
						Description: "Text formatting: 'Markdown' (default) or 'HTML' or '' for plain text.",
					},
					"disable_notification": {
						Type:        "boolean",
						Description: "Send silently without notification sound (default: false).",
					},
				},
				Required: []string{"text"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		sys := sysNotif()
		agentCfg := agentNotifConfigFromCtx(ctx)

		botToken := notifStr(args, "bot_token", agentCfg, "telegram_bot_token", sys.Telegram.BotToken)
		chatID := notifStr(args, "chat_id", agentCfg, "telegram_chat_id", "")
		text := stringArg(args, "text")

		if botToken == "" {
			return "", fmt.Errorf("Telegram bot_token is required: supply bot_token parameter, set agent notification_config telegram_bot_token, or set notifications.telegram.bot_token in configuration.json")
		}
		if chatID == "" {
			return "", fmt.Errorf("Telegram chat_id is required: supply chat_id parameter or set agent notification_config telegram_chat_id")
		}
		if text == "" {
			return "", fmt.Errorf("text is required")
		}

		parseMode := stringArg(args, "parse_mode")
		if parseMode == "" {
			parseMode = "Markdown"
		}

		payload := map[string]interface{}{
			"chat_id":              chatID,
			"text":                 text,
			"parse_mode":           parseMode,
			"disable_notification": boolArg(args, "disable_notification"),
		}

		apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
		resp, err := s.postJSON(ctx, apiURL, nil, payload)
		if err != nil {
			return "", fmt.Errorf("telegram API error: %w", err)
		}

		var tgResp struct {
			OK          bool        `json:"ok"`
			Result      interface{} `json:"result"`
			Description string      `json:"description"`
			ErrorCode   int         `json:"error_code"`
		}
		if jsonErr := json.Unmarshal([]byte(resp), &tgResp); jsonErr == nil && !tgResp.OK {
			// Include chat_id in error so it's easy to spot misconfiguration.
			// Common causes of "chat not found": wrong ID, bot not added to group,
			// user hasn't sent /start to the bot (private chats), or missing -100 prefix for supergroups.
			return "", fmt.Errorf("telegram API error %d (chat_id=%q): %s", tgResp.ErrorCode, chatID, tgResp.Description)
		}

		out, _ := json.Marshal(map[string]interface{}{
			"sent":      true,
			"chat_id":   chatID,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return string(out), nil
	})
}

// ─── Generic Webhook ──────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerWebhookTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "send_webhook",
			Description: "Sends a JSON payload to any HTTP webhook URL. " +
				"Works with Microsoft Teams (use agent notification_config teams_webhook_url or system config notifications.teams.webhook_url), " +
				"Discord (use agent notification_config discord_webhook_url or system config notifications.discord.webhook_url), " +
				"and any service that accepts incoming webhooks. " +
				"The 'url' parameter can be omitted if the channel URL is configured at agent or system level — " +
				"specify 'channel' as 'teams' or 'discord' to auto-resolve.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"url": {
						Type:        "string",
						Description: "Webhook URL to POST to. Can be omitted if 'channel' is 'teams' or 'discord' and the URL is configured at agent or system level.",
					},
					"channel": {
						Type:        "string",
						Description: "Named channel shorthand to auto-resolve URL: 'teams' or 'discord'. Ignored if 'url' is supplied.",
					},
					"payload_json": {
						Type: "string",
						Description: `JSON payload to send (required). For Teams: {"text":"message"} or adaptive card format. ` +
							`For Discord: {"content":"message","username":"IAC Agent"}. ` +
							`For custom: any valid JSON object.`,
					},
					"headers_json": {
						Type:        "string",
						Description: `Optional custom HTTP headers as JSON object. Example: {"Authorization":"Bearer token123"}`,
					},
					"method": {
						Type:        "string",
						Description: "HTTP method: POST (default) or PUT.",
					},
				},
				Required: []string{"payload_json"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		sys := sysNotif()
		agentCfg := agentNotifConfigFromCtx(ctx)

		webhookURL := stringArg(args, "url")
		if webhookURL == "" {
			// Auto-resolve from named channel
			channel := strings.ToLower(stringArg(args, "channel"))
			switch channel {
			case "teams":
				webhookURL = notifStr(args, "", agentCfg, "teams_webhook_url", sys.Teams.WebhookURL)
			case "discord":
				webhookURL = notifStr(args, "", agentCfg, "discord_webhook_url", sys.Discord.WebhookURL)
			}
		}
		if webhookURL == "" {
			return "", fmt.Errorf("webhook URL is required: supply 'url' parameter, set agent notification_config teams_webhook_url/discord_webhook_url, or set channel to 'teams'/'discord' and configure the URL in configuration.json")
		}

		payloadJSON := stringArg(args, "payload_json")
		if payloadJSON == "" {
			return "", fmt.Errorf("payload_json is required")
		}

		var payload interface{}
		if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
			return "", fmt.Errorf("invalid payload_json: %w", err)
		}

		var headers map[string]string
		if headerJSON := stringArg(args, "headers_json"); headerJSON != "" {
			if err := json.Unmarshal([]byte(headerJSON), &headers); err != nil {
				return "", fmt.Errorf("invalid headers_json: %w", err)
			}
		}

		resp, err := s.postJSON(ctx, webhookURL, headers, payload)
		if err != nil {
			return "", fmt.Errorf("webhook error: %w", err)
		}

		out, _ := json.Marshal(map[string]interface{}{
			"sent":      true,
			"url":       webhookURL,
			"response":  resp,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return string(out), nil
	})
}

// ─── Slack shorthand ──────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerSlackTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "send_slack",
			Description: "Sends a message to a Slack channel via an Incoming Webhook. " +
				"The webhook URL is resolved in order: " +
				"(1) explicit webhook_url parameter, " +
				"(2) agent notification_config slack_webhook_url, " +
				"(3) system config notifications.slack.webhook_url in configuration.json. " +
				"If the URL is pre-configured you only need to supply 'text'.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"webhook_url": {
						Type:        "string",
						Description: "Slack Incoming Webhook URL. Falls back to agent notification_config slack_webhook_url, then system config notifications.slack.webhook_url.",
					},
					"text": {
						Type:        "string",
						Description: "Message text (required). Supports Slack mrkdwn: *bold*, _italic_, `code`, >quote",
					},
					"username": {
						Type:        "string",
						Description: "Override the bot display name (optional, e.g. 'IAC Agent').",
					},
					"icon_emoji": {
						Type:        "string",
						Description: "Override the bot icon emoji (optional, e.g. ':robot_face:').",
					},
					"channel": {
						Type:        "string",
						Description: "Override the target channel (optional, e.g. '#alerts' or '@username').",
					},
					"attachments_json": {
						Type:        "string",
						Description: `Optional Slack attachments JSON array for rich formatting. Example: [{"color":"#36a64f","title":"Report Ready","text":"Done.","footer":"IAC Agent"}]`,
					},
				},
				Required: []string{"text"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		sys := sysNotif()
		agentCfg := agentNotifConfigFromCtx(ctx)

		webhookURL := notifStr(args, "webhook_url", agentCfg, "slack_webhook_url", sys.Slack.WebhookURL)
		text := stringArg(args, "text")

		if webhookURL == "" {
			return "", fmt.Errorf("Slack webhook URL is required: supply webhook_url parameter, set agent notification_config slack_webhook_url, or set notifications.slack.webhook_url in configuration.json")
		}
		if text == "" {
			return "", fmt.Errorf("text is required")
		}

		payload := map[string]interface{}{
			"text": text,
		}
		if v := stringArg(args, "username"); v != "" {
			payload["username"] = v
		}
		if v := stringArg(args, "icon_emoji"); v != "" {
			payload["icon_emoji"] = v
		}
		if v := stringArg(args, "channel"); v != "" {
			payload["channel"] = v
		}
		if attachJSON := stringArg(args, "attachments_json"); attachJSON != "" {
			var attachments []interface{}
			if json.Unmarshal([]byte(attachJSON), &attachments) == nil {
				payload["attachments"] = attachments
			}
		}

		_, err := s.postJSON(ctx, webhookURL, nil, payload)
		if err != nil {
			return "", fmt.Errorf("slack webhook error: %w", err)
		}

		out, _ := json.Marshal(map[string]interface{}{
			"sent":      true,
			"channel":   stringArg(args, "channel"),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return string(out), nil
	})
}

// ─── shared HTTP helper ───────────────────────────────────────────────────────

// postJSON sends a JSON POST request and returns the response body as a string.
func (s *AgentToolRegistryService) postJSON(ctx context.Context, url string, headers map[string]string, payload interface{}) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		preview := strings.TrimSpace(string(respBody))
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, preview)
	}

	return string(respBody), nil
}
