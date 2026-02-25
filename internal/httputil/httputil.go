package httputil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// DefaultClient is a shared HTTP client with a 30-second timeout.
var DefaultClient = &http.Client{Timeout: 30 * time.Second}

// PostJSON sends a JSON-encoded POST request to the given URL and checks for
// a successful status code (2xx). Non-2xx responses return an error containing
// the status code and response body.
func PostJSON(url string, payload interface{}) error {
	return PostJSONWithHeaders(url, payload, nil)
}

// PostJSONWithHeaders sends a JSON-encoded POST request with custom headers.
func PostJSONWithHeaders(url string, payload interface{}, headers map[string]string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	slog.Debug("http response", "url", url, "status", resp.StatusCode, "headers", resp.Header, "body", string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// DoRequest sends a custom HTTP request and checks for a successful status code.
// Use this for non-standard methods (PUT), form-encoded bodies, or special response
// handling that PostJSON doesn't cover.
func DoRequest(req *http.Request) error {
	resp, err := DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	slog.Debug("http response", "url", req.URL.String(), "status", resp.StatusCode, "headers", resp.Header, "body", string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
