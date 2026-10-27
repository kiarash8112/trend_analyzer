package hackernews

import "time"

// SearchResponse is the raw response of GET /search.
type SearchResponse struct {
	Hits []Hit `json:"hits"`
}

// Hit is one story in a search response.
type Hit struct {
	ObjectID    string    `json:"objectID"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Author      string    `json:"author"`
	Points      int       `json:"points"`
	NumComments int       `json:"num_comments"`
	CreatedAt   time.Time `json:"created_at"`
}
