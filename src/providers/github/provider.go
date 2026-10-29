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

// HTTP client settings.
const (
	userAgent  = "trend-analyzer/1.0"
	timeout    = 15 * time.Second
	maxRetries = 3
)

// ErrTokenRequired is returned by calls GitHub only allows with authentication.
var ErrTokenRequired = errors.New("needs a GitHub token (set GITHUB_TOKEN)")

// Provider fetches data from github.com and the GitHub REST API.
type Provider struct {
	token string // optional; required for star history
	http  *http.Client
}

// New creates a Provider. token may be empty.
func New(token string) *Provider {
	return &Provider{token: token, http: &http.Client{Timeout: timeout}}
}

// HasToken reports whether authenticated API calls are possible.
func (p *Provider) HasToken() bool {
	return p.token != ""
}

// FetchTrending downloads and parses the trending repositories page for since
// (SinceDaily, SinceWeekly or SinceMonthly; empty means daily).
func (p *Provider) FetchTrending(ctx context.Context, since string) ([]TrendingRepo, error) {
	if since == "" {
		since = SinceDaily
	}
	if !ValidSince(since) {
		return nil, fmt.Errorf("invalid since %q: want daily, weekly or monthly", since)
	}

	body, err := p.get(ctx, TrendingURL(since), "text/html")
	if err != nil {
		return nil, err
	}
	defer body.Close()

	return ParseTrending(body)
}

// FetchRepo returns repository metadata from the REST API.
func (p *Provider) FetchRepo(ctx context.Context, owner, name string) (*RepoResponse, error) {
	var out RepoResponse
	if err := p.getJSON(ctx, RepoURL(owner, name), AcceptJSON, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchReadme returns the repository README as plain text.
func (p *Provider) FetchReadme(ctx context.Context, owner, name string) (string, error) {
	body, err := p.get(ctx, ReadmeURL(owner, name), AcceptRaw)
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
	if err := p.getJSON(ctx, StargazersURL(owner, name, page), AcceptStar, &out); err != nil {
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
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			if err := sleep(ctx, backoff(attempt, lastErr)); err != nil {
				return nil, err
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", accept)
		if strings.HasPrefix(url, APIBaseURL) {
			req.Header.Set("X-GitHub-Api-Version", APIVersion)
			if p.token != "" {
				req.Header.Set("Authorization", "Bearer "+p.token)
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
	return nil, fmt.Errorf("giving up after %d attempts: %w", maxRetries+1, lastErr)
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
