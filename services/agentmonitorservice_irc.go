package services

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// ─── IRC channel_type poller ──────────────────────────────────────────────────
//
// Connects to an IRC server via TCP (or TLS) and handles PRIVMSG events.
// On disconnect, waits 15s and reconnects automatically.
//
// Required config fields:
//   server_host — IRC server hostname
//   server_port — IRC server port (default: 6667, or 6697 for TLS)
//   nickname    — Bot nickname
//   channels    — comma-separated IRC channels to join (e.g. #general,#help)
//
// Optional config fields:
//   sasl_user — Server password / SASL username (sent as PASS command)
//   sasl_pass — SASL password (not used if sasl_user is empty)
//   use_tls   — "true" to enable TLS

// ─── ircPollLoop ──────────────────────────────────────────────────────────────

func (s *AgentMonitorService) ircPollLoop(ctx context.Context, entry *channelMonitorEntry) {
	host := entry.config["server_host"]
	portStr := entry.config["server_port"]
	nickname := entry.config["nickname"]
	channelsRaw := entry.config["channels"]
	saslPass := entry.config["sasl_pass"]
	useTLS := entry.config["use_tls"] == "true"

	if host == "" || nickname == "" || channelsRaw == "" {
		s.iLog.Error(fmt.Sprintf(
			"ircPollLoop: server_host/nickname/channels not configured for channel=%s — poller disabled",
			entry.channelName,
		))
		return
	}

	port := 6667
	if useTLS {
		port = 6697
	}
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	ircChannels := []string{}
	for _, ch := range strings.Split(channelsRaw, ",") {
		ch = strings.TrimSpace(ch)
		if ch != "" {
			ircChannels = append(ircChannels, ch)
		}
	}
	if len(ircChannels) == 0 {
		s.iLog.Error(fmt.Sprintf("ircPollLoop: no valid channels for channel=%s", entry.channelName))
		return
	}

	address := fmt.Sprintf("%s:%d", host, port)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		s.iLog.Info(fmt.Sprintf("ircPollLoop: connecting to %s for channel=%s", address, entry.channelName))

		err := s.ircConnect(ctx, entry, address, nickname, ircChannels, saslPass, useTLS)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			s.iLog.Error(fmt.Sprintf(
				"ircPollLoop: connection error channel=%s: %v — reconnecting in 15s",
				entry.channelName, err,
			))
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(15 * time.Second):
		}
	}
}

// ircConnect establishes a TCP/TLS connection to the IRC server, registers the bot,
// joins channels, and processes incoming PRIVMSG events until the connection drops or
// the context is cancelled.
func (s *AgentMonitorService) ircConnect(
	ctx context.Context,
	entry *channelMonitorEntry,
	address, nickname string,
	ircChannels []string,
	serverPass string,
	useTLS bool,
) error {
	var conn net.Conn
	var err error

	if useTLS {
		conn, err = tls.DialWithDialer(
			&net.Dialer{Timeout: 15 * time.Second},
			"tcp", address,
			&tls.Config{ServerName: strings.Split(address, ":")[0]},
		)
	} else {
		conn, err = net.DialTimeout("tcp", address, 15*time.Second)
	}
	if err != nil {
		return err
	}

	defer conn.Close()

	// writeLine sends a formatted IRC line with CRLF terminator.
	writeLine := func(line string) error {
		_, err := fmt.Fprintf(conn, "%s\r\n", line)
		return err
	}

	// Register connection.
	if serverPass != "" {
		writeLine(fmt.Sprintf("PASS %s", serverPass)) //nolint:errcheck
	}
	writeLine(fmt.Sprintf("NICK %s", nickname))                                                //nolint:errcheck
	writeLine(fmt.Sprintf("USER %s 0 * :IAC Agent (%s)", nickname, entry.agentName))          //nolint:errcheck

	scanner := bufio.NewScanner(conn)
	registered := false

	for {
		// Check for context cancellation before blocking on scanner.
		select {
		case <-ctx.Done():
			// Broadcast offline notification to all joined channels before quitting.
			stopMsg := entry.stopMessage
			for _, ch := range ircChannels {
				for _, chunk := range s.ircSplitMessage(stopMsg, 450) {
					writeLine(fmt.Sprintf("PRIVMSG %s :%s", ch, chunk)) //nolint:errcheck
				}
			}
			writeLine(fmt.Sprintf("QUIT :Agent [%s] shutdown", entry.agentName)) //nolint:errcheck
			return nil
		default:
		}

		// Set a read deadline so we can periodically check the context.
		conn.SetReadDeadline(time.Now().Add(30 * time.Second)) //nolint:errcheck

		if !scanner.Scan() {
			if netErr, ok := scanner.Err().(net.Error); ok && netErr.Timeout() {
				continue // deadline exceeded — loop back to check ctx
			}
			if scanner.Err() != nil {
				return scanner.Err()
			}
			return fmt.Errorf("connection closed by server")
		}

		line := scanner.Text()
		prefix, command, params := s.ircParseLine(line)

		switch command {
		case "PING":
			// Respond to keep-alive pings immediately.
			pongTarget := ""
			if len(params) > 0 {
				pongTarget = params[0]
			}
			writeLine(fmt.Sprintf("PONG :%s", pongTarget)) //nolint:errcheck

		case "001": // RPL_WELCOME — connection fully registered
			registered = true
			s.iLog.Info(fmt.Sprintf(
				"ircConnect: registered on %s as %s channel=%s",
				address, nickname, entry.channelName,
			))
			for _, ch := range ircChannels {
				writeLine(fmt.Sprintf("JOIN %s", ch)) //nolint:errcheck
			}
			// Send startup notification to each joined IRC channel.
			startMsg := entry.startMessage
			for _, ch := range ircChannels {
				for _, chunk := range s.ircSplitMessage(startMsg, 450) {
					writeLine(fmt.Sprintf("PRIVMSG %s :%s", ch, chunk)) //nolint:errcheck
				}
			}

		case "PRIVMSG":
			if !registered || len(params) < 2 {
				continue
			}

			target := params[0]
			text := params[1]

			// Extract sender nick from the :nick!user@host prefix.
			senderNick := prefix
			if idx := strings.Index(prefix, "!"); idx >= 0 {
				senderNick = prefix[:idx]
			}

			// Skip messages from ourselves.
			if strings.EqualFold(senderNick, nickname) {
				continue
			}

			// For private messages (target = our nick), reply directly to sender.
			replyTarget := target
			if strings.EqualFold(target, nickname) {
				replyTarget = senderNick
			}

			text = strings.TrimSpace(text)
			if text == "" {
				continue
			}

			s.iLog.Info(fmt.Sprintf(
				"ircConnect: message from sender=%s target=%s channel=%s text=%q",
				senderNick, target, entry.channelName, text,
			))

			go func(sNick, rTarget, msgText string) {
				response, err := s.channelSvc.HandleInbound(ctx, entry.channelID, sNick, sNick, msgText)
				if err != nil {
					s.iLog.Error(fmt.Sprintf(
						"ircConnect: HandleInbound error channel=%s sender=%s: %v",
						entry.channelName, sNick, err,
					))
					response = "Sorry, an error occurred processing your message. Please try again."
				}
				if response != "" {
					// IRC has a 512-byte message limit (including CRLF and command overhead).
					for _, chunk := range s.ircSplitMessage(response, 450) {
						writeLine(fmt.Sprintf("PRIVMSG %s :%s", rTarget, chunk)) //nolint:errcheck
					}
				}
			}(senderNick, replyTarget, text)

		case "ERROR":
			errMsg := ""
			if len(params) > 0 {
				errMsg = params[0]
			}
			return fmt.Errorf("IRC server error: %s", errMsg)
		}
	}
}

// ircParseLine parses a raw IRC protocol line into prefix, command, and params.
// IRC format: [:prefix] command [param ...] [:trailing]
func (s *AgentMonitorService) ircParseLine(line string) (prefix, command string, params []string) {
	rest := line

	if strings.HasPrefix(rest, ":") {
		parts := strings.SplitN(rest[1:], " ", 2)
		prefix = parts[0]
		if len(parts) > 1 {
			rest = parts[1]
		} else {
			rest = ""
		}
	}

	// Split trailing parameter (after " :").
	mainAndTrailing := strings.SplitN(rest, " :", 2)
	mainParts := strings.Fields(mainAndTrailing[0])
	if len(mainParts) == 0 {
		return
	}

	command = mainParts[0]
	params = mainParts[1:]
	if len(mainAndTrailing) > 1 {
		params = append(params, mainAndTrailing[1])
	}
	return
}

// ircSplitMessage splits a message into UTF-8-aware chunks of at most maxLen bytes.
func (s *AgentMonitorService) ircSplitMessage(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}
	var chunks []string
	for len(text) > maxLen {
		chunk := text[:maxLen]
		// Try to split at a space boundary.
		if idx := strings.LastIndex(chunk, " "); idx > maxLen/2 {
			chunk = text[:idx]
			text = text[idx+1:]
		} else {
			text = text[maxLen:]
		}
		chunks = append(chunks, chunk)
	}
	if text != "" {
		chunks = append(chunks, text)
	}
	return chunks
}
