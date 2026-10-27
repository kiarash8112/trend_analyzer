package models

import "time"

// Repo is a trending repository with whatever details we could collect about it.
type Repo struct {
	Owner         string    `json:"owner"`
	Name          string    `json:"name"`
	URL           string    `json:"url"`
	Description   string    `json:"description"`
	Language      string    `json:"language"`
	Topics        []string  `json:"topics,omitempty"`
	Homepage      string    `json:"homepage,omitempty"`
	License       string    `json:"license,omitempty"`
	Stars         int       `json:"stars"`
	Forks         int       `json:"forks"`
	StarsInPeriod int       `json:"stars_in_period"`
	Period        string    `json:"period"` // daily, weekly or monthly
	BuiltBy       []string  `json:"built_by,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitzero"`
	PushedAt      time.Time `json:"pushed_at,omitzero"`
	Readme        string    `json:"-"`
}

// FullName returns "owner/name".
func (r Repo) FullName() string {
	return r.Owner + "/" + r.Name
}

// StarPoint is the cumulative star count at a moment in time.
type StarPoint struct {
	Date  time.Time `json:"date"`
	Count int       `json:"count"`
}

// Spike is the period with the fastest star growth.
type Spike struct {
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	StarsPerDay float64   `json:"stars_per_day"`
}

// StarHistory is a sampled star growth curve.
type StarHistory struct {
	Points    []StarPoint `json:"points,omitempty"`
	Truncated bool        `json:"truncated"` // GitHub only exposes the first 40k stargazers
	Spike     *Spike      `json:"spike,omitempty"`
}
