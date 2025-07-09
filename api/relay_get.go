package api

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/patrickmn/go-cache"
	"net/http"
	"strings"
	"time"
)

type Relay struct {
	Country string `json:"country"` // A two-letter country code, or 'any'
	City    string `json:"city"`    // A three-letter city code
	Name    string `json:"name"`    // Full city name
}

type RelayLocationsHandler struct {
	cache *cache.Cache
}

func NewRelayLocationsHandler(cacheDuration time.Duration) *RelayLocationsHandler {
	return &RelayLocationsHandler{
		cache: cache.New(cacheDuration, time.Hour),
	}
}

func (h *RelayLocationsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if cachedData, found := h.cache.Get("relays"); found {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(cachedData.([]byte))
		fmt.Printf("Served cached relay data\n")
		return
	}

	relays, err := getRelays()
	if err != nil {
		http.Error(w, "Error retrieving relays", http.StatusInternalServerError)
		fmt.Printf("Error retrieving relays %v\n", err)
		return
	}

	response, err := json.Marshal(relays)
	if err != nil {
		http.Error(w, "Error marshalling response", http.StatusInternalServerError)
		fmt.Printf("Error marshalling response %v\n", err)
		return
	}

	h.cache.Set("relays", response, cache.DefaultExpiration)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

	fmt.Printf("Retrieved %d relays and updated cache\n", len(relays))
}

func getRelays() ([]Relay, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 30 * time.Second
	retryClient.Backoff = retryablehttp.LinearJitterBackoff

	resp, err := retryClient.Get("https://api.mullvad.net/app/v1/relays")

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

	var relays []Relay
	for key, location := range apiResponse.Locations {
		// Split key on "-" to get country and city codes
		parts := strings.Split(key, "-")
		if len(parts) != 2 {
			continue
		}

		relay := Relay{
			Country: parts[0],
			City:    parts[1],
			Name:    location.City,
		}
		relays = append(relays, relay)
	}

	return relays, nil
}
