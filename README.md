## GitHub Fork Manager (TUI)

> **Warning:** This tool deletes repositories you select. Double-check selections and ensure your `GITHUB_TOKEN` has the intended permissions before proceeding.

Fast TUI for listing your GitHub forks and deleting many at once.

### Features
- Lists all forks owned by the authenticated user with paging handled automatically.
- Quick filtering by name/owner/language, multi-select, and batch delete with progress.
- Safe defaults: explicit selection, inline errors, and an action log at `~/.github-fork-manager/actions.log`.
- Works against github.com or a custom API base (GitHub Enterprise).

### Install
1) Download a release binary from GitHub Releases once the workflow runs on a tag (files: `github-fork-manager-{os}-{arch}`).
2) Make it executable and place on your PATH, e.g.:
   ```bash
   chmod +x github-fork-manager-linux-amd64
   mv github-fork-manager-linux-amd64 ~/.local/bin/github-fork-manager
   ```
3) Provide a token with `GITHUB_TOKEN` (classic or PAT with `delete_repo` + `repo` scopes). Create one here: https://github.com/settings/tokens

### Quick install via curl
```bash
curl -fsSL https://raw.githubusercontent.com/seeg/github-fork-manager/main/scripts/install.sh | bash
# optional: VERSION=v0.1.0 INSTALL_DIR=/usr/local/bin bash install.sh
```

### Quick setup script (config)
- Run `./scripts/setup-config.sh` to write `~/.github-fork-manager/config.json` interactively (token, API base, log path).
- You can also just set `GITHUB_TOKEN` and skip the config file.

### Run from source
```bash
go run ./cmd/github-fork-manager
```
Or use the Makefile helpers (Go 1.22+):
```bash
make build          # local binary
make build-all      # cross-compile to dist/
make install        # install to ~/.local/bin by default
make test           # run tests
```

### Manage non-fork repos
- Default view shows forks you own. To manage normal repos instead, run:
  ```bash
  github-fork-manager --non-forks
  ```

### Configuration
- Env: `GITHUB_TOKEN` (required), `GITHUB_API_BASE` (optional, defaults to `https://api.github.com`).
- Optional config file: `~/.github-fork-manager/config.json`
  ```json
  {
    "token": "ghp_your_token",
    "api_base": "https://api.github.com",
    "log_path": "~/.github-fork-manager/actions.log"
  }
  ```

### Keybindings
- `j/k` or arrows: move
- `space`: toggle selection
- `a`: select/deselect all visible
- `/`: filter (type, Enter to apply, Esc to clear)
- `d`: delete selected forks (sequential, with inline status)
- `r`: refresh list
- `q` or `Ctrl+C`: quit

### Notes on deleting
- Deletions are sequential to stay gentle on the API. Errors per repo are surfaced inline.
- Actions are logged to `~/.github-fork-manager/actions.log`.
- A missing token will be shown in the UI; set `GITHUB_TOKEN` before deleting.

### Release pipeline
- Workflow: `.github/workflows/release.yml`
- Trigger: push a tag matching `v*`.
- Builds: Linux/macOS/Windows (amd64) binaries via `go build`, uploads to the tag release.

### Development
- Go 1.22+
- Update deps: `go mod tidy`
- Lint/test as desired; core logic is in `cmd/github-fork-manager` and `internal/gh`.
