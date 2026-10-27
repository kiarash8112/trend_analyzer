// Package cache stores Claude analyses on disk so a repo that stays on
// trending for several days isn't re-analyzed every run.
package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"trend_analyzer/src/models"
)

// Store keeps one JSON file per repository.
type Store struct {
	dir string
}

// New returns a Store rooted at dir. The directory is created on first write.
func New(dir string) *Store {
	return &Store{dir: dir}
}

// Get returns the cached insight for repo ("owner/name") if it is younger than maxAge.
func (s *Store) Get(repo string, maxAge time.Duration) (*models.TrendInsight, bool) {
	b, err := os.ReadFile(s.path(repo))
	if err != nil {
		return nil, false
	}
	var in models.TrendInsight
	if err := json.Unmarshal(b, &in); err != nil || time.Since(in.AnalyzedAt) > maxAge {
		return nil, false
	}
	return &in, true
}

// Put saves an insight, replacing any previous one for the same repo.
func (s *Store) Put(in models.TrendInsight) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path(in.Repo) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path(in.Repo))
}

func (s *Store) path(repo string) string {
	return filepath.Join(s.dir, strings.ReplaceAll(repo, "/", "__")+".json")
}
