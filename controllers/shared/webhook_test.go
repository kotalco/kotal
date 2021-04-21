package shared

import "testing"

func TestIsWebhookEnabled(t *testing.T) {
	expected := true
	got := IsWebhookEnabled()
	if got != expected {
		t.Errorf("Expected webhook enabled to be %t , got %t", expected, got)
	}
}
