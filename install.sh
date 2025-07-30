#!/usr/bin/env bash
#
# Quick installer for epos-opensource with breaking-change guard
#
#   curl -fsSL https://raw.githubusercontent.com/epos-eu/epos-opensource/main/install.sh | bash
#
# ------------------------------------------------------------------------------

set -eu

REPO="epos-eu/epos-opensource"
BINARY="epos-opensource"
API="https://api.github.com/repos/${REPO}/releases/latest"

# ── Utilities ────────────────────────────────────────────────────────────────
# Extract first "x.y.z" triplet from arbitrary text
extract_semver() {
  printf '%s\n' "$1" \
  | sed -n 's/[^0-9]*\([0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\).*/\1/p'
}

# Compare two semvers; echo 0 if $1 and $2 differ only in patch,
# echo 1 when MAJOR or MINOR of $2 > $1 (breaking change)
is_breaking_change() {
  IFS=. read -r aM aN _ <<EOF
$1
EOF
  IFS=. read -r bM bN _ <<EOF
$2
EOF

  [ "$bM" -gt "$aM" ] && { printf 1; return; }
  [ "$bM" -eq "$aM" ] && [ "$bN" -gt "$aN" ] && { printf 1; return; }
  printf 0
}

# ── Check current version ────────────────────────────────────────────────────
echo "⧗ Checking if ${BINARY} is already installed…"
CURRENT_VER=""
if command -v "${BINARY}" >/dev/null 2>&1; then
  if "${BINARY}" --version >/dev/null 2>&1; then
    raw="$("${BINARY}" --version)"
  else
    raw="$("${BINARY}" version 2>/dev/null || true)"
  fi
  CURRENT_VER="$(extract_semver "$raw" || true)"
  if [ -n "$CURRENT_VER" ]; then
    echo "ℹ Current version: ${CURRENT_VER}"
  else
    echo "ℹ Could not parse current version from: $raw"
  fi
else
  echo "ℹ ${BINARY} is not currently installed."
fi

# ── Detect OS & ARCH ─────────────────────────────────────────────────────────
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

# ── Choose install dir by OS ─────────────────────────────────────────────────
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
  *) echo "ℹ ${INSTALL_DIR} isn't on your PATH – add it to your shell profile." ;;
esac

# ── Fetch latest release info ────────────────────────────────────────────────
echo "⧗ Fetching latest release…"
API_JSON="$(curl -fsSL "${API}")"

if command -v jq >/dev/null 2>&1; then
  TAG_RAW="$(printf '%s' "$API_JSON" | jq -r .tag_name)"
  DL_URL="$(printf '%s' "$API_JSON" \
            | jq -r --arg NAME "$ASSET" \
            '.assets[] | select(.name==$NAME) | .browser_download_url')"
else
  TAG_RAW="$(printf '%s' "$API_JSON" \
            | grep -Eo '"tag_name"[ ]*:[ ]*"[^"]+"' \
            | sed -E 's/.*"([^"]+)".*/\1/')"
  DL_URL="$(printf '%s' "$API_JSON" \
            | grep -Eo "https://[^\"[:space:]]*/${ASSET}" \
            | head -n1)"
fi

LATEST_VER="$(extract_semver "$TAG_RAW")"
echo "ℹ Latest release: ${LATEST_VER}"
[ -n "${DL_URL}" ] || { echo "✖ Couldn't find asset ${ASSET} in latest release."; exit 1; }

# ── Breaking-change guard ────────────────────────────────────────────────────
if [ -n "$CURRENT_VER" ] && [ "$(is_breaking_change "$CURRENT_VER" "$LATEST_VER")" -eq 1 ]; then
  echo "⚠️  Breaking change detected:"
  echo "    Your version : ${CURRENT_VER}"
  echo "    New version  : ${LATEST_VER}"
  echo "    Installing will DELETE all existing epos-opensource environments."
  printf "    Proceed anyway? [y/N]: "
  read answer </dev/tty
  case "$answer" in
    [Yy]*) : ;;  
    *) echo "Installation cancelled."; exit 0 ;;
  esac
fi

# ── Download & install ───────────────────────────────────────────────────────
echo "⇣ Downloading ${ASSET} …"
TMP="$(mktemp)"
curl -# -L "${DL_URL}" -o "${TMP}"
chmod +x "${TMP}"

INSTALL_NAME="${BINARY}"
[ "${OS}" = "windows" ] && INSTALL_NAME="${INSTALL_NAME}.exe"
mv "${TMP}" "${INSTALL_DIR}/${INSTALL_NAME}"

echo "✔ Installed ${INSTALL_NAME} to ${INSTALL_DIR}"
