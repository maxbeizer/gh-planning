# WORKSPACE.md — Decisions & Direction

> Living document of design decisions, current priorities, and where
> gh-planning is headed. For repo structure and coding conventions,
> see `.github/copilot-instructions.md`.

---

## Design Decisions

| Decision | Context | Date |
|----------|---------|------|
| Shell out to `gh` CLI instead of Go API client | Simpler auth, leverages user's existing `gh` setup, avoids token management | Pre-2026 |
| 3-second API delay on every call | Was a naive rate-limit guard — **removed**, replaced with parallelized calls | 2026-03-08 |
| YAML config with named profiles | Users need to switch between personal and work projects without re-configuring | 2026-03-08 |
| td-inspired agent workflow | `agent-context --new-session` as the "run this first" pattern, structured handoffs with done/remaining/decisions/uncertain — **removed for v1**, see below | 2026-03-08 |
| Progress logging (`log` command) | Agents need to record decisions and blockers *during* work so the next session has context, not just at handoff | 2026-03-08 |
| Project data cache (2min TTL) | 350+ item projects take ~7s to fetch via paginated GraphQL; cache makes repeated commands instant | 2026-03-08 |
| `go-runewidth` for column alignment | Emoji characters break `len()`-based padding; terminal display width must use Unicode-aware measurement | 2026-03-08 |
| User-first, org-fallback for GraphQL | Rather than requiring config to distinguish user vs org projects, try `user()` then fall back to `organization()` | 2026-03-08 |

---

## Current Priorities

1. Open-source release (MIT license, slimmed command set)
2. Improve board rendering for narrow terminals / large boards
3. Better onboarding docs / README for new users

---

## Future Directions

- **Interactive board** — bubbletea-based TUI with keyboard navigation, drag-to-move issues between columns
- **`gh planning sync`** — refresh WORKSPACE.md dynamic sections from live GitHub data
- **Notifications** — surface @mentions, review requests, and CI failures in `catch-up`
- **Multi-project** — commands that span multiple projects (e.g., standup across work + personal)
- **Re-introduce agent orchestration** — the removed agent commands (see below) could return as a separate extension or opt-in module once the core is stable

---

## Removed for v1 (Open-Source Prep)

The following commands were removed to reduce surface area for the initial open-source release.
The full code is preserved in [PR #43](https://github.com/maxbeizer/gh-planning/pull/43) (closed, not merged) on the `agent-orchestration` branch.

### Agent orchestration commands
Removed because they represent a separate concern (AI agent coordination) that adds complexity for users who just want project visibility and daily workflow tools.

| Command | What it did |
|---------|------------|
| `agent-context` | Summarize project context for an AI agent to start work |
| `claim` | Assign yourself + move issue to In Progress in one step |
| `complete` | Post structured completion summary + move issue forward |
| `handoff` | Post structured handoff comment (done/remaining/decisions/uncertain) |
| `queue` | Show items ready for agent processing, filtered by label |
| `daemon` | Autonomous loop: poll queue → claim → dispatch work |
| `dashboard` | Interactive TUI combining sprint, board, blockers, and focus |
| `breakdown` | Use AI (GitHub Models) to split issues into sub-issues |
| `estimate` | Add T-shirt size effort estimates to issues |

### Planning power-user commands
Removed because they overlap with existing commands or the GitHub UI, and add learning curve for new users.

| Command | What it did | Why removed |
|---------|------------|-------------|
| `sprint` | Sprint overview and progress | Overlaps with `status` and `board` |
| `roadmap` | Vertical timeline of project activity | Niche; most teams don't need CLI roadmaps |
| `critical-path` | Show blocking dependency chains | Only useful with heavy `blocked` usage |
| `prioritize` | Interactive backlog reordering | Complex; GitHub UI drag-and-drop is easier |

---

## Open Questions

- Should the cache TTL be configurable? (Currently hardcoded at 2 minutes)
- Should `gh planning board` become the default output of `gh planning` (instead of the current summary)?
- How should we handle GitHub's GraphQL rate limits for very large projects (1000+ items)?

---

*Last updated: 2026-03-19*
