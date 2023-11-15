package checks

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const defaultRequestTimeout = 5 * time.Second

// Config is the HTTP checker configuration settings container.
type HttpCheck struct {
	// Ctx is the context that will be used for the health check.
	Ctx context.Context
	// URL is the remote service health check URL.
	URL string
	// RequestTimeout is the duration that health check will try to consume published test message.
	// If not set - 5 seconds
	RequestTimeout time.Duration

	Error error
}

func CheckHttpStatus(ctx context.Context, url string, timeout time.Duration) error {
	check := NewHttpCheck(ctx, url, timeout)
	return check.CheckStatus()
}

func NewHttpCheck(ctx context.Context, url string, timeout time.Duration) HttpCheck {
	if timeout == 0 {
		timeout = defaultRequestTimeout
	}

	return HttpCheck{
		Ctx:            ctx,
		URL:            url,
		RequestTimeout: timeout,
		Error:          nil,
	}
}

// New creates new HTTP service health check that verifies the following:
// - connection establishing
// - getting response status from defined URL
// - verifying that status code is less than 500
func (check HttpCheck) CheckStatus() error {
	config := check
	ctx := config.Ctx
	req, err := http.NewRequest(http.MethodGet, check.URL, nil)
	if err != nil {
		check.Error = fmt.Errorf("creating the request for the health check failed: %w", err)
		return check.Error
	}

	ctx, cancel := context.WithTimeout(ctx, config.RequestTimeout)
	defer cancel()

	// Inform remote service to close the connection after the transaction is complete
	req.Header.Set("Connection", "close")
	req = req.WithContext(ctx)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		check.Error = fmt.Errorf("making the request for the health check failed: %w", err)
		return check.Error
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusInternalServerError {
		check.Error = fmt.Errorf("remote service returned status code %d", res.StatusCode)
		return check.Error
	}

	return nil
}
