package shared

import (
	"testing"
)

func TestHost(t *testing.T) {
	expected := "127.0.0.1"
	got := Host(false)

	if got != expected {
		t.Errorf("expected host to be %s but got %s", expected, got)
	}
}
