package models

import "time"

// TrendInsight is Claude's answer to: why is this repo popular, what does it solve,
// and how did it get popular.
type TrendInsight struct {
	Repo          string          `json:"repo"`
	AnalyzedAt    time.Time       `json:"analyzed_at"`
	WhyPopular    []string        `json:"why_popular"`
	Problem       string          `json:"problem"`
	Solution      string          `json:"solution"`
	ProblemSolved string          `json:"problem_solved,omitempty"` // only in analyses made before Problem and Solution
	TargetUsers   string          `json:"target_users"`
	LaunchChannel string          `json:"launch_channel"`
	LaunchURL     string          `json:"launch_url"`
	LaunchDate    string          `json:"launch_date"`
	HowItSpread   string          `json:"how_it_spread"`
	Timeline      []TimelineEvent `json:"timeline"`
	Confidence    string          `json:"confidence"` // high, medium or low
	Sources       []string        `json:"sources"`
	CostUSD       float64         `json:"cost_usd"` // as reported by Claude Code
}

// TimelineEvent is one step in how a repo gained attention.
type TimelineEvent struct {
	Date      string `json:"date"`
	Event     string `json:"event"`
	SourceURL string `json:"source_url"`
}
