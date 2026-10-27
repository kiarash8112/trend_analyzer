package github

import (
	"net/url"
	"strings"
)

// DefaultBaseURL is the public GitHub website root (the trending page is HTML, not the REST API).
const DefaultBaseURL = "https://github.com"

// TrendingPath is the route of the trending repositories page.
// Full form: /trending[/<language>]?since=<daily|weekly|monthly>&spoken_language_code=<code>
const TrendingPath = "/trending"

// Since values accepted by the trending page.
const (
	SinceDaily   = "daily"
	SinceWeekly  = "weekly"
	SinceMonthly = "monthly"
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
