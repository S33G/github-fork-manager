#!/usr/bin/env bash
set -euo pipefail

REPO="S33G/github-fork-manager"
BINARY_NAME="github-fork-manager"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${VERSION:-${1:-}}"

detect_os_arch() {
  local os arch
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  arch=$(uname -m)
  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
  esac
  echo "${os}-${arch}"
}

fetch_latest_tag() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | awk -F'"' '/tag_name/ {print $4; exit}'
}

main() {
  local osarch version tag url ext tmp
  osarch=$(detect_os_arch)
  version="$VERSION"

  if [[ -z "$version" ]]; then
    version=$(fetch_latest_tag)
  fi
  if [[ -z "$version" ]]; then
    echo "Failed to determine version; set VERSION=vX.Y.Z" >&2
    exit 1
  fi

  ext=""
  if [[ "$osarch" == windows-* ]]; then
    ext=".exe"
  fi

  tag="${version}"
  url="https://github.com/${REPO}/releases/download/${tag}/${BINARY_NAME}-${osarch}${ext}"
  tmp="$(mktemp)"

  echo "Installing ${BINARY_NAME} ${version} for ${osarch}"
  echo "Downloading ${url}"
  curl -fL "$url" -o "$tmp"

  mkdir -p "$INSTALL_DIR"
  install "$tmp" "$INSTALL_DIR/${BINARY_NAME}${ext}"
  rm -f "$tmp"

  echo "Installed to $INSTALL_DIR/${BINARY_NAME}${ext}"
  echo "Ensure $INSTALL_DIR is on your PATH"
}

main "$@"
