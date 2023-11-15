package checks

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type PingCheck struct {
	URL     string
	Method  string
	Timeout int
	client  http.Client
	Body    io.Reader
	Headers map[string]string
	Error   error
}

func CheckPingStatus(URL, Method string, Timeout int, Body io.Reader, Headers map[string]string) error {
	check := NewPingCheck(URL, Method, Timeout, Body, Headers)
	return check.checkstatus()
}

func NewPingCheck(URL, Method string, Timeout int, Body io.Reader, Headers map[string]string) PingCheck {
	if Method == "" {
		Method = "GET"
	}

	if Timeout == 0 {
		Timeout = 500
	}

	pingCheck := PingCheck{
		URL:     URL,
		Method:  Method,
		Timeout: Timeout,
		Body:    Body,
		Headers: Headers,
		Error:   nil,
	}
	pingCheck.client = http.Client{
		Timeout: time.Duration(Timeout) * time.Millisecond,
	}

	return pingCheck
}

func (p PingCheck) checkstatus() error {
	req, err := http.NewRequest(p.Method, p.URL, p.Body)

	if err != nil {
		p.Error = fmt.Errorf("Error creating request: %s", err)
		return fmt.Errorf("Error creating request: %s", err)
	}

	for key, value := range p.Headers {
		req.Header.Add(key, value)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		p.Error = fmt.Errorf("Error sending request: %s", err)
		return p.Error
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		p.Error = fmt.Errorf("Error response status code: %d", resp.StatusCode)
		return p.Error
	}
	return nil
}

func (p PingCheck) Name() string {
	return "ping-" + p.URL
}
