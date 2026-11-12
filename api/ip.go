package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

const ipInfoURL = "https://ipinfo.io/json"

type IPInfo struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
}

type IPHandler struct {
	client *retryablehttp.Client
	url    string
}

func NewIPHandler(client *retryablehttp.Client) *IPHandler {
	return &IPHandler{client: client, url: ipInfoURL}
}

func (h *IPHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ipInfo, err := h.getCurrentIPDetails(r.Context())
	if err != nil {
		slog.Error("failed to retrieve current IP info", "error", err)
		http.Error(w, "Could not retrieve IP info", http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, ipInfo)
	slog.Info("retrieved current IP info", "ip", ipInfo.IP, "country", ipInfo.Country, "city", ipInfo.City)
}

func (h *IPHandler) getCurrentIPDetails(ctx context.Context) (IPInfo, error) {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, h.url, nil)
	if err != nil {
		return IPInfo{}, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return IPInfo{}, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return IPInfo{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data IPInfo
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return IPInfo{}, fmt.Errorf("error decoding JSON: %w", err)
	}

	return data, nil
}
