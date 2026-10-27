// Command analyze_trending asks Claude why each repository on GitHub Trending is
// popular, what problem it solves, and how it got popular. It runs Claude through
// the local Claude Code CLI, so it uses your Claude subscription.
//
//	go run ./cmd/analyze_trending
//	go run ./cmd/analyze_trending -since weekly -language go -limit 5
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"trend_analyzer/src/adapters"
	"trend_analyzer/src/cache"
	"trend_analyzer/src/config"
	"trend_analyzer/src/providers/claude"
	"trend_analyzer/src/providers/github"
	"trend_analyzer/src/providers/hackernews"
	"trend_analyzer/src/services"
)

func main() {
	since := flag.String("since", github.SinceDaily, "date range: daily, weekly or monthly")
	language := flag.String("language", "", "programming language filter, e.g. go, python")
	limit := flag.Int("limit", 10, "analyze at most this many repos (0 = all)")
	concurrency := flag.Int("concurrency", 2, "repos analyzed in parallel")
	maxAge := flag.Duration("max-age", 72*time.Hour, "reuse cached analyses younger than this (0 = always re-analyze)")
	starSamples := flag.Int("star-samples", 10, "stargazer pages to sample per repo (needs GITHUB_TOKEN; 0 = skip)")
	model := flag.String("model", "", "Claude model alias or id (default: CLAUDE_MODEL or your account default)")
	flag.Parse()

	settings := config.Load()
	if *model == "" {
		*model = settings.ClaudeModel
	}
	logf := func(format string, args ...any) {
		fmt.Fprintf(os.Stderr, time.Now().Format("15:04:05")+" "+format+"\n", args...)
	}

	gh := github.New(github.Config{
		BaseURL:    settings.GitHubBaseURL,
		APIBaseURL: settings.GitHubAPIBaseURL,
		Token:      settings.GitHubToken,
		UserAgent:  settings.GitHubUserAgent,
		Timeout:    settings.GitHubTimeout,
		MaxRetries: settings.GitHubMaxRetries,
	})
	if !gh.HasToken() && *starSamples > 0 {
		logf("GITHUB_TOKEN not set: skipping star history (GitHub API limit is also only 60 requests/hour)")
		*starSamples = 0
	}

	analyzer := &services.Analyzer{
		GitHub: adapters.NewGitHub(gh),
		HackerNews: adapters.NewHackerNews(hackernews.New(hackernews.Config{
			BaseURL:   settings.HackerNewsBaseURL,
			UserAgent: settings.GitHubUserAgent,
		})),
		Claude: adapters.NewClaude(claude.New(claude.Config{
			Bin:     settings.ClaudeBin,
			Model:   *model,
			Timeout: settings.ClaudeTimeout,
		})),
		Cache:       cache.New(settings.CacheDir),
		MaxAge:      *maxAge,
		StarSamples: *starSamples,
		Concurrency: *concurrency,
		Logf:        logf,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	results, err := analyzer.AnalyzeTrending(ctx, *language, *since, *limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	now := time.Now()
	base := filepath.Join(settings.ReportsDir, now.Format("2006-01-02")+"-"+*since)
	if *language != "" {
		base += "-" + *language
	}
	if err := writeReports(base, results, *since, now); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	failed := 0
	for _, r := range results {
		if r.Insight == nil {
			failed++
		}
	}
	fmt.Printf("Analyzed %d repos (%d failed).\nReport: %s.md\nData:   %s.json\n", len(results), failed, base, base)
	if failed > 0 {
		os.Exit(1)
	}
}

func writeReports(base string, results []services.Result, since string, now time.Time) error {
	if err := os.MkdirAll(filepath.Dir(base), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(base+".md", []byte(services.RenderMarkdown(results, since, now)), 0o644); err != nil {
		return err
	}
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(base+".json", data, 0o644)
}
