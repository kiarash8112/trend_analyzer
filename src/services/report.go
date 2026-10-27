package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// RenderMarkdown writes a readable report of the analysis results.
func RenderMarkdown(results []Result, since string, now time.Time) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# GitHub Trending analysis: %s (%s)\n\n", now.Format("2006-01-02"), since)
	fmt.Fprintf(&b, "%d repositories, analyzed by Claude from github.com/trending, the GitHub API, Hacker News and web search.\n", len(results))

	for i, res := range results {
		r := res.Evidence.Repo
		fmt.Fprintf(&b, "\n---\n\n## %d. [%s](%s)\n\n", i+1, r.FullName(), r.URL)

		stats := []string{
			fmt.Sprintf("⭐ %s total", commas(r.Stars)),
			fmt.Sprintf("+%s %s", commas(r.StarsInPeriod), periodLabel(r.Period)),
		}
		if r.Language != "" {
			stats = append(stats, r.Language)
		}
		if !r.CreatedAt.IsZero() {
			stats = append(stats, "created "+r.CreatedAt.Format("2006-01-02"))
		}
		b.WriteString(strings.Join(stats, " · ") + "\n")
		if r.Description != "" {
			fmt.Fprintf(&b, "\n> %s\n", r.Description)
		}

		in := res.Insight
		if in == nil {
			fmt.Fprintf(&b, "\n**Analysis failed:** %s\n", res.Err)
			continue
		}

		fmt.Fprintf(&b, "\n**Why it got popular:** %s\n", in.WhyPopular)
		fmt.Fprintf(&b, "\n**Problem it solves:** %s\n", in.ProblemSolved)
		if in.TargetUsers != "" {
			fmt.Fprintf(&b, "\n**Who it's for:** %s\n", in.TargetUsers)
		}

		launch := in.LaunchChannel
		if in.LaunchDate != "" && in.LaunchDate != "unknown" {
			launch += " (" + in.LaunchDate + ")"
		}
		if in.LaunchURL != "" {
			launch += ": " + in.LaunchURL
		}
		fmt.Fprintf(&b, "\n**How it got popular:** Introduced on %s. %s\n", launch, in.HowItSpread)

		if len(in.Timeline) > 0 {
			b.WriteString("\n**Timeline**\n")
			for _, t := range in.Timeline {
				line := fmt.Sprintf("- %s: %s", t.Date, t.Event)
				if t.SourceURL != "" {
					line += fmt.Sprintf(" ([source](%s))", t.SourceURL)
				}
				b.WriteString(line + "\n")
			}
		}
		if len(in.Sources) > 0 {
			b.WriteString("\n**Sources**\n")
			for _, s := range in.Sources {
				fmt.Fprintf(&b, "- %s\n", s)
			}
		}

		meta := fmt.Sprintf("Confidence: **%s** · analyzed %s", in.Confidence, in.AnalyzedAt.Format("2006-01-02 15:04 MST"))
		if res.Cached {
			meta += " (cached)"
		}
		fmt.Fprintf(&b, "\n_%s_\n", meta)
	}
	return b.String()
}

// commas formats 38335 as "38,335".
func commas(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
