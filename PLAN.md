TUI Plan: GitHub Fork Manager

Goals
- List all forks owned by the authenticated user quickly, support search/filter, and allow bulk selection.
- Delete many forks in one action via GitHub REST API with clear confirmation and progress feedback.
- Keep operations safe: dry-run mode, per-repo confirmation option, and reversible logging (record what was deleted).

Non-goals
- Managing upstream repos (issues/PRs) or cloning content.
- Supporting GitHub Enterprise without a configured custom API base URL.

Core User Flows
- Authenticate: load token from env/config; prompt once if missing; validate with a lightweight API call.
- List forks: fetch paginated `/user/repos?affiliation=owner&per_page=100` and filter where `fork=true`; cache in memory.
- Browse/search: sort by `pushed_at`/`stargazers_count`; filter by name/owner/language; indicate archived/private.
- Select: multi-select with keyboard (space/enter), toggle all/visible, and mark for deletion queue.
- Review: show selection count; optional detail pane for highlighted repo (size, last push, default branch, upstream URL).
- Delete: confirmation screen summarizing N repos; optional dry-run; execute concurrent deletes with rate limiting; surface successes/failures inline.
- Logs: write action log (timestamp, repo full_name, result) to a local file.

Data/State Model
- AuthToken { value, source }.
- Repo { id, full_name, owner, name, archived, private, fork, size, language, pushed_at, default_branch, parent_full_name?, ssh_url, html_url }.
- UIState { filter_text, sort_key, selection:set<repo_id>, focused_repo_id, page_cursor?, deleting:boolean, delete_results:map<repo_id,status> }.
- Config file (yaml/json) path: `~/.github-fork-manager/config.(yml|json)` with api_base, log_path, concurrency, dry_run default.

TUI Layout & Interaction
- Two-pane layout: left list of repos with badges (archived/private); right detail pane when available.
- Keybindings: `j/k` or arrows to move; `space` to toggle selection; `a` toggle select all visible; `/` filter; `s` cycle sort; `d` delete selected; `r` refresh; `q` quit; `?` help overlay.
- Status bar for counts: total forks, filtered, selected, pending deletes, failures.
- Progress: while deleting, show spinner per repo or aggregate progress bar; errors stay highlighted until acknowledged.

GitHub API Usage
- List: `GET /user/repos?per_page=100&page=N&affiliation=owner` then filter `fork==true`.
- Delete: `DELETE /repos/{owner}/{repo}`; handle 404/403 gracefully (surface and continue); respect abuse-rate limit headers with backoff.
- Optional upstream lookup: `GET /repos/{owner}/{repo}` to display parent info only for focused repo (lazy load).
- Auth: token via `GITHUB_TOKEN` or config; send `Authorization: Bearer <token>` and `Accept: application/vnd.github+json`.

Execution Model
- Language/runtime: TUI-friendly stack (e.g., Go + bubbletea or Python + textual); abstract API via client interface for testing.
- Concurrency: bounded worker pool for delete operations; queue derived from current selection snapshot to avoid UI drift.
- Caching: keep fetched repos in memory; allow manual refresh to re-pull.
- Logging: append-only log file; redact token; include request ids when available.

Error Handling & Safety
- Detect missing/invalid token early; show actionable message.
- For deletes, default to confirmation dialog with count; require typed `DELETE` confirmation when >N repos (configurable).
- Recover from partial failures: continue other deletes; summarize failures with reasons and retry option.
- Respect rate limits: parse `X-RateLimit-Remaining/Reset`; backoff and notify user when pausing.

Testing Strategy
- Unit: API client (list/delete, pagination, error parsing), selection manager, filter/sort helpers.
- Integration: mock GitHub API server; flow tests for list → select → delete with mixed outcomes.
- Snapshot or screenshot tests of key TUI states if framework supports.

Open Questions
- Should we support multiple accounts/profiles in one UI session?
- Maximum default delete batch before requiring typed confirmation?
