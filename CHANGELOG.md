# Changelog

All notable changes to gh-planning are documented here.

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
