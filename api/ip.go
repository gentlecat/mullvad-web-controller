package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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

func getCurrentIPDetails() (IPInfo, error) {
	resp, err := http.Get("https://ipinfo.io/json")
	if err != nil {
		return IPInfo{}, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return IPInfo{}, fmt.Errorf("error: unexpected status code: %d", resp.StatusCode)
	}

	var data IPInfo
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return IPInfo{}, fmt.Errorf("error decoding JSON: %w", err)
	}

	return data, nil
}

func HandleIPRetrieval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ipInfo, err := getCurrentIPDetails()
	if err != nil {
		http.Error(w, fmt.Sprintf("Changed the location but could not retrieve IP info: %+v", err), http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(ipInfo)
	if err != nil {
		http.Error(w, "Error marshalling response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
