package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const relaysResponse = `{
	"locations": {
		"se-got": {"country": "Sweden", "city": "Gothenburg"},
		"de-ber": {"country": "Germany", "city": "Berlin"},
		"se-sto": {"country": "Sweden", "city": "Stockholm"},
		"de-fra": {"country": "Germany", "city": "Frankfurt"},
		"invalid": {"country": "Nowhere", "city": "Nowhere"}
	}
}`

// newTestClient returns a client that doesn't retry, so failures surface immediately.
func newTestClient() *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.RetryMax = 0
	client.Logger = nil
	return client
}

// newUpstream starts a fake upstream API. Responses fail with 500 while fail is set.
func newUpstream(t *testing.T, body string, fail *atomic.Bool) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if fail != nil && fail.Load() {
			http.Error(w, "unavailable", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server, &calls
}

func getRelays(t *testing.T, h *RelayLocationsHandler) (int, []Relay) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.Handle(rec, httptest.NewRequest(http.MethodGet, "/api/relays", nil))
	if rec.Code != http.StatusOK {
		return rec.Code, nil
	}
	var relays []Relay
	if err := json.Unmarshal(rec.Body.Bytes(), &relays); err != nil {
		t.Fatalf("failed to decode relays: %v", err)
	}
	return rec.Code, relays
}

func TestRelayLocationsHandler(t *testing.T) {
	upstream, calls := newUpstream(t, relaysResponse, nil)
	h := NewRelayLocationsHandler(newTestClient(), time.Hour)
	h.url = upstream.URL

	code, relays := getRelays(t, h)
	if code != http.StatusOK {
		t.Fatalf("status = %d, want %d", code, http.StatusOK)
	}

	want := []Relay{
		{Country: "de", City: "ber", Name: "Berlin", CountryName: "Germany"},
		{Country: "de", City: "fra", Name: "Frankfurt", CountryName: "Germany"},
		{Country: "se", City: "got", Name: "Gothenburg", CountryName: "Sweden"},
		{Country: "se", City: "sto", Name: "Stockholm", CountryName: "Sweden"},
	}
	if !reflect.DeepEqual(relays, want) {
		t.Errorf("relays = %+v, want %+v", relays, want)
	}

	getRelays(t, h)
	if got := calls.Load(); got != 1 {
		t.Errorf("upstream calls = %d, want 1 (second request should be cached)", got)
	}
}

func TestRelayLocationsHandlerServesStaleData(t *testing.T) {
	var fail atomic.Bool
	upstream, calls := newUpstream(t, relaysResponse, &fail)
	// A zero TTL forces a refresh on every request.
	h := NewRelayLocationsHandler(newTestClient(), 0)
	h.url = upstream.URL

	_, fresh := getRelays(t, h)

	fail.Store(true)
	code, stale := getRelays(t, h)
	if code != http.StatusOK {
		t.Fatalf("status = %d, want %d", code, http.StatusOK)
	}
	if !reflect.DeepEqual(stale, fresh) {
		t.Errorf("stale relays = %+v, want %+v", stale, fresh)
	}
	if got := calls.Load(); got != 2 {
		t.Errorf("upstream calls = %d, want 2", got)
	}
}

func TestRelayLocationsHandlerUpstreamError(t *testing.T) {
	var fail atomic.Bool
	fail.Store(true)
	upstream, _ := newUpstream(t, relaysResponse, &fail)
	h := NewRelayLocationsHandler(newTestClient(), time.Hour)
	h.url = upstream.URL

	if code, _ := getRelays(t, h); code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", code, http.StatusBadGateway)
	}
}

func TestRelayLocationChangeHandler(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		dryRun      bool
		commandErr  error
		wantStatus  int
		wantArgs    []string
	}{
		{
			name:        "country and city",
			contentType: "application/json",
			body:        `{"country":"se","city":"got"}`,
			wantStatus:  http.StatusOK,
			wantArgs:    []string{"relay", "set", "location", "se", "got"},
		},
		{
			name:        "country only",
			contentType: "application/json; charset=utf-8",
			body:        `{"country":"se"}`,
			wantStatus:  http.StatusOK,
			wantArgs:    []string{"relay", "set", "location", "se"},
		},
		{
			name:        "dry run",
			contentType: "application/json",
			body:        `{"country":"se","city":"got"}`,
			dryRun:      true,
			wantStatus:  http.StatusOK,
		},
		{name: "flag as country", contentType: "application/json", body: `{"country":"-h","city":"got"}`, wantStatus: http.StatusBadRequest},
		{name: "flag as city", contentType: "application/json", body: `{"country":"se","city":"--x"}`, wantStatus: http.StatusBadRequest},
		{name: "uppercase country", contentType: "application/json", body: `{"country":"SE","city":"got"}`, wantStatus: http.StatusBadRequest},
		{name: "any country", contentType: "application/json", body: `{"country":"any"}`, wantStatus: http.StatusBadRequest},
		{name: "missing country", contentType: "application/json", body: `{"city":"got"}`, wantStatus: http.StatusBadRequest},
		{name: "unknown field", contentType: "application/json", body: `{"country":"se","host":"x"}`, wantStatus: http.StatusBadRequest},
		{name: "malformed body", contentType: "application/json", body: `{"country":`, wantStatus: http.StatusBadRequest},
		{name: "oversized body", contentType: "application/json", body: `{"country":"` + strings.Repeat("a", 2000) + `"}`, wantStatus: http.StatusBadRequest},
		{name: "text/plain", contentType: "text/plain", body: `{"country":"se","city":"got"}`, wantStatus: http.StatusUnsupportedMediaType},
		{name: "no content type", body: `{"country":"se","city":"got"}`, wantStatus: http.StatusUnsupportedMediaType},
		{
			name:        "command fails",
			contentType: "application/json",
			body:        `{"country":"se","city":"got"}`,
			commandErr:  errors.New("exit status 1"),
			wantStatus:  http.StatusBadGateway,
			wantArgs:    []string{"relay", "set", "location", "se", "got"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotArgs []string
			h := NewRelayLocationChangeHandler(tt.dryRun)
			h.runCommand = func(ctx context.Context, name string, args ...string) ([]byte, error) {
				if name != "mullvad" {
					t.Errorf("command = %q, want mullvad", name)
				}
				gotArgs = args
				return []byte("output"), tt.commandErr
			}

			req := httptest.NewRequest(http.MethodPost, "/api/relay/location", strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rec := httptest.NewRecorder()
			h.Handle(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %q)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("args = %q, want %q", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestIPHandler(t *testing.T) {
	upstream, _ := newUpstream(t, `{"ip":"1.2.3.4","city":"Gothenburg","country":"SE","org":"Mullvad"}`, nil)
	h := NewIPHandler(newTestClient())
	h.url = upstream.URL

	rec := httptest.NewRecorder()
	h.Handle(rec, httptest.NewRequest(http.MethodGet, "/api/ip", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var info IPInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &info); err != nil {
		t.Fatalf("failed to decode IP info: %v", err)
	}
	want := IPInfo{IP: "1.2.3.4", City: "Gothenburg", Country: "SE", Org: "Mullvad"}
	if info != want {
		t.Errorf("IP info = %+v, want %+v", info, want)
	}
}

func TestRoutes(t *testing.T) {
	s := NewServer(true)
	s.relayChange.runCommand = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		t.Error("command should not run in dry-run mode")
		return nil, nil
	}
	static := fstest.MapFS{"index.html": {Data: []byte("<html></html>")}}
	handler := s.Routes(static)

	tests := []struct {
		name       string
		method     string
		path       string
		headers    map[string]string
		wantStatus int
	}{
		{name: "index", method: http.MethodGet, path: "/", wantStatus: http.StatusOK},
		{name: "health", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK},
		{name: "wrong method", method: http.MethodDelete, path: "/api/ip", wantStatus: http.StatusMethodNotAllowed},
		{name: "put relay change", method: http.MethodPut, path: "/api/relay/location", wantStatus: http.StatusMethodNotAllowed},
		{
			name:       "same-origin change",
			method:     http.MethodPost,
			path:       "/api/relay/location",
			headers:    map[string]string{"Content-Type": "application/json", "Sec-Fetch-Site": "same-origin"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "cross-site change",
			method:     http.MethodPost,
			path:       "/api/relay/location",
			headers:    map[string]string{"Content-Type": "application/json", "Sec-Fetch-Site": "cross-site"},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{"country":"se","city":"got"}`))
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %q)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
