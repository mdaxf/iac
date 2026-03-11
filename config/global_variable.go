// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"github.com/mdaxf/iac/framework/cache"
)

var (
	SessionCache        cache.Cache
	SessionCacheTimeout float64

	ObjectCache        cache.Cache
	ObjectCacheTimeout float64

	TestSessionCache        cache.Cache
	TestSessionCacheTimeout float64
)

var ApiKey string
var OpenAiKey string
var OpenAiModel string

var GlobalConfiguration *GlobalConfig

type GlobalConfig struct {
	Instance           string                   `json:"instance"`
	InstanceType       string                   `json:"type"`
	InstanceName       string                   `json:"name"`
	SingalRConfig      map[string]interface{}   `json:"singalrconfig"`
	LogConfig          map[string]interface{}   `json:"log"`
	DocumentConfig     map[string]interface{}   `json:"documentdb"`
	DocumentStorage    map[string]interface{}   `json:"documentstorage"`
	DatabaseConfig     map[string]interface{}   `json:"database"`
	AltDatabasesConfig []map[string]interface{} `json:"altdatabases"`
	CacheConfig        map[string]interface{}   `json:"cache"`
	TranslationConfig  map[string]interface{}   `json:"translation"`
	WebServerConfig    map[string]interface{}   `json:"webserver"`
	Transaction        map[string]interface{}   `json:"transaction"`
	AppServer          map[string]interface{}   `json:"appserver"`
	Services           []map[string]interface{} `json:"services"`
	JobsConfig         JobsConfiguration        `json:"jobs"`

	// Instance roles configuration - defines what features this instance handles
	// At least one role must be enabled
	Roles *InstanceRolesConfig `json:"roles,omitempty"`

	// Cluster configuration for distributed deployments
	Cluster *ClusterConfig `json:"cluster,omitempty"`

	// Notifications holds system-level defaults for all notification channels (SMTP, Telegram, Slack, etc.)
	Notifications NotificationsConfig `json:"notifications"`
}

// ClusterConfig holds configuration for distributed/clustered deployments
type ClusterConfig struct {
	// Enabled determines if cluster mode is active
	Enabled bool `json:"enabled"`

	// Mode specifies the cluster mode: "standalone", "distributed"
	Mode string `json:"mode"`

	// InstanceName is a unique name for this instance in the cluster
	InstanceName string `json:"instance_name"`

	// Environment specifies the deployment environment: "development", "test", "production"
	Environment string `json:"environment"`

	// Registry configuration for instance registration
	Registry *RegistryConfig `json:"registry,omitempty"`

	// ConfigStore configuration for centralized configuration
	ConfigStore *ConfigStoreConfig `json:"config_store,omitempty"`

	// Session configuration for distributed sessions
	Session *ClusterSessionConfig `json:"session,omitempty"`
}

// RegistryConfig holds configuration for instance registry
type RegistryConfig struct {
	// Collection name in MongoDB for instance registry
	Collection string `json:"collection"`

	// HeartbeatInterval in seconds
	HeartbeatInterval int `json:"heartbeat_interval"`

	// HeartbeatTTL in seconds - instance considered dead if no heartbeat within this time
	HeartbeatTTL int `json:"heartbeat_ttl"`

	// CleanupInterval in seconds - how often to clean up stale instances
	CleanupInterval int `json:"cleanup_interval"`

	// SignalRDiscovery enables SignalR-based instance discovery
	SignalRDiscovery bool `json:"signalr_discovery"`
}

// ConfigStoreConfig holds configuration for centralized configuration store
type ConfigStoreConfig struct {
	// Enabled determines if config store is active
	Enabled bool `json:"enabled"`

	// Collection name in MongoDB for configurations
	Collection string `json:"collection"`

	// WatchEnabled enables MongoDB change stream watching
	WatchEnabled bool `json:"watch_enabled"`

	// AutoReload automatically reloads configuration on changes
	AutoReload bool `json:"auto_reload"`
}

// ClusterSessionConfig holds configuration for distributed session management
type ClusterSessionConfig struct {
	// Distributed enables distributed session mode
	Distributed bool `json:"distributed"`

	// Backend specifies the primary session backend: "redis", "mongodb", "memory"
	Backend string `json:"backend"`

	// FallbackBackend specifies the fallback backend if primary fails
	FallbackBackend string `json:"fallback_backend"`

	// TTL is the session TTL in seconds
	TTL int `json:"ttl"`

	// LockTTL is the session lock TTL in seconds
	LockTTL int `json:"lock_ttl"`

	// AutoExtend enables automatic session extension
	AutoExtend bool `json:"auto_extend"`

	// Redis configuration for Redis backend
	Redis *RedisSessionConfig `json:"redis,omitempty"`

	// MongoDB configuration for MongoDB backend
	MongoDB *MongoSessionConfig `json:"mongodb,omitempty"`
}

// RedisSessionConfig holds Redis-specific session configuration
type RedisSessionConfig struct {
	// URL is the Redis connection URL
	URL string `json:"url"`

	// DB is the Redis database number
	DB int `json:"db"`

	// Password is the Redis password
	Password string `json:"password"`

	// PoolSize is the connection pool size
	PoolSize int `json:"pool_size"`

	// KeyPrefix is the prefix for all session keys
	KeyPrefix string `json:"key_prefix"`
}

// MongoSessionConfig holds MongoDB-specific session configuration
type MongoSessionConfig struct {
	// Collection name for sessions
	Collection string `json:"collection"`

	// Database name (uses global if empty)
	Database string `json:"database"`
}

// NotificationsConfig holds system-level defaults for all notification channels.
// Values here are used as fallbacks when agents do not supply credentials per-call
// and the agent definition does not carry its own notification_config overrides.
type NotificationsConfig struct {
	SMTP    SMTPConfig    `json:"smtp"`
	Telegram TelegramConfig `json:"telegram"`
	Slack   SlackConfig   `json:"slack"`
	Teams   WebhookChannel `json:"teams"`
	Discord WebhookChannel `json:"discord"`
}

// SMTPConfig holds SMTP server connection settings.
//
// Auth modes (auth_type):
//   "plain"  — standard SMTP AUTH PLAIN using username + password (default).
//              Works for: Gmail App Password, Outlook, Amazon SES, Mailgun.
//   "token"  — API token as the SMTP password (username may differ by provider):
//              • SendGrid  → username="apikey",      api_token=<SendGrid key>
//              • Postmark  → username=(leave empty), api_token=<Server token>   (token used as both user & pass)
//              • Brevo     → username=<your email>,  api_token=<Brevo SMTP key>
//              • Custom    → set username + api_token as needed
//   "none"   — no SMTP authentication (open relay / local MTA).
//
// Provider hint (provider) — optional, for documentation only, does not change behaviour:
//   "gmail", "sendgrid", "mailgun", "postmark", "ses", "outlook", "brevo", "custom"
//
// Credentials may also be supplied via environment variables to avoid storing secrets in JSON:
//   IAC_SMTP_HOST, IAC_SMTP_PORT, IAC_SMTP_USERNAME, IAC_SMTP_PASSWORD,
//   IAC_SMTP_API_TOKEN, IAC_SMTP_FROM
type SMTPConfig struct {
	// Host is the SMTP server hostname (e.g. "smtp.gmail.com", "smtp.sendgrid.net").
	Host string `json:"host"`

	// Port is the SMTP port. Common values: 587 (STARTTLS), 465 (TLS), 25 (plain/relay).
	Port int `json:"port"`

	// AuthType controls how authentication credentials are presented.
	// Values: "plain" (default), "token", "none".
	AuthType string `json:"auth_type"`

	// Username is the SMTP auth username.
	// For SendGrid set to "apikey"; for Postmark leave empty when using auth_type="token".
	Username string `json:"username"`

	// Password is used when auth_type="plain".
	Password string `json:"password"`

	// APIToken is used when auth_type="token".
	// Acts as the SMTP password; if Username is empty the token is also used as the username.
	APIToken string `json:"api_token"`

	// From is the default sender address, e.g. "IAC System <noreply@example.com>".
	From string `json:"from"`

	// UseTLS enables STARTTLS (recommended for port 587). Default: true when port is 587.
	UseTLS bool `json:"use_tls"`

	// Provider is an optional label for documentation ("gmail", "sendgrid", "postmark", etc.).
	Provider string `json:"provider"`
}

// TelegramConfig holds system-level Telegram Bot API settings.
type TelegramConfig struct {
	BotToken string `json:"bot_token"`
}

// SlackConfig holds system-level Slack Incoming Webhook settings.
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
}

// WebhookChannel holds a generic webhook URL for Teams, Discord, etc.
type WebhookChannel struct {
	WebhookURL string `json:"webhook_url"`
}

// JobsConfiguration holds the configuration for the background job system
type JobsConfiguration struct {
	Enabled                 bool `json:"enabled"`
	Workers                 int  `json:"workers"`
	PollInterval            int  `json:"poll_interval"`
	MaxRetries              int  `json:"max_retries"`
	SchedulerCheckInterval  int  `json:"scheduler_check_interval"`
	UseRedis                bool `json:"use_redis"`
	JobHistoryRetentionDays int  `json:"job_history_retention_days"`
	EnableMetrics           bool `json:"enable_metrics"`
}
