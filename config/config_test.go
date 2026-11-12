package config

import "testing"

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		dryRun  string
		want    Config
		wantErr bool
	}{
		{name: "defaults", want: Config{Addr: "0.0.0.0:8666"}},
		{name: "custom address", addr: "localhost:9000", want: Config{Addr: "localhost:9000"}},
		{name: "dry run enabled", dryRun: "1", want: Config{Addr: "0.0.0.0:8666", DryRun: true}},
		{name: "dry run disabled", dryRun: "false", want: Config{Addr: "0.0.0.0:8666"}},
		{name: "invalid dry run", dryRun: "yes please", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MULLVAD_WEB_ADDR", tt.addr)
			t.Setenv("MULLVAD_WEB_DRY_RUN", tt.dryRun)

			got, err := Load()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
