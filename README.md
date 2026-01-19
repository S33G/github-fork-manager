<div align="center">

### GitHub Fork Manager · TUI cleanup for your forks (and repos)

[![Release](https://img.shields.io/github/v/tag/seeg/github-fork-manager?logo=github)](https://github.com/seeg/github-fork-manager/releases)
[![Build](https://img.shields.io/github/actions/workflow/status/seeg/github-fork-manager/release.yml?branch=main&logo=githubactions&label=release)](https://github.com/seeg/github-fork-manager/actions/workflows/release.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](#dev)

**Terminal-native, fast, and deliberate.** Multi-select, confirm, delete — with logs and hyperlinks you can Ctrl/Cmd-click.

<br/>

> ⚠️ **Danger zone:** This tool deletes repositories you select. Double-check selections and scopes on your `GITHUB_TOKEN`.

</div>

---

## Features at a glance
- 🔍 Fuzzy-ish filter by name/owner/language with live narrowing.
- ✅ Multi-select with space/a; batch delete with inline progress + logging to `~/.github-fork-manager/actions.log`.
- 🔗 Clickable repo names (hyperlinks) to open in your terminal.
- 🌐 GitHub.com or custom API base (GHE).
- 🔄 `--non-forks` mode to manage your owned repos too.

## Install quickly
```bash
curl -fsSL https://raw.githubusercontent.com/seeg/github-fork-manager/main/scripts/install.sh | bash
# optional: VERSION=v0.1.0 INSTALL_DIR=/usr/local/bin bash install.sh
```
Manual:
```bash
chmod +x github-fork-manager-linux-amd64
mv github-fork-manager-linux-amd64 ~/.local/bin/github-fork-manager
```

## Configure once
- Export `GITHUB_TOKEN` (classic/PAT with `delete_repo` + `repo`).
- Optional config file `~/.github-fork-manager/config.json`:
  ```json
  { "token": "ghp_xxx", "api_base": "https://api.github.com", "log_path": "~/.github-fork-manager/actions.log" }
  ```
- Helper: `./scripts/setup-config.sh` prompts and writes the file.

## Run
```bash
github-fork-manager          # forks view
github-fork-manager --non-forks  # manage owned repos
```
From source:
```bash
go run ./cmd/github-fork-manager
# or
make build && ./github-fork-manager
```

## Keys you’ll use
- `j/k` or arrows: move
- `space`: toggle selection
- `a`: select/deselect all visible
- `/`: filter (Enter apply, Esc clear)
- `d`: delete selected (requires typing `<username> approves`)
- `r`: refresh · `q`/`Ctrl+C`: quit · `?`: help blurb

## Safety + logging
- Confirmation gate: type `<github-username> approves` before deletion runs.
- Sequential deletes to stay gentle on rate limits; inline errors per repo.
- Actions logged to `~/.github-fork-manager/actions.log`.

## Release pipeline
- Tag `v*` → GitHub Actions builds Linux/macOS/Windows binaries + checksums.
- Assets: `github-fork-manager-{os}-{arch}`, `checksums.txt`.

## Dev
- Go 1.22+; `make build`, `make test`, `make build-all`, `make release VERSION=vX.Y.Z`.
- Core code: `cmd/github-fork-manager`, `internal/gh`, `internal/config`.
