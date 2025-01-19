package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
)

type Request struct {
	Country string `json:"country"` // A two-letter country code, or 'any'
	City    string `json:"city"`    // A three-letter city code
}

type Response struct {
	Message string `json:"message"`
}

type RelayLocationChangeHandler struct {
	devMode bool
}

func NewRelayLocationChangeHandler(devMode bool) *RelayLocationChangeHandler {
	return &RelayLocationChangeHandler{devMode: devMode}
}

func (h *RelayLocationChangeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req Request
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !h.devMode {
		err = changeRelay(req.Country, req.City)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	resp := Response{
		Message: "Done",
	}
	jsonResponse, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Error marshalling response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func changeRelay(country string, city string) error {
	if len(country) != 2 {
		return errors.New("invalid country code")
	}

	if len(city) != 3 {
		return errors.New("invalid city code")
	}

	cmd := exec.Command("mullvad", "relay", "set", "location", country, city)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to change relay: %w", err)
	}
	return nil
}
