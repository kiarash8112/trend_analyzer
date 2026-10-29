package adapters

import (
	"context"
	"math"

	"trend_analyzer/src/models"
	"trend_analyzer/src/providers/github"
)

// GitHub turns GitHub provider data into internal models.
type GitHub struct {
	p *github.Provider
}

// NewGitHub wraps a GitHub provider.
func NewGitHub(p *github.Provider) *GitHub {
	return &GitHub{p: p}
}

// Trending returns the repositories on the trending page.
func (a *GitHub) Trending(ctx context.Context, since string) ([]models.Repo, error) {
	raw, err := a.p.FetchTrending(ctx, since)
	if err != nil {
		return nil, err
	}
	if since == "" {
		since = github.SinceDaily
	}

	repos := make([]models.Repo, 0, len(raw))
	for _, r := range raw {
		repos = append(repos, models.Repo{
			Owner:         r.Owner,
			Name:          r.Name,
			URL:           r.URL,
			Description:   r.Description,
			Language:      r.Language,
			Stars:         r.Stars,
			Forks:         r.Forks,
			StarsInPeriod: r.StarsInPeriod,
			Period:        since,
			BuiltBy:       r.BuiltBy,
		})
	}
	return repos, nil
}

// Enrich fills in metadata and the README from the REST API.
func (a *GitHub) Enrich(ctx context.Context, repo *models.Repo) error {
	info, err := a.p.FetchRepo(ctx, repo.Owner, repo.Name)
	if err != nil {
		return err
	}
	if repo.Description == "" && info.Description != nil {
		repo.Description = *info.Description
	}
	if info.Homepage != nil {
		repo.Homepage = *info.Homepage
	}
	if repo.Language == "" && info.Language != nil {
		repo.Language = *info.Language
	}
	if info.License != nil && info.License.SPDXID != "NOASSERTION" {
		repo.License = info.License.SPDXID
	}
	repo.Topics = info.Topics
	repo.Stars = info.StargazersCount
	repo.Forks = info.ForksCount
	repo.CreatedAt = info.CreatedAt
	repo.PushedAt = info.PushedAt

	readme, err := a.p.FetchReadme(ctx, repo.Owner, repo.Name)
	if err != nil && !github.IsNotFound(err) {
		return err
	}
	repo.Readme = readme
	return nil
}

// StarSamples samples the star growth curve by reading a spread of stargazer pages.
// Each sampled page gives two points: its first and last star. truncated is true when
// the repo has more stars than GitHub lets us page through.
func (a *GitHub) StarSamples(ctx context.Context, owner, name string, totalStars, samples int) (points []models.StarPoint, truncated bool, err error) {
	pages := (totalStars + github.StargazersPerPage - 1) / github.StargazersPerPage
	if pages > github.MaxStargazerPages {
		pages, truncated = github.MaxStargazerPages, true
	}

	for _, page := range samplePages(pages, samples) {
		gazers, err := a.p.FetchStargazers(ctx, owner, name, page)
		if err != nil {
			return points, truncated, err
		}
		if len(gazers) == 0 {
			continue
		}
		base := (page - 1) * github.StargazersPerPage
		points = append(points, models.StarPoint{Date: gazers[0].StarredAt, Count: base + 1})
		if last := len(gazers) - 1; last > 0 {
			points = append(points, models.StarPoint{Date: gazers[last].StarredAt, Count: base + len(gazers)})
		}
	}
	return points, truncated, nil
}

// samplePages picks up to n page numbers spread evenly over 1..total, always
// including the first and last page.
func samplePages(total, n int) []int {
	if total <= 0 {
		return nil
	}
	n = max(n, 2)
	if total <= n {
		pages := make([]int, total)
		for i := range pages {
			pages[i] = i + 1
		}
		return pages
	}

	step := float64(total-1) / float64(n-1)
	pages := make([]int, 0, n)
	for i := range n {
		page := 1 + int(math.Round(float64(i)*step))
		if len(pages) == 0 || pages[len(pages)-1] != page {
			pages = append(pages, page)
		}
	}
	return pages
}
