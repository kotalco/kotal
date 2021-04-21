package shared

import "os"

// IsWebhookEnabled checks if webhooks are enabled
func IsWebhookEnabled() bool {
	return os.Getenv("ENABLE_WEBHOOKS") != "false"
}
