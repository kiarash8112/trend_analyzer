package adapters

import (
	"context"
	"sort"

	"trend_analyzer/src/models"
	"trend_analyzer/src/providers/hackernews"
)

// HackerNews turns Hacker News search results into mentions.
type HackerNews struct {
	p *hackernews.Provider
}

// NewHackerNews wraps a Hacker News provider.
func NewHackerNews(p *hackernews.Provider) *HackerNews {
	return &HackerNews{p: p}
}

// Mentions returns stories linking to any of targets, most upvoted first.
func (a *HackerNews) Mentions(ctx context.Context, targets ...string) ([]models.Mention, error) {
	seen := map[string]bool{}
	var out []models.Mention
	for _, target := range targets {
		hits, err := a.p.SearchStoriesByURL(ctx, target)
		if err != nil {
			return out, err
		}
		for _, h := range hits {
			if seen[h.ObjectID] {
				continue
			}
			seen[h.ObjectID] = true
			out = append(out, models.Mention{
				Source:        "Hacker News",
				Title:         h.Title,
				URL:           h.URL,
				DiscussionURL: hackernews.ItemURL(h.ObjectID),
				Points:        h.Points,
				Comments:      h.NumComments,
				PostedAt:      h.CreatedAt,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Points > out[j].Points })
	return out, nil
}
