package api

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"net/http"
	"time"
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
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 30 * time.Second
	retryClient.Backoff = retryablehttp.LinearJitterBackoff

	resp, err := retryClient.Get("https://ipinfo.io/json")
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
		http.Error(w, fmt.Sprintf("Could not retrieve IP info: %+v", err), http.StatusInternalServerError)
		fmt.Printf("Failed to retrieve current IP info: %+v\n", err)
		return
	}

	jsonResponse, err := json.Marshal(ipInfo)
	if err != nil {
		http.Error(w, "Error marshalling response", http.StatusInternalServerError)
		fmt.Printf("Failed to serialize current IP info: %+v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

	fmt.Printf("Retrieved current IP info: %+v\n", ipInfo)
}
