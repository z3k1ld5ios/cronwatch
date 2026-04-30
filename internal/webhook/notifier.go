package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AlertType categorizes the kind of alert being sent.
type AlertType string

const (
	AlertMissedRun AlertType = "missed_run"
	AlertDrift     AlertType = "drift"
)

// Payload is the JSON body sent to the webhook endpoint.
type Payload struct {
	JobName   string    `json:"job_name"`
	AlertType AlertType `json:"alert_type"`
	Message   string    `json:"message"`
	OccuredAt time.Time `json:"occured_at"`
	DriftSecs float64   `json:"drift_seconds,omitempty"`
}

// Notifier sends alert payloads to a configured webhook URL.
type Notifier struct {
	URL    string
	client *http.Client
}

// NewNotifier creates a Notifier with a sensible HTTP timeout.
func NewNotifier(url string) *Notifier {
	return &Notifier{
		URL: url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send marshals the payload and POSTs it to the webhook URL.
func (n *Notifier) Send(p Payload) error {
	if n.URL == "" {
		return fmt.Errorf("webhook: URL is not configured")
	}

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: POST failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}
