package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

// Status describes the state of the VPN tunnel.
type Status struct {
	State       string `json:"state"`                 // connected, connecting, disconnected, disconnecting or error
	Hostname    string `json:"hostname,omitempty"`    // Relay hostname, such as se-got-wg-001
	Country     string `json:"country,omitempty"`     // A two-letter country code, taken from the hostname
	City        string `json:"city,omitempty"`        // A three-letter city code, taken from the hostname
	CountryName string `json:"countryName,omitempty"` // Full country name
	CityName    string `json:"cityName,omitempty"`    // Full city name
	IPv4        string `json:"ipv4,omitempty"`
	IPv6        string `json:"ipv6,omitempty"`
}

type StatusHandler struct {
	runCommand CommandRunner
}

func NewStatusHandler(runCommand CommandRunner) *StatusHandler {
	return &StatusHandler{runCommand: runCommand}
}

func (h *StatusHandler) Handle(w http.ResponseWriter, r *http.Request) {
	status, err := h.getStatus(r.Context())
	if err != nil {
		slog.Error("failed to get tunnel status", "error", err)
		http.Error(w, "Could not get the tunnel status", http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, status)
}

func (h *StatusHandler) getStatus(ctx context.Context) (Status, error) {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	output, err := h.runCommand(ctx, "mullvad", "status", "--json")
	if err != nil {
		return Status{}, fmt.Errorf("mullvad status: %w: %s", err, bytes.TrimSpace(output))
	}

	return parseStatus(output)
}

func parseStatus(output []byte) (Status, error) {
	var cliStatus struct {
		State   string `json:"state"`
		Details struct {
			Location struct {
				Hostname string `json:"hostname"`
				City     string `json:"city"`
				Country  string `json:"country"`
				IPv4     string `json:"ipv4"`
				IPv6     string `json:"ipv6"`
			} `json:"location"`
		} `json:"details"`
	}

	if err := json.Unmarshal(output, &cliStatus); err != nil {
		return Status{}, fmt.Errorf("failed to decode mullvad status: %w", err)
	}

	location := cliStatus.Details.Location
	status := Status{
		State:       cliStatus.State,
		Hostname:    location.Hostname,
		CountryName: location.Country,
		CityName:    location.City,
		IPv4:        location.IPv4,
		IPv6:        location.IPv6,
	}

	// Relay hostnames start with the country and city codes, such as se-got-wg-001
	if parts := strings.SplitN(location.Hostname, "-", 3); len(parts) == 3 &&
		countryCodePattern.MatchString(parts[0]) && cityCodePattern.MatchString(parts[1]) {
		status.Country = parts[0]
		status.City = parts[1]
	}

	return status, nil
}
