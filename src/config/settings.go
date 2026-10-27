package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// Settings holds configuration read from the environment.
type Settings struct {
	GitHubBaseURL    string
	GitHubAPIBaseURL string
	GitHubToken      string
	GitHubUserAgent  string
	GitHubTimeout    time.Duration
	GitHubMaxRetries int

	HackerNewsBaseURL string

	ClaudeBin     string
	ClaudeModel   string
	ClaudeTimeout time.Duration

	CacheDir   string
	ReportsDir string
}

// Load reads settings from the environment, after loading a local .env file if present.
// Variables already set in the environment take precedence over .env.
func Load() Settings {
	loadDotEnv(".env")
	return Settings{
		GitHubBaseURL:    getEnv("GITHUB_BASE_URL", "https://github.com"),
		GitHubAPIBaseURL: getEnv("GITHUB_API_BASE_URL", "https://api.github.com"),
		GitHubToken:      os.Getenv("GITHUB_TOKEN"),
		GitHubUserAgent:  getEnv("GITHUB_USER_AGENT", "trend-analyzer/1.0"),
		GitHubTimeout:    time.Duration(getEnvInt("GITHUB_TIMEOUT_SECONDS", 15)) * time.Second,
		GitHubMaxRetries: getEnvInt("GITHUB_MAX_RETRIES", 3),

		HackerNewsBaseURL: getEnv("HACKERNEWS_BASE_URL", "https://hn.algolia.com/api/v1"),

		ClaudeBin:     getEnv("CLAUDE_BIN", "claude"),
		ClaudeModel:   os.Getenv("CLAUDE_MODEL"),
		ClaudeTimeout: time.Duration(getEnvInt("CLAUDE_TIMEOUT_SECONDS", 600)) * time.Second,

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

func getEnvInt(key string, fallback int) int {
	if n, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return n
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
