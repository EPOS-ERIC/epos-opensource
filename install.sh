#!/usr/bin/env sh
#
# Quick installer for epos-opensource
#
#   curl -fsSL https://raw.githubusercontent.com/epos-eu/epos-opensource/main/install.sh | sh
#
# ------------------------------------------------------------------------------

set -eu

REPO="epos-eu/epos-opensource"
BINARY="epos-opensource"
API="https://api.github.com/repos/${REPO}/releases/latest"

# ── Check current version ──────────────────────────────────────────────────────
echo "⧗ Checking if ${BINARY} is already installed…"
if command -v "${BINARY}" >/dev/null 2>&1; then
  # try two common version flags
  if "${BINARY}" --version >/dev/null 2>&1; then
    CURRENT_VERSION="$("${BINARY}" --version)"
  else
    CURRENT_VERSION="$("${BINARY}" version 2>/dev/null)"
  fi
  echo "ℹ Current version: ${CURRENT_VERSION}"
else
  echo "ℹ ${BINARY} is not currently installed."
fi

# ── Detect OS & ARCH ───────────────────────────────────────────────────────────
uname_s="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${uname_s}" in
  darwin)    OS="darwin"   ;;
  linux)     OS="linux"    ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *) echo "✖ Unsupported OS: ${uname_s}"; exit 1 ;;
esac

uname_m="$(uname -m)"
case "${uname_m}" in
  x86_64|amd64)    ARCH="amd64" ;;
  arm64|aarch64)   ARCH="arm64" ;;
  *) echo "✖ Unsupported architecture: ${uname_m}"; exit 1 ;;
esac

ASSET="${BINARY}-${OS}-${ARCH}"
[ "${OS}" = "windows" ] && ASSET="${ASSET}.exe"

# ── Choose install dir by OS ───────────────────────────────────────────────────
choose_dir() {
  case "${OS}" in
    darwin)
      [ -w "/opt/homebrew/bin" ] && echo "/opt/homebrew/bin" && return
      [ -w "/usr/local/bin" ]    && echo "/usr/local/bin" && return
      echo "$HOME/.local/bin" ;;
    linux)
      [ -w "/usr/local/bin" ] && echo "/usr/local/bin" || echo "$HOME/.local/bin" ;;
    windows)
      echo "$HOME/bin" ;;
  esac
}

INSTALL_DIR="$(choose_dir)"
mkdir -p "${INSTALL_DIR}"

case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *) echo "ℹ ${INSTALL_DIR} isn’t on your PATH – add it to your shell profile." ;;
esac

# ── Fetch latest release info ─────────────────────────────────────────────────
echo "⧗ Fetching latest release…"
API_JSON="$(curl -fsSL "${API}")"

if command -v jq >/dev/null 2>&1; then
  LATEST_TAG="$(printf '%s' "$API_JSON" | jq -r .tag_name)"
  DL_URL="$(printf '%s' "$API_JSON" | jq -r --arg NAME "$ASSET" '.assets[] | select(.name==$NAME) | .browser_download_url')"
else
  LATEST_TAG="$(printf '%s' "$API_JSON" | grep -Eo '"tag_name"[ ]*:[ ]*"[^"]+"' | sed -E 's/.*"([^"]+)".*/\1/')"
  DL_URL="$(printf '%s' "$API_JSON" | grep -Eo "https://[^\"[:space:]]*/${ASSET}" | head -n1)"
fi

echo "ℹ Latest release: ${LATEST_TAG}"
[ -n "${DL_URL}" ] || { echo "✖ Couldn’t find asset ${ASSET} in latest release."; exit 1; }

# ── Download & install ─────────────────────────────────────────────────────────
echo "⇣ Downloading ${ASSET} …"
TMP="$(mktemp)"
curl -# -L "${DL_URL}" -o "${TMP}"
chmod +x "${TMP}"

INSTALL_NAME="${BINARY}"
[ "${OS}" = "windows" ] && INSTALL_NAME="${INSTALL_NAME}.exe"
mv "${TMP}" "${INSTALL_DIR}/${INSTALL_NAME}"

echo "✔ Installed ${INSTALL_NAME} to ${INSTALL_DIR}"
