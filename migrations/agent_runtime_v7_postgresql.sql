-- Agent Runtime v7: Per-agent notification configuration
-- Adds notification_config JSONB column to agent_definitions so each agent
-- can store its own channel credentials (Telegram bot token / chat ID,
-- Slack webhook URL, SMTP from-address override, Teams/Discord webhook URL).
-- These values override the system-level defaults in configuration.json, and
-- are themselves overridden by credentials supplied as explicit tool parameters.
--
-- Run AFTER agent_runtime_v6_postgresql.sql.

ALTER TABLE agent_definitions
    ADD COLUMN IF NOT EXISTS notification_config JSONB NOT NULL DEFAULT '{}';

-- Supported notification_config keys (all values are strings):
--
-- Email (SMTP):
--   smtp_from       — sender address override, e.g. "Agent Reports <reports@example.com>"
--                     All other SMTP settings (host, port, TLS, credentials) are
--                     system-level only (configuration.json → notifications.smtp).
--
-- Telegram:
--   telegram_bot_token — Bot API token from @BotFather
--   telegram_chat_id   — target chat / group ID
--
-- Slack:
--   slack_webhook_url  — Incoming Webhook URL
--
-- Teams / Discord:
--   teams_webhook_url   — Teams Incoming Webhook connector URL
--   discord_webhook_url — Discord channel webhook URL
