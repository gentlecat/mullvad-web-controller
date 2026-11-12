package api

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const relaysURL = "https://api.mullvad.net/app/v1/relays"

type Relay struct {
	Country     string `json:"country"`     // A two-letter country code
	City        string `json:"city"`        // A three-letter city code
	Name        string `json:"name"`        // Full city name
	CountryName string `json:"countryName"` // Full country name
}

type RelayLocationsHandler struct {
	client *retryablehttp.Client
	url    string
	ttl    time.Duration

	mu        sync.Mutex
	relays    []Relay
	fetchedAt time.Time
}

func NewRelayLocationsHandler(client *retryablehttp.Client, ttl time.Duration) *RelayLocationsHandler {
	return &RelayLocationsHandler{client: client, url: relaysURL, ttl: ttl}
}

func (h *RelayLocationsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	relays, err := h.getRelays(r.Context())
	if err != nil {
		slog.Error("failed to retrieve relays", "error", err)
		http.Error(w, "Error retrieving relays", http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, relays)
}

// getRelays returns cached relays, refreshing them once the cache expires.
// If a refresh fails, stale data is served rather than an error.
func (h *RelayLocationsHandler) getRelays(ctx context.Context) ([]Relay, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.relays != nil && time.Since(h.fetchedAt) < h.ttl {
		return h.relays, nil
	}

	relays, err := h.fetchRelays(ctx)
	if err != nil {
		if h.relays != nil {
			slog.Warn("failed to refresh relays, serving stale data", "error", err)
			return h.relays, nil
		}
		return nil, err
	}

	h.relays = relays
	h.fetchedAt = time.Now()
	slog.Info("updated relay cache", "count", len(relays))
	return relays, nil
}

func (h *RelayLocationsHandler) fetchRelays(ctx context.Context) ([]Relay, error) {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, h.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relays: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	var apiResponse struct {
		Locations map[string]struct {
			Country string `json:"country"`
			City    string `json:"city"`
		} `json:"locations"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	relays := make([]Relay, 0, len(apiResponse.Locations))
	for key, location := range apiResponse.Locations {
		// Split key on "-" to get country and city codes
		parts := strings.Split(key, "-")
		if len(parts) != 2 {
			continue
		}

		relays = append(relays, Relay{
			Country:     parts[0],
			City:        parts[1],
			Name:        location.City,
			CountryName: location.Country,
		})
	}

	slices.SortFunc(relays, func(a, b Relay) int {
		return cmp.Or(cmp.Compare(a.CountryName, b.CountryName), cmp.Compare(a.Name, b.Name))
	})

	return relays, nil
}
