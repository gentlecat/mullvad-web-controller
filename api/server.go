package api

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type Server struct {
	relays      *RelayLocationsHandler
	relayChange *RelayLocationChangeHandler
	ip          *IPHandler
}

func NewServer(dryRun bool) *Server {
	client := newHTTPClient()
	return &Server{
		relays:      NewRelayLocationsHandler(client, 24*time.Hour),
		relayChange: NewRelayLocationChangeHandler(dryRun),
		ip:          NewIPHandler(client),
	}
}

// Routes returns the handler for the API and the static content, with CSRF protection applied.
func (s *Server) Routes(static fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServerFS(static))
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /api/relays", s.relays.Handle)
	mux.HandleFunc("POST /api/relay/location", s.relayChange.Handle)
	mux.HandleFunc("GET /api/ip", s.ip.Handle)

	return http.NewCrossOriginProtection().Handler(mux)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// newHTTPClient returns a client for upstream APIs that retries transient failures.
// Retries are kept short since a user is waiting on the other end.
func newHTTPClient() *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.RetryMax = 2
	client.RetryWaitMin = 500 * time.Millisecond
	client.RetryWaitMax = 3 * time.Second
	client.Backoff = retryablehttp.LinearJitterBackoff
	client.HTTPClient.Timeout = 10 * time.Second
	client.Logger = slog.Default()
	return client
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		slog.Error("failed to marshal response", "error", err)
		http.Error(w, "Error marshalling response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}
