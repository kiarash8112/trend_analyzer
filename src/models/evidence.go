package models

import "time"

// Mention is a post or discussion that links to a repository.
type Mention struct {
	Source        string    `json:"source"` // e.g. "Hacker News"
	Title         string    `json:"title"`
	URL           string    `json:"url"`            // the link that was shared
	DiscussionURL string    `json:"discussion_url"` // the post itself
	Points        int       `json:"points"`
	Comments      int       `json:"comments"`
	PostedAt      time.Time `json:"posted_at"`
}

// Evidence is everything collected about a repo before asking Claude.
type Evidence struct {
	Repo     Repo        `json:"repo"`
	Stars    StarHistory `json:"stars"`
	Mentions []Mention   `json:"mentions,omitempty"`
	Notes    []string    `json:"notes,omitempty"` // sources that were unavailable, and why
}
