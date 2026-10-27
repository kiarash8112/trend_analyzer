package hackernews

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Config holds the settings the provider needs.
type Config struct {
	BaseURL   string
	UserAgent string
	Timeout   time.Duration
}

// Provider searches Hacker News.
type Provider struct {
	cfg  Config
	http *http.Client
}

// New creates a Provider, filling in defaults for empty config values.
func New(cfg Config) *Provider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "trend-analyzer/1.0"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &Provider{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout}}
}

// SearchStoriesByURL returns stories that link to a URL containing target.
func (p *Provider) SearchStoriesByURL(ctx context.Context, target string) ([]Hit, error) {
	url := SearchByURL(p.cfg.BaseURL, target)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", p.cfg.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: unexpected status %d", url, resp.StatusCode)
	}

	var out SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode %s: %w", url, err)
	}
	return out.Hits, nil
}
