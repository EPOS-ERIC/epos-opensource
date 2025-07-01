#!/usr/bin/env bash
#
# Quick installer for epos-opensource
#
#   curl -fsSL https://raw.githubusercontent.com/<owner>/<repo>/main/install.sh | bash
#
# ------------------------------------------------------------------------------

set -euo pipefail

REPO="epos-eu/epos-opensource"
BINARY="epos-opensource"
API="https://api.github.com/repos/${REPO}/releases/latest"

# ── Detect OS & ARCH ───────────────────────────────────────────────────────────
uname_s="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$uname_s" in
  darwin)   OS="darwin"   ;;
  linux)    OS="linux"    ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *) echo "✖ Unsupported OS: $uname_s"; exit 1 ;;
esac

arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "✖ Unsupported architecture: $arch"; exit 1 ;;
esac

ASSET="${BINARY}-${OS}-${ARCH}"
[ "$OS" = "windows" ] && ASSET="${ASSET}.exe"

# ── Choose install dir by OS ───────────────────────────────────────────────────
choose_dir() {
  case "$OS" in
    darwin)
      if [ "$ARCH" = "arm64" ] && [ -w "/opt/homebrew/bin" ]; then
        echo "/opt/homebrew/bin"
      elif [ -w "/usr/local/bin" ]; then
        echo "/usr/local/bin"
      else
        echo "$HOME/.local/bin"
      fi ;;
    linux)
      [ -w "/usr/local/bin" ] && echo "/usr/local/bin" || echo "$HOME/.local/bin" ;;
    windows)
      echo "$HOME/bin" ;;
  esac
}

INSTALL_DIR="$(choose_dir)"
mkdir -p "$INSTALL_DIR"

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "ℹ $INSTALL_DIR isn’t on your PATH – add it to your shell profile." ;;
esac

# ── Fetch download URL ─────────────────────────────────────────────────────────
echo "⧗ Checking latest release…"
if command -v jq >/dev/null 2>&1; then
  DL_URL="$(curl -fsSL "$API" | jq -r --arg NAME "$ASSET" '.assets[] | select(.name==$NAME) | .browser_download_url')"
else
  DL_URL="$(curl -fsSL "$API" | grep -Eo "https://[^\"[:space:]]*/$ASSET" | head -n1)"
fi
[ -n "$DL_URL" ] || { echo "✖ Couldn’t find asset $ASSET in latest release."; exit 1; }

# ── Download, rename, install ─────────────────────────────────────────────────
TMP="$(mktemp)"
echo "⇣ Downloading $ASSET …"
curl -# -L "$DL_URL" -o "$TMP"
chmod +x "$TMP"

INSTALL_NAME="$BINARY"
[ "$OS" = "windows" ] && INSTALL_NAME="${INSTALL_NAME}.exe"

mv "$TMP" "$INSTALL_DIR/$INSTALL_NAME"

echo "✔ Installed $INSTALL_NAME to $INSTALL_DIR"
