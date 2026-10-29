package github

import (
	"fmt"
	"net/url"
)

// BaseURL is the public GitHub website root (the trending page is HTML, not the REST API).
const BaseURL = "https://github.com"

// APIBaseURL is the GitHub REST API root.
const APIBaseURL = "https://api.github.com"

// TrendingPath is the route of the trending repositories page.
// Full form: /trending?since=<daily|weekly|monthly>
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

// TrendingURL builds the trending page URL for a date range.
func TrendingURL(since string) string {
	return BaseURL + TrendingPath + "?since=" + url.QueryEscape(since)
}

// RepoURL is GET /repos/{owner}/{name}.
func RepoURL(owner, name string) string {
	return APIBaseURL + "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(name)
}

// ReadmeURL is GET /repos/{owner}/{name}/readme.
func ReadmeURL(owner, name string) string {
	return RepoURL(owner, name) + "/readme"
}

// StargazersURL is GET /repos/{owner}/{name}/stargazers for one page.
func StargazersURL(owner, name string, page int) string {
	return fmt.Sprintf("%s/stargazers?per_page=%d&page=%d", RepoURL(owner, name), StargazersPerPage, page)
}
