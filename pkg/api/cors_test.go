package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddlewareAllowsAllOrigins(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/event", nil)
	req.Header.Set("Origin", "https://algomind.pro")

	rec := httptest.NewRecorder()
	called := false
	NewCORSMiddleware(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusAccepted)
	})(rec, req)

	if !called {
		t.Fatal("expected wrapped handler to be called")
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("unexpected status code: got %d want %d", rec.Code, http.StatusAccepted)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://algomind.pro" {
		t.Fatalf("unexpected allow-origin header: got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, OPTIONS" {
		t.Fatalf("unexpected allow-methods header: got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type" {
		t.Fatalf("unexpected allow-headers header: got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("unexpected allow-credentials header: got %q", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("unexpected vary header: got %q", got)
	}
}

func TestCORSMiddlewareHandlesPreflight(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/events", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	called := false
	NewCORSMiddleware(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})(rec, req)

	if called {
		t.Fatal("expected wrapped handler not to be called for preflight")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://evil.example" {
		t.Fatalf("unexpected allow-origin header: got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type" {
		t.Fatalf("unexpected allow-headers header: got %q", got)
	}
}

func TestCORSMiddlewareUsesWildcardWithoutOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/sites", nil)

	rec := httptest.NewRecorder()
	NewCORSMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("unexpected allow-origin header: got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("expected no allow-credentials header, got %q", got)
	}
}
