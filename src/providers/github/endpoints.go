package github

import (
	"fmt"
	"net/url"
	"strings"
)

// DefaultBaseURL is the public GitHub website root (the trending page is HTML, not the REST API).
const DefaultBaseURL = "https://github.com"

// DefaultAPIBaseURL is the GitHub REST API root.
const DefaultAPIBaseURL = "https://api.github.com"

// TrendingPath is the route of the trending repositories page.
// Full form: /trending[/<language>]?since=<daily|weekly|monthly>&spoken_language_code=<code>
const TrendingPath = "/trending"

// Since values accepted by the trending page.
const (
	SinceDaily   = "daily"
	SinceWeekly  = "weekly"
	SinceMonthly = "monthly"
)

// REST API media types.
const (
	AcceptJSON = "application/vnd.github+json"
	AcceptRaw  = "application/vnd.github.raw+json"  // README as plain text
	AcceptStar = "application/vnd.github.star+json" // stargazers with starred_at
)

// APIVersion pins the REST API version.
const APIVersion = "2022-11-28"

// Stargazer listing limits: 100 per page, and GitHub stops paginating after 400 pages.
const (
	StargazersPerPage = 100
	MaxStargazerPages = 400
)

// ValidSince reports whether s is a date range the trending page supports.
func ValidSince(s string) bool {
	return s == SinceDaily || s == SinceWeekly || s == SinceMonthly
}

// TrendingURL builds the trending page URL for the given filters. Empty values are omitted.
func TrendingURL(baseURL, language, since, spokenLanguage string) string {
	path := TrendingPath
	if language != "" {
		path += "/" + url.PathEscape(strings.ToLower(language))
	}

	q := url.Values{}
	if since != "" {
		q.Set("since", since)
	}
	if spokenLanguage != "" {
		q.Set("spoken_language_code", spokenLanguage)
	}

	u := strings.TrimRight(baseURL, "/") + path
	if enc := q.Encode(); enc != "" {
		u += "?" + enc
	}
	return u
}

// RepoURL is GET /repos/{owner}/{name}.
func RepoURL(apiBaseURL, owner, name string) string {
	return strings.TrimRight(apiBaseURL, "/") + "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(name)
}

// ReadmeURL is GET /repos/{owner}/{name}/readme.
func ReadmeURL(apiBaseURL, owner, name string) string {
	return RepoURL(apiBaseURL, owner, name) + "/readme"
}

// StargazersURL is GET /repos/{owner}/{name}/stargazers for one page.
func StargazersURL(apiBaseURL, owner, name string, page int) string {
	return fmt.Sprintf("%s/stargazers?per_page=%d&page=%d", RepoURL(apiBaseURL, owner, name), StargazersPerPage, page)
}
