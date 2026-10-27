package github

import "time"

// RepoResponse is the raw shape of GET /repos/{owner}/{name} (fields we use).
type RepoResponse struct {
	FullName        string    `json:"full_name"`
	Description     *string   `json:"description"`
	Homepage        *string   `json:"homepage"`
	Language        *string   `json:"language"`
	Topics          []string  `json:"topics"`
	StargazersCount int       `json:"stargazers_count"`
	ForksCount      int       `json:"forks_count"`
	CreatedAt       time.Time `json:"created_at"`
	PushedAt        time.Time `json:"pushed_at"`
	License         *struct {
		SPDXID string `json:"spdx_id"`
	} `json:"license"`
}

// Stargazer is one entry of GET /repos/{owner}/{name}/stargazers with the star media type.
type Stargazer struct {
	StarredAt time.Time `json:"starred_at"`
}

// TrendingRepo is the raw shape of one repository row scraped from github.com/trending.
// Fields mirror what the page shows; converting to internal models is the adapter's job.
type TrendingRepo struct {
	Owner         string   `json:"owner"`
	Name          string   `json:"name"`
	URL           string   `json:"url"`
	Description   string   `json:"description"`
	Language      string   `json:"language"`
	LanguageColor string   `json:"language_color"`
	Stars         int      `json:"stars"`
	Forks         int      `json:"forks"`
	StarsInPeriod int      `json:"stars_in_period"` // e.g. "2,859 stars today"
	BuiltBy       []string `json:"built_by"`        // contributor logins shown on the row
}

// FullName returns "owner/name".
func (r TrendingRepo) FullName() string {
	return r.Owner + "/" + r.Name
}
