package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// maxReadmeBytes caps how much of a README is downloaded.
const maxReadmeBytes = 1 << 20

// ErrTokenRequired is returned by calls GitHub only allows with authentication.
var ErrTokenRequired = errors.New("needs a GitHub token (set GITHUB_TOKEN)")

// Config holds the settings the provider needs.
type Config struct {
	BaseURL    string // website, for the trending page
	APIBaseURL string // REST API
	Token      string // optional; required for star history
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

// Provider fetches data from github.com and the GitHub REST API.
type Provider struct {
	cfg  Config
	http *http.Client
}

// New creates a Provider, filling in defaults for empty config values.
func New(cfg Config) *Provider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = DefaultAPIBaseURL
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

// HasToken reports whether authenticated API calls are possible.
func (p *Provider) HasToken() bool {
	return p.cfg.Token != ""
}

// FetchTrending downloads and parses the trending repositories page.
func (p *Provider) FetchTrending(ctx context.Context, q TrendingQuery) ([]TrendingRepo, error) {
	if q.Since == "" {
		q.Since = SinceDaily
	}
	if !ValidSince(q.Since) {
		return nil, fmt.Errorf("invalid since %q: want daily, weekly or monthly", q.Since)
	}

	body, err := p.get(ctx, TrendingURL(p.cfg.BaseURL, q.Language, q.Since, q.SpokenLanguage), "text/html")
	if err != nil {
		return nil, err
	}
	defer body.Close()

	return ParseTrending(body, p.cfg.BaseURL)
}

// FetchRepo returns repository metadata from the REST API.
func (p *Provider) FetchRepo(ctx context.Context, owner, name string) (*RepoResponse, error) {
	var out RepoResponse
	if err := p.getJSON(ctx, RepoURL(p.cfg.APIBaseURL, owner, name), AcceptJSON, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchReadme returns the repository README as plain text.
func (p *Provider) FetchReadme(ctx context.Context, owner, name string) (string, error) {
	body, err := p.get(ctx, ReadmeURL(p.cfg.APIBaseURL, owner, name), AcceptRaw)
	if err != nil {
		return "", err
	}
	defer body.Close()

	b, err := io.ReadAll(io.LimitReader(body, maxReadmeBytes))
	return string(b), err
}

// FetchStargazers returns one page of stargazers with the time each star was given.
func (p *Provider) FetchStargazers(ctx context.Context, owner, name string, page int) ([]Stargazer, error) {
	if !p.HasToken() {
		return nil, ErrTokenRequired
	}
	var out []Stargazer
	if err := p.getJSON(ctx, StargazersURL(p.cfg.APIBaseURL, owner, name, page), AcceptStar, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (p *Provider) getJSON(ctx context.Context, url, accept string, v any) error {
	body, err := p.get(ctx, url, accept)
	if err != nil {
		return err
	}
	defer body.Close()

	if err := json.NewDecoder(body).Decode(v); err != nil {
		return fmt.Errorf("decode %s: %w", url, err)
	}
	return nil
}

// get performs a GET with retries on network errors, 429 and 5xx responses.
func (p *Provider) get(ctx context.Context, url, accept string) (io.ReadCloser, error) {
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
		req.Header.Set("Accept", accept)
		if strings.HasPrefix(url, p.cfg.APIBaseURL) {
			req.Header.Set("X-GitHub-Api-Version", APIVersion)
			if p.cfg.Token != "" {
				req.Header.Set("Authorization", "Bearer "+p.cfg.Token)
			}
		}

		resp, err := p.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("GET %s: %w", url, err)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			return resp.Body, nil
		}
		resp.Body.Close()

		statusErr := &StatusError{
			URL:         url,
			StatusCode:  resp.StatusCode,
			RetryAfter:  parseRetryAfter(resp.Header.Get("Retry-After")),
			RateLimited: resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0",
		}
		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
			return nil, statusErr
		}
		lastErr = statusErr
	}
	return nil, fmt.Errorf("giving up after %d attempts: %w", p.cfg.MaxRetries+1, lastErr)
}

// StatusError is returned for non-200 responses.
type StatusError struct {
	URL         string
	StatusCode  int
	RetryAfter  time.Duration
	RateLimited bool
}

func (e *StatusError) Error() string {
	msg := fmt.Sprintf("GET %s: unexpected status %d", e.URL, e.StatusCode)
	switch {
	case e.RateLimited:
		msg += " (GitHub API rate limit reached; set GITHUB_TOKEN for a higher limit)"
	case e.StatusCode == http.StatusUnauthorized:
		msg += " (check GITHUB_TOKEN)"
	}
	return msg
}

// IsNotFound reports whether err is a 404 from GitHub.
func IsNotFound(err error) bool {
	var se *StatusError
	return errors.As(err, &se) && se.StatusCode == http.StatusNotFound
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
