//go:build windows
// +build windows

package web

import (
	"strings"
	"testing"
)

func TestStart_ReturnsError(t *testing.T) {
	err := Start(8080)
	if err == nil {
		t.Fatal("Expected error on Windows, got nil")
	}
	if !strings.Contains(err.Error(), "not supported on Windows") {
		t.Errorf("Expected error message about Windows support, got: %v", err)
	}
}

func TestStartWithLogging_ReturnsError(t *testing.T) {
	err := StartWithLogging(8080, true)
	if err == nil {
		t.Fatal("Expected error on Windows, got nil")
	}
	if !strings.Contains(err.Error(), "not supported on Windows") {
		t.Errorf("Expected error message about Windows support, got: %v", err)
	}
}
