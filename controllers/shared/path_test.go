package shared

import "testing"

const (
	testHomeDir = "/users/test"
)

func TestPathData(t *testing.T) {
	expected := "/users/test/kotal-data"
	got := PathData(testHomeDir)

	if got != expected {
		t.Errorf("expected data directory to be %s, got %s", expected, got)
	}
}

func TestPathConfig(t *testing.T) {
	expected := "/users/test/kotal-config"
	got := PathConfig(testHomeDir)

	if got != expected {
		t.Errorf("expected configuration directory to be %s, got %s", expected, got)
	}
}

func TestPathSecrets(t *testing.T) {
	expected := "/users/test/.kotal-secrets"
	got := PathSecrets(testHomeDir)

	if got != expected {
		t.Errorf("expected secrets directory to be %s, got %s", expected, got)
	}
}
