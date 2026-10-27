package mocks

import (
	"os"
	"testing"

	"trend_analyzer/src/providers/github"
)

// Parses a saved copy of github.com/trending so the scraper can be tested offline.
func TestParseTrendingFixture(t *testing.T) {
	f, err := os.Open("../fixtures/github_trending.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	repos, err := github.ParseTrending(f, github.DefaultBaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) == 0 {
		t.Fatal("expected repositories, got none")
	}

	first := repos[0]
	if first.Owner != "Nutlope" || first.Name != "hallmark" {
		t.Errorf("first repo = %s, want Nutlope/hallmark", first.FullName())
	}
	if first.Language != "CSS" || first.Stars != 38335 || first.Forks != 2001 || first.StarsInPeriod != 2859 {
		t.Errorf("unexpected fields: %+v", first)
	}
	if first.Description == "" || len(first.BuiltBy) == 0 {
		t.Errorf("missing description or contributors: %+v", first)
	}
}
