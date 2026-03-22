package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddlewareAllowsConfiguredOrigin(t *testing.T) {
	middleware := NewCORSMiddleware(CORSOptions{
		AllowedOrigins: []string{"https://algomind.pro/"},
		AllowedMethods: []string{http.MethodPost, http.MethodOptions},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/event", nil)
	req.Header.Set("Origin", "https://algomind.pro")

	rec := httptest.NewRecorder()
	called := false
	middleware(func(w http.ResponseWriter, r *http.Request) {
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
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "POST, OPTIONS" {
		t.Fatalf("unexpected allow-methods header: got %q", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("unexpected vary header: got %q", got)
	}
}

func TestCORSMiddlewareRejectsDisallowedOrigin(t *testing.T) {
	middleware := NewCORSMiddleware(CORSOptions{
		AllowedOrigins: []string{"https://algomind.pro"},
		AllowedMethods: []string{http.MethodPost, http.MethodOptions},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/event", nil)
	req.Header.Set("Origin", "https://evil.example")

	rec := httptest.NewRecorder()
	called := false
	middleware(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusAccepted)
	})(rec, req)

	if called {
		t.Fatal("expected wrapped handler not to be called")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("unexpected status code: got %d want %d", rec.Code, http.StatusForbidden)
	}
}

func TestCORSMiddlewareAllowsRequestsWithoutOrigin(t *testing.T) {
	middleware := NewCORSMiddleware(CORSOptions{
		AllowedOrigins: []string{"https://analytics.algomind.pro"},
		AllowedMethods: []string{http.MethodGet, http.MethodOptions},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)

	rec := httptest.NewRecorder()
	called := false
	middleware(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})(rec, req)

	if !called {
		t.Fatal("expected wrapped handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no allow-origin header, got %q", got)
	}
}

func TestCORSMiddlewareAllowsAnyOriginWhenUnset(t *testing.T) {
	middleware := NewCORSMiddleware(CORSOptions{
		AllowedMethods: []string{http.MethodPost, http.MethodOptions},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/event", nil)
	req.Header.Set("Origin", "https://sahirashifal.com")

	rec := httptest.NewRecorder()
	middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("unexpected status code: got %d want %d", rec.Code, http.StatusAccepted)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("unexpected allow-origin header: got %q", got)
	}
}
