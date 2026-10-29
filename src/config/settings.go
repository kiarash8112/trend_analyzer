package config

import (
	"bufio"
	"os"
	"strings"
)

// Settings holds configuration read from the environment.
type Settings struct {
	GitHubToken string
	ClaudeModel string

	CacheDir   string
	ReportsDir string
}

// Load reads settings from the environment, after loading a local .env file if present.
// Variables already set in the environment take precedence over .env.
func Load() Settings {
	loadDotEnv(".env")
	return Settings{
		GitHubToken: os.Getenv("GITHUB_TOKEN"),
		ClaudeModel: os.Getenv("CLAUDE_MODEL"),

		CacheDir:   getEnv("CACHE_DIR", ".cache/insights"),
		ReportsDir: getEnv("REPORTS_DIR", "reports"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, strings.Trim(strings.TrimSpace(val), `"'`))
		}
	}
}
