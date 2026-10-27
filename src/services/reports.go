package services

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ReportSummary describes one saved report file.
type ReportSummary struct {
	Name      string `json:"name"` // file name without .json, e.g. 2026-10-27-weekly-go
	Date      string `json:"date"`
	Since     string `json:"since"`
	Language  string `json:"language,omitempty"`
	RepoCount int    `json:"repo_count"`
}

// ErrReportNotFound is returned for unknown or invalid report names.
var ErrReportNotFound = errors.New("report not found")

var reportName = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(daily|weekly|monthly)(?:-([\w.+#-]+))?$`)

// ListReports returns the JSON reports in dir, newest first.
func ListReports(dir string) ([]ReportSummary, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}

	out := []ReportSummary{}
	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".json")
		m := reportName.FindStringSubmatch(name)
		if m == nil {
			continue
		}
		var results []json.RawMessage
		if b, err := os.ReadFile(f); err != nil || json.Unmarshal(b, &results) != nil {
			continue
		}
		out = append(out, ReportSummary{Name: name, Date: m[1], Since: m[2], Language: m[3], RepoCount: len(results)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name > out[j].Name })
	return out, nil
}

// ReadReport returns the raw JSON of the named report.
func ReadReport(dir, name string) ([]byte, error) {
	if !reportName.MatchString(name) {
		return nil, ErrReportNotFound
	}
	b, err := os.ReadFile(filepath.Join(dir, name+".json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrReportNotFound
	}
	return b, err
}
