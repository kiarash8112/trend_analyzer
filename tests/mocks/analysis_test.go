package mocks

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"trend_analyzer/src/adapters"
	"trend_analyzer/src/models"
	"trend_analyzer/src/providers/claude"
	"trend_analyzer/src/services"
)

func TestDetectSpike(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }
	points := []models.StarPoint{
		{Date: day(1), Count: 1},
		{Date: day(10), Count: 100},  // ~11/day
		{Date: day(12), Count: 3100}, // 1,500/day: the launch
		{Date: day(20), Count: 4000}, // ~112/day
	}

	spike := services.DetectSpike(points)
	if spike == nil {
		t.Fatal("expected a spike")
	}
	if !spike.Start.Equal(day(10)) || !spike.End.Equal(day(12)) || spike.StarsPerDay != 1500 {
		t.Errorf("spike = %+v, want Oct 10-12 at 1500/day", spike)
	}
	if services.DetectSpike(points[:1]) != nil {
		t.Error("expected no spike from a single point")
	}
}

func TestDecodeInsightFromCLIFixture(t *testing.T) {
	b, err := os.ReadFile("../fixtures/claude_cli_result.json")
	if err != nil {
		t.Fatal(err)
	}
	var res claude.CLIResult
	if err := json.Unmarshal(b, &res); err != nil {
		t.Fatal(err)
	}

	in, err := adapters.DecodeInsight(&res)
	if err != nil {
		t.Fatal(err)
	}
	if len(in.WhyPopular) != 2 || in.LaunchChannel != "Hacker News" || in.Confidence != "high" || len(in.Timeline) != 2 || len(in.Sources) == 0 {
		t.Errorf("unexpected insight: %+v", in)
	}
	if in.CostUSD == 0 {
		t.Error("expected cost to be copied from the CLI result")
	}
}

func TestBuildPromptIncludesEvidence(t *testing.T) {
	ev := models.Evidence{
		Repo:     models.Repo{Owner: "acme", Name: "rocket", URL: "https://github.com/acme/rocket", Period: "daily", Stars: 5000, StarsInPeriod: 900, Readme: "# Rocket"},
		Mentions: []models.Mention{{Source: "Hacker News", Title: "Show HN: Rocket", Points: 812, DiscussionURL: "https://news.ycombinator.com/item?id=1"}},
		Notes:    []string{"Star history: needs a GitHub token"},
	}
	prompt := services.BuildPrompt(ev)
	for _, want := range []string{"acme/rocket", "+900 today", "Show HN: Rocket", "needs a GitHub token", "<readme>"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt is missing %q", want)
		}
	}
}
