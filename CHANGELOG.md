# Changelog

All notable changes to gh-planning are documented here.

## [v0.4.0] — 2026-07-14

### Added
- **Field-aware project filters** — `status`, `board`, and `catch-up` now support repeatable `--field Field=Value` filters for custom GitHub Projects fields such as Manager, Owner, Workstream, Target Date, Trending, and Status. The special value `me` resolves to the current GitHub login.
- **Profile default filters** — profiles can now store default field filters with `filters` or `filter.<field>`, so repeated commands automatically scope to a user’s operating view.
- **Board hygiene reports** — new `gh planning hygiene` command identifies stale active work, unowned active items, missing planning fields, closed work still active on the board, Done items that remain open, blocked items without linked blockers, stale or unowned child issues, and merged PRs still active.
- **Copilot skill and MCP coverage** for board hygiene and field filters, including the new `planning-hygiene` MCP tool.
- **TUI dashboard expansion** — added the tab-based TUI dashboard, project switching, manual refresh, auto-refresh, issue type badges, dependency data, and a status change action picker.

### Changed
- **Faster project fetches** — project item pagination now uses one `gh api graphql --paginate --slurp` call instead of spawning a separate `gh` process for every page.
- **Richer project field extraction** — project item loading now captures single-select, text, number, date, iteration, and user fields for filtering and JSON output.
- **Project display and draft handling** — project display was improved and draft items are skipped.

### Fixed
- **MCP guide workflow docs** — removed stale guide workflow enums from the MCP tool and documentation.
- **Documentation cleanup** — removed references to removed commands and documented previously missing commands.

## [v0.3.0] — 2026-03-23

### Added
- **Profile and project display in root command** — Running `gh planning` now shows the active profile name and project owner/number so you always know which context you're working in. Auto-detected profiles are labeled accordingly.
- **Unit tests** for `parseDuration`, `humanizeDuration`, `projectURL`, `filterProjectItems`, `decorateStatus`, `truncate`, `findStatusOption`, `filterNonGlobRepos`, `kindPrefix`, `RepositoryNameFromURL`, `IssueURL`, `maxTime`.

### Fixed
- **Standup scoped to profile repos** — `gh planning standup` now scopes searches to repos configured in the active profile instead of showing all work across GitHub. Falls back to repos in the project when no profile repos are set.
- **`truncate()` uses runewidth** — emoji and CJK characters now render correctly in status output instead of being mis-measured with `len()`.
- **`output.PrintJSON` accepts `io.Writer`** — all output functions now use Cobra's `cmd.OutOrStdout()` instead of hardcoded `os.Stdout`, enabling proper test isolation and output redirection.
- **Board/status/standup rendering functions accept `io.Writer`** — `printBoardView`, `printSwimlaneBoardView`, `printStatusGroups`, and `printStandupReport` are now testable.
- **Unhandled `w.Flush()` error** in `printStatusGroups` is now checked.
- **gofmt violation** in `totalItems()` indentation fixed.

### Changed
- **Consolidated `splitTeam()` into `splitAndTrim()`** — removed duplicate function, moved canonical version to `helpers.go`.
- **Consolidated two `init()` functions** in `blocked.go` into one.
- **Clarifying comment** added to intentional empty SUCCESS switch case in `review.go`.

## [v0.2.0] — 2026-03-23

### Changed
- Code health refactors (output writer pattern, test coverage, style fixes).

## [v0.1.0] — 2026-03-23

- Initial release.
