// Command github_trending prints the repositories on https://github.com/trending.
//
//	go run ./cmd/github_trending -since weekly -language go
//	go run ./cmd/github_trending -json
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"trend_analyzer/src/config"
	"trend_analyzer/src/providers/github"
)

func main() {
	since := flag.String("since", github.SinceDaily, "date range: daily, weekly or monthly")
	language := flag.String("language", "", "programming language filter, e.g. go, python")
	spoken := flag.String("spoken", "", "spoken language code filter, e.g. en")
	asJSON := flag.Bool("json", false, "print JSON instead of a table")
	flag.Parse()

	settings := config.Load()
	provider := github.New(github.Config{
		BaseURL:    settings.GitHubBaseURL,
		UserAgent:  settings.GitHubUserAgent,
		Timeout:    settings.GitHubTimeout,
		MaxRetries: settings.GitHubMaxRetries,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	repos, err := provider.FetchTrending(ctx, github.TrendingQuery{
		Language:       *language,
		Since:          *since,
		SpokenLanguage: *spoken,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(repos); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "#\tREPO\tLANGUAGE\tSTARS\tFORKS\t+STARS (%s)\n", *since)
	for i, r := range repos {
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%d\t%d\n", i+1, r.FullName(), r.Language, r.Stars, r.Forks, r.StarsInPeriod)
	}
	w.Flush()
}
