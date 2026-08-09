package reliability

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestLabServerRegisterSiteUsesAdminAPI(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/sites" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-admin-token" {
			t.Errorf("authorization = %q", got)
		}
		var site struct {
			SiteID        string   `json:"site_id"`
			Timezone      string   `json:"timezone"`
			RetentionDays int      `json:"retention_days"`
			Domains       []string `json:"domains"`
		}
		if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
			t.Errorf("decode site: %v", err)
		}
		if site.SiteID != "profile-site" || site.Timezone != "UTC" || site.RetentionDays != 365 {
			t.Errorf("unexpected site: %+v", site)
		}
		if len(site.Domains) != 1 || site.Domains[0] != domainForSite("profile-site") {
			t.Errorf("unexpected domains: %v", site.Domains)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(server.Close)

	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse test server URL: %v", err)
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		t.Fatalf("parse test server port: %v", err)
	}
	labServer := &LabServer{Port: port, AdminToken: "test-admin-token"}
	if err := labServer.RegisterSite(context.Background(), "profile-site"); err != nil {
		t.Fatalf("RegisterSite returned error: %v", err)
	}
}

func TestLabServerRegisterSiteReportsAPIRejection(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse test server URL: %v", err)
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		t.Fatalf("parse test server port: %v", err)
	}
	labServer := &LabServer{Port: port, AdminToken: "wrong-token"}
	err = labServer.RegisterSite(context.Background(), "profile-site")
	if err == nil || err.Error() != "register site returned HTTP 401: Unauthorized" {
		t.Fatalf("RegisterSite error = %v", err)
	}
}
