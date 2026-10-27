package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"trend_analyzer/src/models"
	"trend_analyzer/src/providers/claude"
)

// Claude turns Claude analyses into TrendInsights.
type Claude struct {
	p *claude.Provider
}

// NewClaude wraps a Claude provider.
func NewClaude(p *claude.Provider) *Claude {
	return &Claude{p: p}
}

// Analyze runs the prompt and returns Claude's structured answer.
func (a *Claude) Analyze(ctx context.Context, prompt string) (models.TrendInsight, error) {
	res, err := a.p.Run(ctx, prompt, claude.AnalysisSchema)
	if err != nil {
		return models.TrendInsight{}, err
	}
	return DecodeInsight(res)
}

// DecodeInsight converts a CLI result into a TrendInsight.
func DecodeInsight(res *claude.CLIResult) (models.TrendInsight, error) {
	raw := res.StructuredOutput
	if len(raw) == 0 || string(raw) == "null" {
		raw = json.RawMessage(res.Result) // older CLIs put the JSON in result
	}
	if len(raw) == 0 {
		return models.TrendInsight{}, errors.New("claude returned no analysis")
	}

	var out claude.AnalysisOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		return models.TrendInsight{}, fmt.Errorf("decode claude analysis: %w", err)
	}

	timeline := make([]models.TimelineEvent, 0, len(out.Timeline))
	for _, t := range out.Timeline {
		timeline = append(timeline, models.TimelineEvent{Date: t.Date, Event: t.Event, SourceURL: t.SourceURL})
	}
	return models.TrendInsight{
		WhyPopular:    out.WhyPopular,
		ProblemSolved: out.ProblemSolved,
		TargetUsers:   out.TargetUsers,
		LaunchChannel: out.LaunchChannel,
		LaunchURL:     out.LaunchURL,
		LaunchDate:    out.LaunchDate,
		HowItSpread:   out.HowItSpread,
		Timeline:      timeline,
		Confidence:    out.Confidence,
		Sources:       out.Sources,
		CostUSD:       res.TotalCostUSD,
	}, nil
}
