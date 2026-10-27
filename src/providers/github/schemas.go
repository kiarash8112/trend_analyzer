package github

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
