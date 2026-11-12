package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	maxRequestBodySize = 1 << 10
	commandTimeout     = 10 * time.Second
)

var (
	countryCodePattern = regexp.MustCompile(`^[a-z]{2}$`)
	cityCodePattern    = regexp.MustCompile(`^[a-z]{3}$`)
)

type RelayLocationChangeRequest struct {
	Country string `json:"country"` // A two-letter country code
	City    string `json:"city"`    // A three-letter city code, or empty for any city in the country
}

func (r RelayLocationChangeRequest) validate() error {
	if !countryCodePattern.MatchString(r.Country) {
		return errors.New("invalid country code")
	}
	if r.City != "" && !cityCodePattern.MatchString(r.City) {
		return errors.New("invalid city code")
	}
	return nil
}

type RelayLocationChangeResponse struct {
	Message string `json:"message"`
}

// CommandRunner runs a command and returns its combined output.
type CommandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

func runCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

type RelayLocationChangeHandler struct {
	dryRun     bool
	runCommand CommandRunner
}

func NewRelayLocationChangeHandler(dryRun bool) *RelayLocationChangeHandler {
	return &RelayLocationChangeHandler{dryRun: dryRun, runCommand: runCommand}
}

func (h *RelayLocationChangeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mediaType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodySize))
	decoder.DisallowUnknownFields()

	var req RelayLocationChangeRequest
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if h.dryRun {
		slog.Info("dry run, not switching relay", "country", req.Country, "city", req.City)
	} else {
		if err := h.changeRelay(r.Context(), req); err != nil {
			slog.Error("failed to change relay", "error", err)
			http.Error(w, "Failed to change relay", http.StatusBadGateway)
			return
		}
		slog.Info("switched relay", "country", req.Country, "city", req.City)
	}

	writeJSON(w, http.StatusOK, RelayLocationChangeResponse{Message: "Done"})
}

func (h *RelayLocationChangeHandler) changeRelay(ctx context.Context, req RelayLocationChangeRequest) error {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	args := []string{"relay", "set", "location", req.Country}
	if req.City != "" {
		args = append(args, req.City)
	}

	output, err := h.runCommand(ctx, "mullvad", args...)
	if err != nil {
		return fmt.Errorf("mullvad %s: %w: %s", strings.Join(args, " "), err, bytes.TrimSpace(output))
	}
	return nil
}
