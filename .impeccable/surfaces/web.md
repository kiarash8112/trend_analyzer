---
version: 1
slug: "web"
primary_target: "web"
related_targets: []
---

---
version: 1
slug: "web"
primary_target: "web"
related_targets: []
---

# Report reader (web/)

Scope: the daily report page and its date navigation. Mode: Read.

Audience and job: GitHub community readers with no context skim today's trending repos, then open one to learn why it is trending. The "why it is trending" reasons are always visible without a click (one-line lead in the row; the full list in the open row).

Constraints: Angular, data from `/api/reports` and `/api/reports/{name}`; a failed analysis shows its error in its row and never breaks the page; filter by text stays.

Memorable moment: opening a row draws that repo's route, launch to today, spaced by real dates.

Navigation (chosen 2026-10-30, surface round seed key 6f9536ff): list + slides. The timetable is the overview; opening a repo shows it as a slide (route + top three reasons on one screen, the rest below) with a segmented progress bar, previous/next arrows, ← → keys and Esc back to the list. Sketch: .impeccable/mocks/pager/slide-stories.html.

Phone timeline (chosen 2026-11-01): strip map. The route stays horizontal and to scale, pinned while its event is on screen, and shows one event at a time. Earlier and Later move through events; the slide's round arrows change repo. Sketch: .impeccable/mocks/timeline-mobile/compare.html.

## Direction contract

THESIS: Today's trending repos as a departure timetable; each repo is a numbered line and its spread is a metro route. Refuses the card grid and the GitHub-trending list.

OWN-WORLD: Cobalt paper. Cool near-white ground, blue-grey ink and hairline rules, one cobalt signal; Familjen Grotesk heads, IBM Plex Sans body, JetBrains Mono for dates and counts; keycap date tabs; outlined line roundels, filled cobalt when selected; cobalt metro route with 45° steps only at real phase changes.

STORY: The reader sees today's lines, scans the one-line why per row, opens a line, follows its route from launch to now, and reads every reason on its own line with sources.

FIRST VIEWPORT: Top bar with title and keycap date tabs (today filled). Timetable header (Line, Repo, Stars today, Via, Why), row 1 open on a cobalt wash: route panel full width, reasons one per line below, then rows 2 and 3.

FORM: Midnight Timetable · Cobalt paper, a user-steered fusion of the Departure Timetable (grounded) and the Midnight Transit challenger, chosen after re-rolls. seed:fe77a4ccfe460ee5

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance
