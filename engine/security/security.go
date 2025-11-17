package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

// Permission represents a security permission
type Permission string

const (
	PermissionExecuteFunction    Permission = "execute:function"
	PermissionExecuteScript      Permission = "execute:script"
	PermissionAccessDatabase     Permission = "access:database"
	PermissionAccessSession      Permission = "access:session"
	PermissionExecuteWorkflow    Permission = "execute:workflow"
	PermissionSendMessage        Permission = "send:message"
	PermissionAccessExternalAPI  Permission = "access:external_api"
	PermissionAdministrator      Permission = "admin:all"
)

// Role represents a user role with associated permissions
type Role struct {
	Name        string
	Permissions []Permission
	Description string
}

// PredefinedRoles contains standard role definitions
var PredefinedRoles = map[string]*Role{
	"admin": {
		Name:        "Administrator",
		Permissions: []Permission{PermissionAdministrator},
		Description: "Full system access",
	},
	"developer": {
		Name: "Developer",
		Permissions: []Permission{
			PermissionExecuteFunction,
			PermissionExecuteScript,
			PermissionAccessDatabase,
			PermissionAccessSession,
		},
		Description: "Development and testing access",
	},
	"operator": {
		Name: "Operator",
		Permissions: []Permission{
			PermissionExecuteFunction,
			PermissionExecuteWorkflow,
			PermissionSendMessage,
		},
		Description: "Operational access",
	},
	"readonly": {
		Name:        "Read Only",
		Permissions: []Permission{},
		Description: "Read-only access",
	},
}

// SecurityContext holds security information for a request
type SecurityContext struct {
	UserID      string
	Username    string
	Roles       []string
	Permissions []Permission
	IPAddress   string
	SessionID   string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

// NewSecurityContext creates a new security context
func NewSecurityContext(userID, username string, roles []string) *SecurityContext {
	ctx := &SecurityContext{
		UserID:      userID,
		Username:    username,
		Roles:       roles,
		Permissions: []Permission{},
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	// Collect permissions from roles
	for _, roleName := range roles {
		if role, exists := PredefinedRoles[roleName]; exists {
			ctx.Permissions = append(ctx.Permissions, role.Permissions...)
		}
	}

	return ctx
}

// HasPermission checks if the context has a specific permission
func (sc *SecurityContext) HasPermission(permission Permission) bool {
	// Admin has all permissions
	for _, p := range sc.Permissions {
		if p == PermissionAdministrator {
			return true
		}
		if p == permission {
			return true
		}
	}
	return false
}

// HasRole checks if the context has a specific role
func (sc *SecurityContext) HasRole(roleName string) bool {
	for _, r := range sc.Roles {
		if r == roleName {
			return true
		}
	}
	return false
}

// IsExpired checks if the security context has expired
func (sc *SecurityContext) IsExpired() bool {
	return time.Now().After(sc.ExpiresAt)
}

// InputSanitizer provides input sanitization functions
type InputSanitizer struct{}

// SanitizeSQL prevents SQL injection by removing dangerous characters
func (is *InputSanitizer) SanitizeSQL(input string) string {
	// Remove common SQL injection patterns
	dangerous := []string{
		"--", ";--", "/*", "*/", "@@", "@",
		"char", "nchar", "varchar", "nvarchar",
		"alter", "begin", "cast", "create", "cursor",
		"declare", "delete", "drop", "end", "exec",
		"execute", "fetch", "insert", "kill", "select",
		"sys", "sysobjects", "syscolumns", "table", "update",
	}

	sanitized := strings.ToLower(input)
	for _, word := range dangerous {
		sanitized = strings.ReplaceAll(sanitized, word, "")
	}

	return sanitized
}

// ValidateSQLInput validates SQL input for dangerous patterns
func (is *InputSanitizer) ValidateSQLInput(input string) error {
	// Check for SQL injection patterns
	patterns := []string{
		`(?i)(union.*select)`,
		`(?i)(select.*from)`,
		`(?i)(insert.*into)`,
		`(?i)(delete.*from)`,
		`(?i)(drop.*table)`,
		`(?i)(update.*set)`,
		`(?i)(exec.*\()`,
		`(?i)(execute.*\()`,
		`--`,
		`;--`,
		`/\*`,
		`\*/`,
	}

	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, input)
		if err != nil {
			return fmt.Errorf("regex error: %w", err)
		}
		if matched {
			return fmt.Errorf("SQL injection pattern detected: %s", pattern)
		}
	}

	return nil
}

// SanitizeHTML removes HTML tags and dangerous characters
func (is *InputSanitizer) SanitizeHTML(input string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	sanitized := re.ReplaceAllString(input, "")

	// Escape special characters
	sanitized = strings.ReplaceAll(sanitized, "<", "&lt;")
	sanitized = strings.ReplaceAll(sanitized, ">", "&gt;")
	sanitized = strings.ReplaceAll(sanitized, "&", "&amp;")
	sanitized = strings.ReplaceAll(sanitized, "\"", "&quot;")
	sanitized = strings.ReplaceAll(sanitized, "'", "&#x27;")
	sanitized = strings.ReplaceAll(sanitized, "/", "&#x2F;")

	return sanitized
}

// SanitizeFilePath prevents path traversal attacks
func (is *InputSanitizer) SanitizeFilePath(input string) string {
	// Remove path traversal patterns
	sanitized := strings.ReplaceAll(input, "..", "")
	sanitized = strings.ReplaceAll(sanitized, "~", "")
	sanitized = strings.ReplaceAll(sanitized, "/./", "/")
	sanitized = strings.ReplaceAll(sanitized, "\\", "/")

	// Remove leading slashes
	sanitized = strings.TrimPrefix(sanitized, "/")

	return sanitized
}

// ValidateFilePath validates file path for dangerous patterns
func (is *InputSanitizer) ValidateFilePath(input string) error {
	dangerous := []string{"..", "~", "/etc", "/proc", "/sys", "c:\\", "c:/"}

	lowerInput := strings.ToLower(input)
	for _, pattern := range dangerous {
		if strings.Contains(lowerInput, pattern) {
			return fmt.Errorf("dangerous file path pattern detected: %s", pattern)
		}
	}

	return nil
}

// SanitizeCommand prevents command injection
func (is *InputSanitizer) SanitizeCommand(input string) string {
	// Remove command injection characters
	dangerous := []string{";", "|", "&", "$", "`", "\n", "\r", "(", ")", "<", ">"}

	sanitized := input
	for _, char := range dangerous {
		sanitized = strings.ReplaceAll(sanitized, char, "")
	}

	return sanitized
}

// ValidateCommand validates command input for dangerous patterns
func (is *InputSanitizer) ValidateCommand(input string) error {
	dangerous := []string{";", "|", "&", "$", "`", "&&", "||", ">", "<", "$(", "${"}

	for _, pattern := range dangerous {
		if strings.Contains(input, pattern) {
			return fmt.Errorf("command injection pattern detected: %s", pattern)
		}
	}

	return nil
}

// Global input sanitizer
var Sanitizer = &InputSanitizer{}

// EncryptionService provides encryption/decryption functionality
type EncryptionService struct {
	key []byte
}

// NewEncryptionService creates a new encryption service
func NewEncryptionService(passphrase string) *EncryptionService {
	// Derive a 32-byte key from passphrase using SHA-256
	hash := sha256.Sum256([]byte(passphrase))
	return &EncryptionService{
		key: hash[:],
	}
}

// Encrypt encrypts data using AES-GCM
func (es *EncryptionService) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(es.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts data using AES-GCM
func (es *EncryptionService) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(es.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// HashPassword creates a secure hash of a password
func (es *EncryptionService) HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// VerifyPassword verifies a password against its hash
func (es *EncryptionService) VerifyPassword(password, hash string) bool {
	return es.HashPassword(password) == hash
}

// ScriptSandbox provides security constraints for script execution
type ScriptSandbox struct {
	AllowedModules   []string
	MaxMemoryMB      int
	MaxCPUSeconds    int
	AllowNetwork     bool
	AllowFileAccess  bool
	AllowedFilePaths []string
	Timeout          time.Duration
}

// DefaultScriptSandbox returns a default secure sandbox configuration
func DefaultScriptSandbox() *ScriptSandbox {
	return &ScriptSandbox{
		AllowedModules:   []string{"json", "math", "string", "datetime"},
		MaxMemoryMB:      128,
		MaxCPUSeconds:    30,
		AllowNetwork:     false,
		AllowFileAccess:  false,
		AllowedFilePaths: []string{},
		Timeout:          30 * time.Second,
	}
}

// ValidateModuleImport checks if a module import is allowed
func (ss *ScriptSandbox) ValidateModuleImport(moduleName string) error {
	if len(ss.AllowedModules) == 0 {
		return nil // No restrictions
	}

	for _, allowed := range ss.AllowedModules {
		if moduleName == allowed {
			return nil
		}
	}

	return fmt.Errorf("module '%s' is not in the allowed list", moduleName)
}

// ValidateFilePath checks if file access is allowed
func (ss *ScriptSandbox) ValidateFilePath(path string) error {
	if !ss.AllowFileAccess {
		return fmt.Errorf("file access is disabled in sandbox")
	}

	// Validate path for security
	if err := Sanitizer.ValidateFilePath(path); err != nil {
		return err
	}

	if len(ss.AllowedFilePaths) == 0 {
		return nil // No restrictions
	}

	for _, allowed := range ss.AllowedFilePaths {
		if strings.HasPrefix(path, allowed) {
			return nil
		}
	}

	return fmt.Errorf("file path '%s' is not in the allowed list", path)
}

// RateLimiter limits the rate of operations
type RateLimiter struct {
	tokens     chan struct{}
	maxTokens  int
	refillRate time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())

	rl := &RateLimiter{
		tokens:     make(chan struct{}, maxTokens),
		maxTokens:  maxTokens,
		refillRate: refillRate,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Fill initial tokens
	for i := 0; i < maxTokens; i++ {
		rl.tokens <- struct{}{}
	}

	// Start refill goroutine
	go rl.refill()

	return rl
}

// refill periodically adds tokens back to the bucket
func (rl *RateLimiter) refill() {
	ticker := time.NewTicker(rl.refillRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Try to add a token (non-blocking)
			select {
			case rl.tokens <- struct{}{}:
				// Token added
			default:
				// Bucket is full, skip
			}

		case <-rl.ctx.Done():
			return
		}
	}
}

// Allow checks if an operation is allowed (blocks until token is available)
func (rl *RateLimiter) Allow() {
	<-rl.tokens
}

// TryAllow tries to get a token without blocking
func (rl *RateLimiter) TryAllow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() {
	rl.cancel()
}

// AuditLog records security-relevant events
type AuditLog struct {
	Timestamp   time.Time
	UserID      string
	Action      string
	Resource    string
	Result      string // Success, Failure, Denied
	Details     string
	IPAddress   string
	SessionID   string
}

// AuditLogger logs security events
type AuditLogger interface {
	LogAccess(userID, action, resource, result string)
	LogSecurityEvent(event *AuditLog)
}
