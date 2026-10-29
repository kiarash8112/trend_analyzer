package github

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// ParseTrending extracts repositories from a github.com/trending HTML page.
func ParseTrending(r io.Reader) ([]TrendingRepo, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse trending html: %w", err)
	}

	var repos []TrendingRepo
	for _, row := range findAll(doc, func(n *html.Node) bool {
		return n.Data == "article" && hasClass(n, "Box-row")
	}) {
		repo, ok := parseRow(row)
		if ok {
			repos = append(repos, repo)
		}
	}
	return repos, nil
}

func parseRow(row *html.Node) (TrendingRepo, bool) {
	var repo TrendingRepo

	// Repo link: <h2><a href="/owner/name">
	h2 := findFirst(row, func(n *html.Node) bool { return n.Data == "h2" })
	if h2 == nil {
		return repo, false
	}
	link := findFirst(h2, func(n *html.Node) bool { return n.Data == "a" && attr(n, "href") != "" })
	if link == nil {
		return repo, false
	}
	parts := strings.Split(strings.Trim(attr(link, "href"), "/"), "/")
	if len(parts) != 2 {
		return repo, false
	}
	repo.Owner, repo.Name = parts[0], parts[1]
	repo.URL = BaseURL + "/" + repo.Owner + "/" + repo.Name

	if p := findFirst(row, func(n *html.Node) bool { return n.Data == "p" }); p != nil {
		repo.Description = text(p)
	}
	if lang := findFirst(row, func(n *html.Node) bool { return attr(n, "itemprop") == "programmingLanguage" }); lang != nil {
		repo.Language = text(lang)
	}

	for _, a := range findAll(row, func(n *html.Node) bool { return n.Data == "a" }) {
		href := attr(a, "href")
		switch {
		case strings.HasSuffix(href, "/stargazers"):
			repo.Stars = parseCount(text(a))
		case strings.HasSuffix(href, "/forks"):
			repo.Forks = parseCount(text(a))
		}
	}

	// "1,234 stars today/this week/this month"
	if period := findFirst(row, func(n *html.Node) bool { return hasClass(n, "float-sm-right") }); period != nil {
		repo.StarsInPeriod = parseCount(text(period))
	}

	// Contributor avatars: <img alt="@login">
	for _, img := range findAll(row, func(n *html.Node) bool { return n.Data == "img" && hasClass(n, "avatar") }) {
		if login := strings.TrimPrefix(attr(img, "alt"), "@"); login != "" {
			repo.BuiltBy = append(repo.BuiltBy, login)
		}
	}

	return repo, true
}

// parseCount turns strings like "38,335" or "2,859 stars today" into an int.
func parseCount(s string) int {
	var digits strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		} else if r != ',' && digits.Len() > 0 {
			break
		}
	}
	n, _ := strconv.Atoi(digits.String())
	return n
}

func findAll(root *html.Node, match func(*html.Node) bool) []*html.Node {
	var out []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && match(n) {
			out = append(out, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return out
}

func findFirst(root *html.Node, match func(*html.Node) bool) *html.Node {
	if found := findAll(root, match); len(found) > 0 {
		return found[0]
	}
	return nil
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, class string) bool {
	for _, c := range strings.Fields(attr(n, "class")) {
		if c == class {
			return true
		}
	}
	return false
}

// text returns the node's visible text with whitespace collapsed.
func text(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			b.WriteByte(' ')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}
