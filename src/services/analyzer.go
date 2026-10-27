package services

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"trend_analyzer/src/adapters"
	"trend_analyzer/src/cache"
	"trend_analyzer/src/models"
)

// Analyzer explains why trending repositories are popular.
type Analyzer struct {
	GitHub      *adapters.GitHub
	HackerNews  *adapters.HackerNews
	Claude      *adapters.Claude
	Cache       *cache.Store  // optional
	MaxAge      time.Duration // reuse cached analyses younger than this; 0 always re-analyzes
	StarSamples int           // stargazer pages to sample per repo
	Concurrency int           // repos analyzed in parallel
	Logf        func(format string, args ...any)
}

// Result is the outcome for one repository.
type Result struct {
	Evidence models.Evidence      `json:"evidence"`
	Insight  *models.TrendInsight `json:"insight,omitempty"`
	Cached   bool                 `json:"cached"`
	Err      string               `json:"error,omitempty"`
}

// AnalyzeTrending analyzes the first limit repos on the trending page (all if limit <= 0).
func (a *Analyzer) AnalyzeTrending(ctx context.Context, language, since string, limit int) ([]Result, error) {
	repos, err := a.GitHub.Trending(ctx, language, since)
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(repos) > limit {
		repos = repos[:limit]
	}
	a.logf("found %d trending repos", len(repos))

	results := make([]Result, len(repos))
	sem := make(chan struct{}, max(a.Concurrency, 1))
	var wg sync.WaitGroup
	for i, repo := range repos {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i] = a.Analyze(ctx, repo)
		}()
	}
	wg.Wait()
	return results, nil
}

// Analyze gathers evidence for one repo and asks Claude to explain its popularity.
func (a *Analyzer) Analyze(ctx context.Context, repo models.Repo) Result {
	name := repo.FullName()
	if a.Cache != nil && a.MaxAge > 0 {
		if in, ok := a.Cache.Get(name, a.MaxAge); ok {
			a.logf("%s: using cached analysis from %s", name, in.AnalyzedAt.Format(time.DateTime))
			return Result{Evidence: models.Evidence{Repo: repo}, Insight: in, Cached: true}
		}
	}

	ev := a.gather(ctx, repo)

	a.logf("%s: asking Claude...", name)
	start := time.Now()
	in, err := a.Claude.Analyze(ctx, BuildPrompt(ev))
	if err != nil {
		a.logf("%s: analysis failed: %v", name, err)
		return Result{Evidence: ev, Err: err.Error()}
	}
	in.Repo = name
	in.AnalyzedAt = time.Now().UTC()
	a.logf("%s: done in %s", name, time.Since(start).Round(time.Second))

	if a.Cache != nil {
		if err := a.Cache.Put(in); err != nil {
			a.logf("%s: could not cache analysis: %v", name, err)
		}
	}
	return Result{Evidence: ev, Insight: &in}
}

// gather collects evidence from every source. A failing source becomes a note
// for Claude rather than an error, so one blocked API doesn't stop the analysis.
func (a *Analyzer) gather(ctx context.Context, repo models.Repo) models.Evidence {
	ev := models.Evidence{Repo: repo}
	name := repo.FullName()

	if err := a.GitHub.Enrich(ctx, &ev.Repo); err != nil {
		ev.Notes = append(ev.Notes, "GitHub API details: "+err.Error())
	}

	if a.StarSamples > 0 {
		points, truncated, err := a.GitHub.StarSamples(ctx, repo.Owner, repo.Name, ev.Repo.Stars, a.StarSamples)
		if err != nil {
			ev.Notes = append(ev.Notes, "Star history: "+err.Error())
		}
		ev.Stars = models.StarHistory{Points: points, Truncated: truncated, Spike: DetectSpike(points)}
	}

	if a.HackerNews != nil {
		targets := []string{"github.com/" + name}
		if host := homepageTarget(ev.Repo.Homepage); host != "" {
			targets = append(targets, host)
		}
		mentions, err := a.HackerNews.Mentions(ctx, targets...)
		if err != nil {
			ev.Notes = append(ev.Notes, "Hacker News search: "+err.Error())
		}
		ev.Mentions = mentions
	}

	a.logf("%s: evidence ready (%d star points, %d posts, %d notes)", name, len(ev.Stars.Points), len(ev.Mentions), len(ev.Notes))
	return ev
}

// homepageTarget returns the homepage as "host/path" for link searches, skipping GitHub links.
func homepageTarget(homepage string) string {
	u, err := url.Parse(strings.TrimSpace(homepage))
	if err != nil || u.Host == "" || strings.HasSuffix(u.Host, "github.com") || strings.HasSuffix(u.Host, "github.io") {
		return ""
	}
	return strings.TrimPrefix(u.Host, "www.") + strings.TrimRight(u.Path, "/")
}

func (a *Analyzer) logf(format string, args ...any) {
	if a.Logf != nil {
		a.Logf(format, args...)
	}
}
