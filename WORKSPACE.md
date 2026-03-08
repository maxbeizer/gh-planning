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
| td-inspired agent workflow | `agent-context --new-session` as the "run this first" pattern, structured handoffs with done/remaining/decisions/uncertain | 2026-03-08 |
| Progress logging (`log` command) | Agents need to record decisions and blockers *during* work so the next session has context, not just at handoff | 2026-03-08 |
| Project data cache (2min TTL) | 350+ item projects take ~7s to fetch via paginated GraphQL; cache makes repeated commands instant | 2026-03-08 |
| `go-runewidth` for column alignment | Emoji characters break `len()`-based padding; terminal display width must use Unicode-aware measurement | 2026-03-08 |
| User-first, org-fallback for GraphQL | Rather than requiring config to distinguish user vs org projects, try `user()` then fall back to `organization()` | 2026-03-08 |

---

## Current Priorities

1. Merge open PRs (#6–#10): tests, MCP tools, org support, board cmd, CI
2. Release v0.1.0 once PRs land and CI is green
3. Improve board rendering for narrow terminals / large boards
4. Consider interactive TUI mode (bubbletea) for the board

---

## Future Directions

- **Interactive board** — bubbletea-based TUI with keyboard navigation, drag-to-move issues between columns
- **`gh planning sync`** — refresh WORKSPACE.md dynamic sections from live GitHub data
- **Notifications** — surface @mentions, review requests, and CI failures in `catch-up`
- **Multi-project** — commands that span multiple projects (e.g., standup across work + personal)
- **Copilot agent loop** — packaged script/action that runs the full claim→log→complete cycle autonomously
- **Metrics / velocity** — track cycle time, throughput, and blocked-time trends over time

---

## Open Questions

- Should the cache TTL be configurable? (Currently hardcoded at 2 minutes)
- Should `gh planning board` become the default output of `gh planning` (instead of the current summary)?
- How should we handle GitHub's GraphQL rate limits for very large projects (1000+ items)?

---

*Last updated: 2026-03-08*
