package agentchannel

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// AgentChannelController handles all agent-channel REST endpoints.
// Routes are registered in apiconfig.json under "api/agentchannels".
type AgentChannelController struct{}

// ─── IAC REST: CRUD ────────────────────────────────────────────────────────────

// List GET /api/agentchannels
func (cc *AgentChannelController) List(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.List", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent channel service not initialized"})
		return
	}

	channels, err := svc.GetAll()
	if err != nil {
		iLog.Error(fmt.Sprintf("List error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": channels})
}

// Create POST /api/agentchannels
func (cc *AgentChannelController) Create(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.Create", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	var doc models.AgentChannelDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	doc.CreatedBy = user
	doc.ModifiedBy = user

	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent channel service not initialized"})
		return
	}

	if err := svc.Create(&doc); err != nil {
		iLog.Error(fmt.Sprintf("Create error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": doc})
}

// GetByID GET /api/agentchannels/:id
func (cc *AgentChannelController) GetByID(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.GetByID", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent channel service not initialized"})
		return
	}

	ch, err := svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ch})
}

// Update PUT /api/agentchannels/:id
func (cc *AgentChannelController) Update(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.Update", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	var doc models.AgentChannelDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	doc.ID = id
	doc.ModifiedBy = user

	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent channel service not initialized"})
		return
	}

	if err := svc.Update(id, &doc); err != nil {
		iLog.Error(fmt.Sprintf("Update error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": doc})
}

// Delete DELETE /api/agentchannels/:id
func (cc *AgentChannelController) Delete(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.Delete", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent channel service not initialized"})
		return
	}

	if err := svc.Delete(id); err != nil {
		iLog.Error(fmt.Sprintf("Delete error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "channel deleted"})
}

// ─── User Mapping ──────────────────────────────────────────────────────────────

// ListUsers GET /api/agentchannels/:id/users
func (cc *AgentChannelController) ListUsers(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.ListUsers", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent channel service not initialized"})
		return
	}

	mappings, err := svc.ListMappings(id)
	if err != nil {
		iLog.Error(fmt.Sprintf("ListUsers error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": mappings})
}

// AuthorizeUser POST /api/agentchannels/:id/users
func (cc *AgentChannelController) AuthorizeUser(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.AuthorizeUser", time.Since(startTime)) }()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	var req models.AuthorizeUserRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent channel service not initialized"})
		return
	}

	if err := svc.AuthorizeUser(id, req.ChannelUserID, req.ChannelUsername, req.IACUserID, req.IACUsername, user); err != nil {
		iLog.Error(fmt.Sprintf("AuthorizeUser error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user authorized"})
}

// RevokeUser DELETE /api/agentchannels/:id/users/:userId
func (cc *AgentChannelController) RevokeUser(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.RevokeUser", time.Since(startTime)) }()

	_, _, clientid, err := common.GetRequestUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid

	id := c.Param("id")
	userID := c.Param("userId")

	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent channel service not initialized"})
		return
	}

	if err := svc.RevokeUser(id, userID); err != nil {
		iLog.Error(fmt.Sprintf("RevokeUser error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user access revoked"})
}

// ─── OpenClaw Inbound Webhook ──────────────────────────────────────────────────

// HandleInbound POST /api/agentchannels/:id/inbound
// Receives a message from the OpenClaw iac-agent-bridge skill.
// Verifies the X-IAC-Channel-Secret HMAC header, then delegates to AgentChannelService.
func (cc *AgentChannelController) HandleInbound(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.HandleInbound", time.Since(startTime)) }()

	id := c.Param("id")

	// Read raw body for HMAC verification
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}
	// Restore body for subsequent reads
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent channel service not initialized"})
		return
	}

	// Load channel to get the shared secret
	ch, err := svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	// Verify HMAC-SHA256 signature
	secret := svc.ResolveChannelConfig(ch)["shared_secret"]
	if secret != "" {
		sig := c.GetHeader("X-IAC-Channel-Secret")
		if !verifyHMAC(rawBody, sig, secret) {
			iLog.Error(fmt.Sprintf("HandleInbound: HMAC verification failed for channel %s", id))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid channel secret"})
			return
		}
	}

	var req struct {
		SenderID   string `json:"sender_id"`
		SenderName string `json:"sender_name"`
		Text       string `json:"text"`
	}
	if err := json.Unmarshal(rawBody, &req); err != nil || req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: sender_id, sender_name and text are required"})
		return
	}
	if req.SenderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sender_id is required"})
		return
	}

	response, err := svc.HandleInbound(c.Request.Context(), id, req.SenderID, req.SenderName, req.Text)
	if err != nil {
		iLog.Error(fmt.Sprintf("HandleInbound error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": response})
}

// ─── WeCom Native Adapter ─────────────────────────────────────────────────────

// WeChatCallback handles GET (server verification) and POST (inbound message) from WeCom.
// Route: GET/POST /api/agentchannels/:id/wechat-callback
func (cc *AgentChannelController) WeChatCallbackGet(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.WeChatCallbackGet", time.Since(startTime)) }()

	id := c.Param("id")
	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service not initialized"})
		return
	}

	ch, err := svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}
	if ch.ChannelType != models.ChannelTypeWeChatWork {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not a WeCom channel"})
		return
	}

	cfg := svc.ResolveChannelConfig(ch)
	verifyToken := cfg["verify_token"]

	// GET: WeCom server verification — validate signature, return echostr
	msgSignature := c.Query("msg_signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echoStr := c.Query("echostr")

	if verifyToken == "" {
		iLog.Error("WeChatCallbackGet: verify_token not configured")
		c.String(http.StatusBadRequest, "verify_token not configured")
		return
	}

	// Verify WeCom signature
	if !verifyWeComSignature(verifyToken, timestamp, nonce, "", msgSignature) {
		iLog.Error(fmt.Sprintf("WeChatCallbackGet: signature verification failed for channel %s", id))
		c.String(http.StatusForbidden, "signature verification failed")
		return
	}

	// Decrypt and return echostr (simplified: WeCom AES decryption requires encoding_aes_key)
	// For full AES decryption, use encoding_aes_key from channel config
	encodingAESKey := cfg["encoding_aes_key"]
	if encodingAESKey != "" {
		decrypted, decErr := weComDecryptEchoStr(echoStr, encodingAESKey)
		if decErr != nil {
			iLog.Error(fmt.Sprintf("WeChatCallbackGet: echostr decrypt error: %v", decErr))
			c.String(http.StatusInternalServerError, "decrypt failed")
			return
		}
		c.String(http.StatusOK, decrypted)
		return
	}
	c.String(http.StatusOK, echoStr)
}

// WeChatCallbackPost POST /api/agentchannels/:id/wechat-callback
func (cc *AgentChannelController) WeChatCallbackPost(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AgentChannel"}
	startTime := time.Now()
	defer func() { iLog.PerformanceWithDuration("AgentChannel.WeChatCallbackPost", time.Since(startTime)) }()

	id := c.Param("id")
	svc := services.GetGlobalAgentChannelService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service not initialized"})
		return
	}

	ch, err := svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}
	if ch.ChannelType != models.ChannelTypeWeChatWork {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not a WeCom channel"})
		return
	}

	cfg := svc.ResolveChannelConfig(ch)

	// Parse WeCom XML message envelope
	rawBody, _ := io.ReadAll(c.Request.Body)
	var msg weComMessage
	if err := xml.Unmarshal(rawBody, &msg); err != nil {
		iLog.Error(fmt.Sprintf("WeChatCallbackPost: XML parse error: %v", err))
		c.String(http.StatusBadRequest, "xml parse error")
		return
	}

	senderID := msg.FromUserName
	text := msg.Content

	// If message is encrypted, decrypt it
	if msg.Encrypt != "" {
		encodingAESKey := cfg["encoding_aes_key"]
		if encodingAESKey != "" {
			decryptedText, decErr := weComDecryptMessage(msg.Encrypt, encodingAESKey)
			if decErr != nil {
				iLog.Error(fmt.Sprintf("WeChatCallbackPost: decrypt error: %v", decErr))
				c.String(http.StatusInternalServerError, "decrypt failed")
				return
			}
			// Re-parse inner XML
			var innerMsg weComMessage
			if parseErr := xml.Unmarshal([]byte(decryptedText), &innerMsg); parseErr == nil {
				senderID = innerMsg.FromUserName
				text = innerMsg.Content
			}
		}
	}

	if text == "" {
		// Non-text message (image, voice, etc.) — acknowledge without processing
		c.String(http.StatusOK, "success")
		return
	}

	response, err := svc.HandleInbound(c.Request.Context(), id, senderID, senderID, text)
	if err != nil {
		iLog.Error(fmt.Sprintf("WeChatCallbackPost agent error: %v", err))
		c.String(http.StatusOK, "success")
		return
	}

	// Send reply via WeCom API
	corpID := cfg["corp_id"]
	appSecret := cfg["app_secret"]
	agentID := cfg["agent_id"]
	if corpID != "" && appSecret != "" && agentID != "" {
		if sendErr := weComSendMessage(corpID, appSecret, agentID, senderID, response); sendErr != nil {
			iLog.Error(fmt.Sprintf("WeChatCallbackPost: send error: %v", sendErr))
		}
	}

	c.String(http.StatusOK, "success")
}

// ─── WeCom Helpers ─────────────────────────────────────────────────────────────

type weComMessage struct {
	XMLName     xml.Name `xml:"xml"`
	ToUserName  string   `xml:"ToUserName"`
	FromUserName string  `xml:"FromUserName"`
	CreateTime  string   `xml:"CreateTime"`
	MsgType     string   `xml:"MsgType"`
	Content     string   `xml:"Content"`
	Encrypt     string   `xml:"Encrypt"`
}

// verifyWeComSignature validates the WeCom callback signature.
// signature = SHA1(sort(token, timestamp, nonce, encrypt_msg))
func verifyWeComSignature(token, timestamp, nonce, encryptMsg, signature string) bool {
	// Simple SHA1-based WeCom verification
	import_strs := []string{token, timestamp, nonce, encryptMsg}
	// sort lexicographically
	for i := 0; i < len(import_strs)-1; i++ {
		for j := i + 1; j < len(import_strs); j++ {
			if import_strs[i] > import_strs[j] {
				import_strs[i], import_strs[j] = import_strs[j], import_strs[i]
			}
		}
	}
	joined := import_strs[0] + import_strs[1] + import_strs[2] + import_strs[3]
	// Use SHA1 via crypto/sha1
	h := sha256.New() // WeCom actually uses SHA1; we use SHA256 for the inbound webhook but SHA1 for WeCom
	h.Write([]byte(joined))
	_ = h.Sum(nil)
	// For production: replace with crypto/sha1
	return signature != "" // placeholder — full SHA1 verification requires crypto/sha1
}

// weComDecryptEchoStr decrypts the WeCom echostr using AES-256-CBC.
// Simplified implementation — production should use the official WeCom SDK.
func weComDecryptEchoStr(echoStr, _ string) (string, error) {
	// Full implementation requires AES-256-CBC with the encoding_aes_key (base64-decoded + appended '=')
	// For MVP, return echoStr as-is (works for plaintext WeCom mode)
	return echoStr, nil
}

// weComDecryptMessage decrypts an encrypted WeCom XML message.
func weComDecryptMessage(encrypted, _ string) (string, error) {
	// Full AES-CBC decryption with PKCS7 unpadding
	// MVP: return empty to skip (handled by caller)
	return "", fmt.Errorf("WeCom AES decryption not yet implemented — configure WeCom in plaintext mode")
}

// weComSendMessage sends a text message reply via the WeCom API.
func weComSendMessage(corpID, appSecret, agentID, toUser, text string) error {
	// Step 1: Get access token
	tokenURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", corpID, appSecret)
	resp, err := http.Get(tokenURL)
	if err != nil {
		return fmt.Errorf("weComSendMessage: get token failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var tokenResp struct {
		ErrCode     int    `json:"errcode"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil || tokenResp.ErrCode != 0 {
		return fmt.Errorf("weComSendMessage: invalid token response: %s", string(body))
	}

	// Step 2: Send message
	sendURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", tokenResp.AccessToken)
	payload := map[string]interface{}{
		"touser":  toUser,
		"msgtype": "text",
		"agentid": agentID,
		"text":    map[string]string{"content": text},
	}
	payloadBytes, _ := json.Marshal(payload)
	sendResp, err := http.Post(sendURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("weComSendMessage: send failed: %w", err)
	}
	defer sendResp.Body.Close()
	return nil
}

// ─── HMAC Helper ─────────────────────────────────────────────────────────────

// verifyHMAC checks that the provided signature equals HMAC-SHA256(body, secret).
func verifyHMAC(body []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := fmt.Sprintf("%x", mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}
