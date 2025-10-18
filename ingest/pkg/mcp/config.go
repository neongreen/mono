package mcp

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config controls how the MCP client connects to a remote server.
type Config struct {
	// Endpoint is the base URL for the server's SSE transport.
	Endpoint string
	// AuthToken, when set, is sent as a bearer token on every HTTP request.
	AuthToken string
	// Headers are additional HTTP headers to attach to requests.
	Headers map[string]string
	// Timeout bounds network operations. Defaults to 30 seconds.
	Timeout time.Duration
	// Retry configures connection retry behaviour.
	Retry RetryConfig
}

func (c Config) validate() error {
	if c.Endpoint == "" {
		return errors.New("mcp: endpoint is required")
	}
	return nil
}

func (c Config) withDefaults() Config {
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
	c.Retry = c.Retry.withDefaults()
	return c
}

// RetryConfig controls connection retries.
type RetryConfig struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

func (r RetryConfig) withDefaults() RetryConfig {
	if r.MaxAttempts <= 0 {
		r.MaxAttempts = 3
	}
	if r.InitialBackoff <= 0 {
		r.InitialBackoff = 500 * time.Millisecond
	}
	if r.MaxBackoff <= 0 {
		r.MaxBackoff = 5 * time.Second
	}
	return r
}

// ResolveConfig merges overrides with environment variables.
func ResolveConfig(provider string, overrides Config) (Config, error) {
	cfg := overrides

	lowerProvider := strings.ToLower(strings.TrimSpace(provider))
	upperProvider := strings.ToUpper(lowerProvider)

	if cfg.Endpoint == "" {
		if upperProvider != "" {
			cfg.Endpoint = os.Getenv(fmt.Sprintf("INGEST_%s_MCP_ENDPOINT", upperProvider))
		}
		if cfg.Endpoint == "" {
			cfg.Endpoint = os.Getenv("INGEST_MCP_ENDPOINT")
		}
		if cfg.Endpoint == "" && lowerProvider != "" {
			if endpoint, ok := providerDefaultEndpoint(lowerProvider); ok {
				cfg.Endpoint = endpoint
			}
		}
	}

	if cfg.AuthToken == "" {
		if upperProvider != "" {
			cfg.AuthToken = os.Getenv(fmt.Sprintf("INGEST_%s_MCP_TOKEN", upperProvider))
		}
		if cfg.AuthToken == "" {
			cfg.AuthToken = os.Getenv("INGEST_MCP_TOKEN")
		}
		if cfg.AuthToken == "" && lowerProvider == "github" {
			if token := os.Getenv("MISE_GITHUB_TOKEN"); token != "" {
				cfg.AuthToken = token
			}
			if cfg.AuthToken == "" {
				cfg.AuthToken = os.Getenv("GITHUB_TOKEN")
			}
		}
	}

	cfg.Headers = mergeHeaders(cfg.Headers, loadHeadersFromEnv(upperProvider))

	if lowerProvider != "" {
		if _, known := providerPresets[lowerProvider]; !known && cfg.Endpoint == "" {
			return Config{}, fmt.Errorf("unknown provider %q and no endpoint configured", lowerProvider)
		}
	}

	if cfg.Timeout == 0 {
		if v := readDurationEnv("INGEST_MCP_TIMEOUT"); v != 0 {
			cfg.Timeout = v
		}
	}

	if cfg.Retry.MaxAttempts == 0 {
		if v := readIntEnv("INGEST_MCP_RETRY_MAX_ATTEMPTS"); v > 0 {
			cfg.Retry.MaxAttempts = v
		}
	}
	if cfg.Retry.InitialBackoff == 0 {
		if d := readDurationEnv("INGEST_MCP_RETRY_INITIAL_BACKOFF"); d > 0 {
			cfg.Retry.InitialBackoff = d
		}
	}
	if cfg.Retry.MaxBackoff == 0 {
		if d := readDurationEnv("INGEST_MCP_RETRY_MAX_BACKOFF"); d > 0 {
			cfg.Retry.MaxBackoff = d
		}
	}

	cfg = cfg.withDefaults()
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func mergeHeaders(dst map[string]string, src map[string]string) map[string]string {
	if dst == nil {
		dst = make(map[string]string, len(src))
	}
	for k, v := range src {
		if _, exists := dst[k]; !exists {
			dst[k] = v
		}
	}
	return dst
}

func loadHeadersFromEnv(provider string) map[string]string {
	headers := map[string]string{}

	parseHeader := func(value string) {
		for _, segment := range strings.Split(value, ",") {
			segment = strings.TrimSpace(segment)
			if segment == "" {
				continue
			}
			parts := strings.SplitN(segment, "=", 2)
			if len(parts) != 2 {
				continue
			}
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	if provider != "" {
		parseHeader(os.Getenv(fmt.Sprintf("INGEST_%s_MCP_HEADERS", provider)))
	}
	parseHeader(os.Getenv("INGEST_MCP_HEADERS"))

	return headers
}

func readDurationEnv(name string) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return 0
	}
	dur, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}
	return dur
}

func readIntEnv(name string) int {
	value := os.Getenv(name)
	if value == "" {
		return 0
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return n
}
