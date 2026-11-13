package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// dryRunConnectDelay is how long the fake tunnel reports "connecting" after a relay change.
const dryRunConnectDelay = 2 * time.Second

// dryRunMullvad stands in for the mullvad CLI in dry-run mode. It remembers the requested
// location and reports it back in the status, so the UI can be used without touching the real tunnel.
type dryRunMullvad struct {
	mu      sync.Mutex
	country string
	city    string
	setAt   time.Time
}

func newDryRunMullvad() *dryRunMullvad {
	return &dryRunMullvad{country: "se", city: "got"}
}

func (m *dryRunMullvad) run(ctx context.Context, name string, args ...string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch {
	case len(args) >= 4 && args[0] == "relay" && args[1] == "set" && args[2] == "location":
		slog.Info("dry run, not running command", "command", name+" "+strings.Join(args, " "))
		m.country = args[3]
		m.city = ""
		if len(args) >= 5 {
			m.city = args[4]
		}
		m.setAt = time.Now()
		return nil, nil

	case len(args) >= 1 && args[0] == "status":
		state := "connected"
		if time.Since(m.setAt) < dryRunConnectDelay {
			state = "connecting"
		}
		city := m.city
		if city == "" {
			// A real tunnel would pick a city; the fake one doesn't know any, so it uses a placeholder.
			city = "xxx"
		}
		return json.Marshal(map[string]any{
			"state": state,
			"details": map[string]any{
				"location": map[string]any{
					"hostname": fmt.Sprintf("%s-%s-wg-001", m.country, city),
					"ipv4":     "192.0.2.1",
				},
			},
		})
	}

	return nil, fmt.Errorf("dry run: unsupported command %q", strings.Join(args, " "))
}
