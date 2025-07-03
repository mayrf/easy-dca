// Package notifications provides notification functionality for the DCA application.
package notifications

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mayrf/easy-dca/internal/config"
)

// Notifier is an interface for sending notifications.
type Notifier interface {
	Notify(ctx context.Context, subject, message string) error
}

// NtfyNotifier sends notifications via ntfy.sh or a custom ntfy server.
type NtfyNotifier struct {
	Topic string
	URL   string
}

// Notify sends a notification via ntfy.
func (n *NtfyNotifier) Notify(ctx context.Context, subject, message string) error {
	if n.Topic == "" {
		return fmt.Errorf("ntfy topic is not set")
	}
	if n.URL == "" {
		return fmt.Errorf("ntfy URL is not set")
	}
	ntfyURL := fmt.Sprintf("%s/%s", strings.TrimRight(n.URL, "/"), n.Topic)
	req, err := http.NewRequestWithContext(ctx, "POST", ntfyURL, strings.NewReader(message))
	if err != nil {
		return err
	}
	if subject != "" {
		req.Header.Set("Title", subject)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy notification failed: %s", resp.Status)
	}
	return nil
}

// CreateNotifier creates a Notifier based on configuration.
func CreateNotifier(cfg config.Config) Notifier {
	switch strings.ToLower(cfg.NotifyMethod) {
	case "ntfy":
		if cfg.NotifyNtfyURL == "" {
			log.Print("Warning: NOTIFY_NTFY_URL is required for ntfy notifications but is not set. Notifications will be disabled.")
			return nil
		}
		return &NtfyNotifier{Topic: cfg.NotifyNtfyTopic, URL: cfg.NotifyNtfyURL}
	// Add more cases for other notification methods (slack, email, etc.)
	default:
		return nil
	}
} 