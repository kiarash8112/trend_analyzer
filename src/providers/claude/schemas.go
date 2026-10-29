package claude

import "encoding/json"

// CLIResult is the raw JSON printed by `claude -p --output-format json`.
type CLIResult struct {
	Type             string          `json:"type"`
	Subtype          string          `json:"subtype"` // "success" or an error kind
	IsError          bool            `json:"is_error"`
	Result           string          `json:"result"`
	StructuredOutput json.RawMessage `json:"structured_output"` // set when --json-schema is used
	NumTurns         int             `json:"num_turns"`
	DurationMS       int             `json:"duration_ms"`
	TotalCostUSD     float64         `json:"total_cost_usd"`
}

// AnalysisOutput is the structured answer Claude returns for one repository.
type AnalysisOutput struct {
	WhyPopular    []string        `json:"why_popular"`
	ProblemSolved string          `json:"problem_solved"`
	TargetUsers   string          `json:"target_users"`
	LaunchChannel string          `json:"launch_channel"`
	LaunchURL     string          `json:"launch_url"`
	LaunchDate    string          `json:"launch_date"`
	HowItSpread   string          `json:"how_it_spread"`
	Timeline      []TimelineEntry `json:"timeline"`
	Confidence    string          `json:"confidence"`
	Sources       []string        `json:"sources"`
}

// TimelineEntry is one dated event in AnalysisOutput.Timeline.
type TimelineEntry struct {
	Date      string `json:"date"`
	Event     string `json:"event"`
	SourceURL string `json:"source_url"`
}

// AnalysisSchema is the JSON Schema passed to --json-schema; it must match AnalysisOutput.
const AnalysisSchema = `{
  "type": "object",
  "additionalProperties": false,
  "required": ["why_popular", "problem_solved", "target_users", "launch_channel", "launch_url",
               "launch_date", "how_it_spread", "timeline", "confidence", "sources"],
  "properties": {
    "why_popular":    {"type": "array", "items": {"type": "string"},
                       "description": "Concrete reasons it got popular now, one reason per item, without numbering"},
    "problem_solved": {"type": "string", "description": "The problem it solves"},
    "target_users":   {"type": "string", "description": "Who it is for"},
    "launch_channel": {"type": "string", "description": "Where it was first introduced, e.g. Hacker News, X, Reddit, Product Hunt, blog; or unknown"},
    "launch_url":     {"type": "string", "description": "URL of the launch post, or empty if unknown"},
    "launch_date":    {"type": "string", "description": "YYYY-MM-DD, or unknown"},
    "how_it_spread":  {"type": "string", "description": "How attention spread after the launch"},
    "timeline": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["date", "event", "source_url"],
        "properties": {
          "date":       {"type": "string"},
          "event":      {"type": "string"},
          "source_url": {"type": "string"}
        }
      }
    },
    "confidence": {"type": "string", "enum": ["high", "medium", "low"]},
    "sources":    {"type": "array", "items": {"type": "string"}}
  }
}`
