#!/usr/bin/env bash
set -euo pipefail

CONFIG_DIR="${HOME}/.github-fork-manager"
CONFIG_FILE="${CONFIG_DIR}/config.json"

echo "This will create/update ${CONFIG_FILE}"
mkdir -p "${CONFIG_DIR}"

default_token="${GITHUB_TOKEN:-}"
read -r -p "GitHub token [${default_token:+******}]: " token_input
TOKEN="${token_input:-$default_token}"

read -r -p "GitHub API base [https://api.github.com]: " api_input
API_BASE="${api_input:-https://api.github.com}"

read -r -p "Log path [${CONFIG_DIR}/actions.log]: " log_input
LOG_PATH="${log_input:-${CONFIG_DIR}/actions.log}"

cat > "${CONFIG_FILE}" <<EOF
{
  "token": "${TOKEN}",
  "api_base": "${API_BASE}",
  "log_path": "${LOG_PATH}"
}
EOF

echo "Wrote ${CONFIG_FILE}"
echo "Note: Token will also be read from GITHUB_TOKEN if set at runtime."
