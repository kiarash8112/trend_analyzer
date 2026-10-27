package github

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Config holds the settings the provider needs.
type Config struct {
	BaseURL    string
	UserAgent  string
	Timeout    time.Duration
	MaxRetries int
}

// TrendingQuery filters the trending page. All fields are optional.
type TrendingQuery struct {
	Language       string // e.g. "go", "python"
	Since          string // SinceDaily (default), SinceWeekly, SinceMonthly
	SpokenLanguage string // e.g. "en"
}

// Provider fetches data from github.com.
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
		cfg.Timeout = 15 * time.Second
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	return &Provider{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout}}
}

// FetchTrending downloads and parses the trending repositories page.
func (p *Provider) FetchTrending(ctx context.Context, q TrendingQuery) ([]TrendingRepo, error) {
	if q.Since == "" {
		q.Since = SinceDaily
	}
	if !ValidSince(q.Since) {
		return nil, fmt.Errorf("invalid since %q: want daily, weekly or monthly", q.Since)
	}

	body, err := p.get(ctx, TrendingURL(p.cfg.BaseURL, q.Language, q.Since, q.SpokenLanguage))
	if err != nil {
		return nil, err
	}
	defer body.Close()

	return ParseTrending(body, p.cfg.BaseURL)
}

// get performs a GET with retries on network errors, 429 and 5xx responses.
func (p *Provider) get(ctx context.Context, url string) (io.ReadCloser, error) {
	var lastErr error
	for attempt := 0; attempt <= p.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := sleep(ctx, backoff(attempt, lastErr)); err != nil {
				return nil, err
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", p.cfg.UserAgent)
		req.Header.Set("Accept", "text/html")

		resp, err := p.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("GET %s: %w", url, err)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			return resp.Body, nil
		}
		resp.Body.Close()

		statusErr := &StatusError{URL: url, StatusCode: resp.StatusCode, RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After"))}
		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
			return nil, statusErr
		}
		lastErr = statusErr
	}
	return nil, fmt.Errorf("giving up after %d attempts: %w", p.cfg.MaxRetries+1, lastErr)
}

// StatusError is returned for non-200 responses.
type StatusError struct {
	URL        string
	StatusCode int
	RetryAfter time.Duration
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("GET %s: unexpected status %d", e.URL, e.StatusCode)
}

func backoff(attempt int, lastErr error) time.Duration {
	var se *StatusError
	if errors.As(lastErr, &se) && se.RetryAfter > 0 {
		return se.RetryAfter
	}
	return time.Duration(1<<(attempt-1)) * time.Second // 1s, 2s, 4s, ...
}

func parseRetryAfter(v string) time.Duration {
	if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return 0
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
