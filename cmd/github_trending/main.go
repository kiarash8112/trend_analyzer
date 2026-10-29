// Command github_trending prints the repositories on https://github.com/trending.
//
//	go run ./cmd/github_trending -since weekly
package main

import (
	"context"
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
	flag.Parse()

	provider := github.New(config.Load().GitHubToken)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	repos, err := provider.FetchTrending(ctx, *since)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "#\tREPO\tLANGUAGE\tSTARS\tFORKS\t+STARS (%s)\n", *since)
	for i, r := range repos {
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%d\t%d\n", i+1, r.FullName(), r.Language, r.Stars, r.Forks, r.StarsInPeriod)
	}
	w.Flush()
}
