package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cadbri/internal/config"
	"cadbri/internal/engine"
	"cadbri/internal/engine/engines"
	"cadbri/internal/network"
)

func TestHealthAndStatsEndpoints(t *testing.T) {
	engines.RegisterAll()
	cfg := config.DefaultConfig()
	client := network.NewClient(network.ClientOptions{Timeout: 2 * time.Second})
	srv := NewServer(cfg, client, engine.GlobalRegistry)

	// 1. Test /healthz
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	srv.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 on /healthz, got %d", rr.Code)
	}

	var healthResp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &healthResp); err != nil {
		t.Fatalf("failed to decode /healthz JSON: %v", err)
	}

	if healthResp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", healthResp["status"])
	}

	// 2. Test /autocompleter
	reqSug := httptest.NewRequest(http.MethodGet, "/autocompleter?q=face", nil)
	reqSug.Header.Set("X-Requested-With", "XMLHttpRequest")
	rrSug := httptest.NewRecorder()
	srv.Router().ServeHTTP(rrSug, reqSug)

	if rrSug.Code != http.StatusOK {
		t.Fatalf("expected status 200 on /autocompleter, got %d", rrSug.Code)
	}

	var suggestions []string
	if err := json.Unmarshal(rrSug.Body.Bytes(), &suggestions); err != nil {
		t.Fatalf("failed to decode autocompleter JSON: %v", err)
	}
	t.Logf("Autocompleter returned %d suggestions for 'face': %v", len(suggestions), suggestions)
}
