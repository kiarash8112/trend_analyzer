package hackernews

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTP client settings.
const (
	userAgent = "trend-analyzer/1.0"
	timeout   = 10 * time.Second
)

// Provider searches Hacker News.
type Provider struct {
	http *http.Client
}

// New creates a Provider.
func New() *Provider {
	return &Provider{http: &http.Client{Timeout: timeout}}
}

// SearchStoriesByURL returns stories that link to a URL containing target.
func (p *Provider) SearchStoriesByURL(ctx context.Context, target string) ([]Hit, error) {
	url := SearchByURL(target)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
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
