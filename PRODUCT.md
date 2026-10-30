# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

The GitHub community: developers and technically curious people who want to read what is trending on GitHub each day and learn from it. Assume a reader who is intelligent but has never heard of the project, its authors or its ecosystem.

They come to:

- learn how projects get popular (launch channels, timing, framing) and apply it to their own launches;
- find tools worth trying, and understand what they do and who they are for;
- follow which themes and ecosystems are rising over time.

## Product Purpose

Trend Analyzer takes the repositories on GitHub's daily Trending list and explains, for each one, why it is popular: what it does, what problem it solves and for whom, where it launched, how it spread, and why it is trending now. Success is a reader who finishes a report understanding the story behind each repo without prior context, and who can reuse the lessons.

## Positioning

GitHub Trending shows what is popular. Trend Analyzer explains why, as an evidence-backed story per repo: a dated launch, a timeline with a source for each event, a confidence rating, and the list of sources it used.

## Operating Context

- `cmd/analyze_trending` collects evidence (GitHub repo data, star history samples, Hacker News posts), asks Claude to research each repo through the local Claude Code CLI, and writes `reports/<date>-daily.json` and `.md`.
- `cmd/report_server` serves the reports as JSON and the built Angular app (`web/`), which lists reports by date and shows one card per repo with a text filter.
- Reports are produced daily; analyses are cached so a repo that stays trending is not re-researched within the cache window.

## Capabilities and Constraints

- Scope is the daily GitHub Trending list only. Features that do not serve that are removed rather than kept as options.
- Each analysis has: why it got popular (a list of reasons), problem solved, target users, launch channel, URL and date, how it spread, a dated timeline with sources, confidence (high, medium, low), sources, and the analysis time.
- An analysis can fail; the UI must show the failure for that repo and still render the rest of the report.
- Star history needs `GITHUB_TOKEN`; without it the evidence has no star points.
- Undecided: how and where the reports are published to the public (currently a local server only).

## Evidence on Hand

- Real reports in `reports/` (e.g. `2026-10-29-daily.json`, 10 repos; `2026-10-30-daily.json`, 1 repo).
- No users, testimonials, metrics or press exist yet. Do not invent any.

## Product Principles

1. Explain for someone with no context: define terms, name who people and organisations are, never assume the ecosystem.
2. Every claim traces to a source or is marked as uncertain; confidence is shown, not hidden.
3. Lessons over listings: the value is the reusable why, not a copy of the trending list.
4. One job, done well: daily trending analysis, with nothing that does not serve it.
