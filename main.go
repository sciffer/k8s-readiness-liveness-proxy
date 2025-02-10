package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type ProbeConfig struct {
	Type                    string `yaml:"type"`
	Path                    string `yaml:"path"`
	Port                    int    `yaml:"port"`
	ExpectedStatus          int    `yaml:"expected_status"`
	TargetServiceAddress    string `yaml:"target_service_address"`
	ExpectedBodyRegex       string `yaml:"expected_body_regex"`
	NegateExpectedBodyRegex bool   `yaml:"negate_expected_body_regex"`
}

type Config struct {
	Liveness  ProbeConfig `yaml:"liveness"`
	Readiness ProbeConfig `yaml:"readiness"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func checkProbe(config ProbeConfig) (bool, error) {
	// In a real implementation, we would check for other probe types here
	if config.Type == "http" {
		resp, err := http.Get(fmt.Sprintf("http://%s:%d%s", config.TargetServiceAddress, config.Port, config.Path))
		if err != nil {
			return false, fmt.Errorf("HTTP request failed: %w", err)
		}
		defer resp.Body.Close()

		if config.ExpectedBodyRegex != "" {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return false, fmt.Errorf("failed to read response body: %w", err)
			}

			re, err := regexp.Compile(config.ExpectedBodyRegex)
			if err != nil {
				// Log the error or handle it appropriately
				return false, fmt.Errorf("invalid regex: %w", err)
			}

			match := re.Match(body)
			if config.NegateExpectedBodyRegex {
				if !match {
					return true, nil
				} else {
					return false, fmt.Errorf("regex matched, but negation was expected")
				}
			}
			if !match {
				return false, fmt.Errorf("regex did not match")
			}
		}

		if resp.StatusCode != config.ExpectedStatus {
			return false, fmt.Errorf("unexpected status code: got %d, expected %d", resp.StatusCode, config.ExpectedStatus)
		}
		return true, nil
	}
	return false, fmt.Errorf("unsupported probe type: %s", config.Type)
}

func main() {
	config, err := loadConfig("config/config.yaml")
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		ok, err := checkProbe(config.Liveness)
		if ok {
			fmt.Fprintln(w, "OK")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, "NOT OK:", err)
		}
	})

	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ok, err := checkProbe(config.Readiness)
		if ok {
			fmt.Fprintln(w, "OK")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, "NOT OK:", err)
		}
	})

	fmt.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", nil)
}
