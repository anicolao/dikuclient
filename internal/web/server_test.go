//go:build !windows
// +build !windows

package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleRoot_NoSessionID_NoCookie(t *testing.T) {
	server := NewServer(8080)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.handleRoot(w, req)

	// Should redirect to a new session
	if w.Code != http.StatusFound {
		t.Errorf("expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.HasPrefix(location, "/?id=") {
		t.Errorf("expected redirect to /?id=<uuid>, got %s", location)
	}

	// Should not be "/?id=new"
	if location == "/?id=new" {
		t.Errorf("should not redirect to /?id=new, should be a UUID")
	}
}

func TestHandleRoot_NoSessionID_WithCookie(t *testing.T) {
	server := NewServer(8080)
	req := httptest.NewRequest("GET", "/", nil)
	// Add cookie with last session ID
	req.AddCookie(&http.Cookie{
		Name:  "dikuclient_last_session",
		Value: "test-session-123",
	})
	w := httptest.NewRecorder()

	server.handleRoot(w, req)

	// Should redirect to the last session
	if w.Code != http.StatusFound {
		t.Errorf("expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	expected := "/?id=test-session-123"
	if location != expected {
		t.Errorf("expected redirect to %s, got %s", expected, location)
	}
}

func TestHandleRoot_ExplicitNew(t *testing.T) {
	server := NewServer(8080)
	req := httptest.NewRequest("GET", "/?id=new", nil)
	// Add cookie to verify it's ignored when id=new
	req.AddCookie(&http.Cookie{
		Name:  "dikuclient_last_session",
		Value: "test-session-123",
	})
	w := httptest.NewRecorder()

	server.handleRoot(w, req)

	// Should redirect to a new session (not the cookie value)
	if w.Code != http.StatusFound {
		t.Errorf("expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.HasPrefix(location, "/?id=") {
		t.Errorf("expected redirect to /?id=<uuid>, got %s", location)
	}

	// Should not redirect to the cookie value
	if location == "/?id=test-session-123" {
		t.Errorf("should not redirect to cookie value, should create new UUID")
	}

	// Should not be "/?id=new"
	if location == "/?id=new" {
		t.Errorf("should not redirect to /?id=new, should be a UUID")
	}
}

func TestHandleRoot_WithSpecificSessionID(t *testing.T) {
	server := NewServer(8080)
	req := httptest.NewRequest("GET", "/?id=my-custom-session", nil)
	w := httptest.NewRecorder()

	server.handleRoot(w, req)

	// Should not redirect (will try to serve the page)
	// The actual status code depends on whether the file exists
	// but we should NOT get a redirect (302)
	if w.Code == http.StatusFound {
		t.Errorf("should not redirect when session ID is provided")
	}

	// Should not redirect
	location := w.Header().Get("Location")
	if location != "" {
		t.Errorf("expected no redirect, got redirect to %s", location)
	}
}

func TestHandleRoot_EmptyCookieValue(t *testing.T) {
	server := NewServer(8080)
	req := httptest.NewRequest("GET", "/", nil)
	// Add cookie with empty value
	req.AddCookie(&http.Cookie{
		Name:  "dikuclient_last_session",
		Value: "",
	})
	w := httptest.NewRecorder()

	server.handleRoot(w, req)

	// Should create a new session (ignore empty cookie)
	if w.Code != http.StatusFound {
		t.Errorf("expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.HasPrefix(location, "/?id=") {
		t.Errorf("expected redirect to /?id=<uuid>, got %s", location)
	}
}

func TestHandleRoot_NewPath(t *testing.T) {
	server := NewServer(8080)
	req := httptest.NewRequest("GET", "/new", nil)
	// Add cookie to verify it's ignored when using /new path
	req.AddCookie(&http.Cookie{
		Name:  "dikuclient_last_session",
		Value: "test-session-123",
	})
	w := httptest.NewRecorder()

	server.handleRoot(w, req)

	// Should redirect to a new session (not the cookie value)
	if w.Code != http.StatusFound {
		t.Errorf("expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.HasPrefix(location, "/?id=") {
		t.Errorf("expected redirect to /?id=<uuid>, got %s", location)
	}

	// Should not redirect to the cookie value
	if location == "/?id=test-session-123" {
		t.Errorf("should not redirect to cookie value, should create new UUID")
	}

	// Should not be "/?id=new"
	if location == "/?id=new" {
		t.Errorf("should not redirect to /?id=new, should be a UUID")
	}
}
