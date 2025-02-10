package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCheckProbe(t *testing.T) {
	tests := []struct {
		name        string
		config      ProbeConfig
		handler     http.HandlerFunc
		wantSuccess bool
		wantError   string
	}{
		{
			name: "successful health check with expected status",
			config: ProbeConfig{
				Type:                 "http",
				Path:                 "/health",
				Port:                 8080,
				ExpectedStatus:       200,
				TargetServiceAddress: "localhost",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "OK")
			},
			wantSuccess: true,
			wantError:   "",
		},
		{
			name: "failed health check with unexpected status",
			config: ProbeConfig{
				Type:                 "http",
				Path:                 "/health",
				Port:                 8080,
				ExpectedStatus:       200,
				TargetServiceAddress: "localhost",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintln(w, "NOT OK")
			},
			wantSuccess: false,
			wantError:   "unexpected status code: got 503, expected 200",
		},
		{
			name: "successful health check with regex match",
			config: ProbeConfig{
				Type:                 "http",
				Path:                 "/health",
				Port:                 8080,
				ExpectedStatus:       200,
				TargetServiceAddress: "localhost",
				ExpectedBodyRegex:    "OK",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "OK")
			},
			wantSuccess: true,
			wantError:   "",
		},
		{
			name: "failed health check with regex mismatch",
			config: ProbeConfig{
				Type:                 "http",
				Path:                 "/health",
				Port:                 8080,
				ExpectedStatus:       200,
				TargetServiceAddress: "localhost",
				ExpectedBodyRegex:    "OK",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "FAILED")
			},
			wantSuccess: false,
			wantError:   "regex did not match",
		},
		{
			name: "successful health check with negated regex mismatch",
			config: ProbeConfig{
				Type:                    "http",
				Path:                    "/health",
				Port:                    8080,
				ExpectedStatus:          200,
				TargetServiceAddress:    "localhost",
				ExpectedBodyRegex:       "OK",
				NegateExpectedBodyRegex: true,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "FAILED")
			},
			wantSuccess: true,
			wantError:   "",
		},
		{
			name: "failed health check with negated regex match",
			config: ProbeConfig{
				Type:                    "http",
				Path:                    "/health",
				Port:                    8080,
				ExpectedStatus:          200,
				TargetServiceAddress:    "localhost",
				ExpectedBodyRegex:       "OK",
				NegateExpectedBodyRegex: true,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "OK")
			},
			wantSuccess: false,
			wantError:   "regex matched, but negation was expected",
		},
		{
			name: "invalid regex",
			config: ProbeConfig{
				Type:                 "http",
				Path:                 "/health",
				Port:                 8080,
				ExpectedStatus:       200,
				TargetServiceAddress: "localhost",
				ExpectedBodyRegex:    "?!",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "OK")
			},
			wantSuccess: false,
			wantError:   "invalid regex",
		},
		{
			name: "unsupported probe type",
			config: ProbeConfig{
				Type: "invalid",
			},
			wantSuccess: false,
			wantError:   "unsupported probe type: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			tt.config.TargetServiceAddress = "localhost"
			tt.config.Port = server.Listener.Addr().(*net.TCPAddr).Port

			gotSuccess, err := checkProbe(tt.config)
			if gotSuccess != tt.wantSuccess {
				t.Errorf("checkProbe() gotSuccess = %v, want %v", gotSuccess, tt.wantSuccess)
			}
			if err != nil && tt.wantError == "" {
				t.Errorf("checkProbe() unexpected error = %v", err)
			}
			if err == nil && tt.wantError != "" {
				t.Errorf("checkProbe() expected error = %v, got nil", tt.wantError)
			}
			if err != nil && tt.wantError != "" && !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("checkProbe() error = %v, want substring %v", err, tt.wantError)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		_, err := loadConfig("config/config.yaml")
		if err != nil {
			t.Errorf("loadConfig() error = %v, want nil", err)
		}
	})

	t.Run("invalid config file", func(t *testing.T) {
		// Create a temporary file with invalid YAML content
		tmpfile, err := os.CreateTemp("", "invalid_config.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name()) // clean up

		if _, err := tmpfile.Write([]byte("invalid yaml")); err != nil {
			t.Fatal(err)
		}
		if err := tmpfile.Close(); err != nil {
			t.Fatal(err)
		}

		_, err = loadConfig(tmpfile.Name())
		if err == nil {
			t.Errorf("loadConfig() with invalid YAML, expected error, got nil")
		}
	})

	t.Run("config file not found", func(t *testing.T) {
		_, err := loadConfig("nonexistent_config.yaml")
		if err == nil {
			t.Errorf("loadConfig() with nonexistent file, expected error, got nil")
		}
	})
}
