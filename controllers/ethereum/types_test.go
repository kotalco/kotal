package controllers

import (
	"os"
	"testing"
)

func TestBesuImage(t *testing.T) {
	// without environment variables
	expected := DefaultBesuImage
	got := BesuImage()
	if got != expected {
		t.Errorf("Expecting besu image to be %s got %s", expected, got)
	}
	// with environment variables
	expected = "kotalco/besu:v2.0"
	os.Setenv(EnvBesuImage, expected)
	got = BesuImage()
	if got != expected {
		t.Errorf("Expecting besu image to be %s got %s", expected, got)
	}
}

func TestGethImage(t *testing.T) {
	// without environment variables
	expected := DefaultGethImage
	got := GethImage()
	if got != expected {
		t.Errorf("Expecting besu image to be %s got %s", expected, got)
	}
	// with environment variables
	expected = "kotalco/geth:v2.0"
	os.Setenv(EnvGethImage, expected)
	got = GethImage()
	if got != expected {
		t.Errorf("Expecting besu image to be %s got %s", expected, got)
	}
}
