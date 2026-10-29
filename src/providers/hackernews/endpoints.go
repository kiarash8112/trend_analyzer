package hackernews

import (
	"net/url"
)

// BaseURL is the Algolia-hosted Hacker News search API (free, no key).
const BaseURL = "https://hn.algolia.com/api/v1"

// SearchPath searches stories, sorted by relevance.
const SearchPath = "/search"

// SearchByURL builds a search for stories whose link contains target (e.g. "github.com/owner/name").
func SearchByURL(target string) string {
	q := url.Values{}
	q.Set("query", target)
	q.Set("restrictSearchableAttributes", "url")
	q.Set("tags", "story")
	q.Set("hitsPerPage", "20")
	return BaseURL + SearchPath + "?" + q.Encode()
}

// ItemURL is the public discussion page of a story.
func ItemURL(id string) string {
	return "https://news.ycombinator.com/item?id=" + url.QueryEscape(id)
}
