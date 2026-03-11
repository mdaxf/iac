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

// ─── Signal channel_type poller ───────────────────────────────────────────────
//
// Polls signal-cli REST API for new messages every 2s.
// Requires signal-cli-rest-api daemon: https://github.com/bbernhard/signal-cli-rest-api
//
// Required config fields:
//   phone_number    — E.164 number registered with signal-cli (e.g. +15551234567)
//   signal_cli_host — hostname/IP of the signal-cli-rest-api server
//   signal_cli_port — port (default 8080)

// ─── Signal API response structures ──────────────────────────────────────────

type signalEnvelope struct {
	Envelope struct {
		Source       string `json:"source"`
		SourceNumber string `json:"sourceNumber"`
		SourceName   string `json:"sourceName"`
		DataMessage  *struct {
			Timestamp int64  `json:"timestamp"`
			Message   string `json:"message"`
		} `json:"dataMessage"`
	} `json:"envelope"`
	Account string `json:"account"`
}

// ─── signalPollLoop ───────────────────────────────────────────────────────────

func (s *AgentMonitorService) signalPollLoop(ctx context.Context, entry *channelMonitorEntry) {
	phoneNumber := entry.config["phone_number"]
	cliHost := entry.config["signal_cli_host"]
	cliPort := entry.config["signal_cli_port"]

	if phoneNumber == "" || cliHost == "" {
		s.iLog.Error(fmt.Sprintf(
			"signalPollLoop: phone_number/signal_cli_host not configured for channel=%s — poller disabled",
			entry.channelName,
		))
		return
	}
	if cliPort == "" {
		cliPort = "8080"
	}

	baseURL := fmt.Sprintf("http://%s:%s", cliHost, cliPort)
	httpClient := &http.Client{Timeout: 10 * time.Second}

	s.iLog.Info(fmt.Sprintf(
		"signalPollLoop: started for channel=%s agent=%s number=%s endpoint=%s",
		entry.channelName, entry.agentName, phoneNumber, baseURL,
	))

	// activeSenders tracks phone numbers that have sent messages; used to broadcast
	// the offline notification on shutdown.
	activeSenders := make(map[string]bool)

	defer func() {
		if len(activeSenders) > 0 {
			stopMsg := entry.stopMessage
			for sNum := range activeSenders {
				s.sendSignalMessage(httpClient, baseURL, phoneNumber, sNum, stopMsg)
			}
		}
		s.iLog.Info(fmt.Sprintf("signalPollLoop: stopped for channel=%s", entry.channelName))
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		receiveURL := fmt.Sprintf("%s/v1/receive/%s", baseURL, phoneNumber)
		req, err := http.NewRequestWithContext(ctx, "GET", receiveURL, nil)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("signalPollLoop: build request error channel=%s: %v", entry.channelName, err))
			continue
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			s.iLog.Error(fmt.Sprintf("signalPollLoop: receive error channel=%s: %v", entry.channelName, err))
			time.Sleep(5 * time.Second)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			s.iLog.Error(fmt.Sprintf(
				"signalPollLoop: API error %d channel=%s: %s",
				resp.StatusCode, entry.channelName, string(body),
			))
			time.Sleep(5 * time.Second)
			continue
		}

		var envelopes []signalEnvelope
		if err := json.Unmarshal(body, &envelopes); err != nil {
			// Empty array or null = no pending messages; not an error.
			continue
		}

		for _, env := range envelopes {
			if env.Envelope.DataMessage == nil || env.Envelope.DataMessage.Message == "" {
				continue // skip receipts, typing notifications, etc.
			}

			senderNumber := env.Envelope.SourceNumber
			if senderNumber == "" {
				senderNumber = env.Envelope.Source
			}
			senderName := env.Envelope.SourceName
			if senderName == "" {
				senderName = senderNumber
			}
			text := strings.TrimSpace(env.Envelope.DataMessage.Message)
			if text == "" {
				continue
			}

			s.iLog.Info(fmt.Sprintf(
				"signalPollLoop: message from sender=%s name=%s channel=%s text=%q",
				senderNumber, senderName, entry.channelName, text,
			))

			// Remember this sender so we can notify on shutdown.
			activeSenders[senderNumber] = true

			go func(sNum, sName, msgText string) {
				response, err := s.channelSvc.HandleInbound(ctx, entry.channelID, sNum, sName, msgText)
				if err != nil {
					s.iLog.Error(fmt.Sprintf(
						"signalPollLoop: HandleInbound error channel=%s sender=%s: %v",
						entry.channelName, sNum, err,
					))
					response = "Sorry, an error occurred processing your message. Please try again."
				}
				if response != "" {
					s.sendSignalMessage(httpClient, baseURL, phoneNumber, sNum, response)
				}
			}(senderNumber, senderName, text)
		}
	}
}

// sendSignalMessage sends a text reply to a Signal number via signal-cli REST API.
func (s *AgentMonitorService) sendSignalMessage(client *http.Client, baseURL, fromNumber, toNumber, text string) {
	apiURL := baseURL + "/v2/send"
	payload, _ := json.Marshal(map[string]interface{}{
		"message":    text,
		"number":     fromNumber,
		"recipients": []string{toNumber},
	})

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(payload)))
	if err != nil {
		s.iLog.Error(fmt.Sprintf("sendSignalMessage: build request error to=%s: %v", toNumber, err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("sendSignalMessage: send error to=%s: %v", toNumber, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.iLog.Error(fmt.Sprintf(
			"sendSignalMessage: API error %d to=%s: %s",
			resp.StatusCode, toNumber, string(body),
		))
		return
	}
	s.iLog.Info(fmt.Sprintf("sendSignalMessage: delivered to=%s", toNumber))
}
