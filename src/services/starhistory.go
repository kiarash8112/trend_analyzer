package services

import (
	"sort"

	"trend_analyzer/src/models"
)

// DetectSpike finds the stretch between two sampled points with the fastest star growth.
// Gaps shorter than a day count as one day, so a burst of stars in a few hours doesn't
// outrank sustained growth. Returns nil with fewer than two points.
func DetectSpike(points []models.StarPoint) *models.Spike {
	if len(points) < 2 {
		return nil
	}
	sorted := append([]models.StarPoint(nil), points...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Date.Before(sorted[j].Date) })

	var best *models.Spike
	for i := 1; i < len(sorted); i++ {
		prev, cur := sorted[i-1], sorted[i]
		days := max(cur.Date.Sub(prev.Date).Hours()/24, 1)
		rate := float64(cur.Count-prev.Count) / days
		if best == nil || rate > best.StarsPerDay {
			best = &models.Spike{Start: prev.Date, End: cur.Date, StarsPerDay: rate}
		}
	}
	return best
}
