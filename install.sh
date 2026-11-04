#!/bin/sh
# Install the paguro binary from GitHub releases into ~/.local/bin.
#
#   curl -fsSL https://raw.githubusercontent.com/sebag90/paguro/main/install.sh | sh
#   curl -fsSL .../install.sh | sh -s -- v0.3.1        # pin a version
#   PAGURO_INSTALL_DIR=~/bin sh install.sh             # install elsewhere
set -eu

repo=sebag90/paguro
version=${1:-${PAGURO_VERSION:-latest}}
dir=${PAGURO_INSTALL_DIR:-$HOME/.local/bin}

case "$(uname -s)-$(uname -m)" in
  Linux-x86_64)              asset=paguro-linux-amd64 ;;
  Linux-aarch64|Linux-arm64) asset=paguro-linux-arm64 ;;
  Linux-armv6l|Linux-armv7l) asset=paguro-linux-arm ;;
  Darwin-x86_64)             asset=paguro-darwin-amd64 ;;
  Darwin-arm64)              asset=paguro-darwin-arm64 ;;
  FreeBSD-x86_64)            asset=paguro-freebsd-amd64 ;;
  FreeBSD-arm64)             asset=paguro-freebsd-arm64 ;;
  *) echo "unsupported platform: $(uname -s)-$(uname -m)" >&2; exit 1 ;;
esac

url="https://github.com/$repo/releases/$version/download/$asset"
echo "==> downloading $url"

tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
curl -fsSL -o "$tmp" "$url"
[ -s "$tmp" ] || { echo "download failed: empty response" >&2; exit 1; }

mkdir -p "$dir"
install -m 0755 "$tmp" "$dir/paguro"
ln -sf paguro "$dir/pg"
echo "==> installed $dir/paguro (and the pg alias) ($version)"

case ":$PATH:" in
  *":$dir:"*) ;;
  *) echo "note: $dir is not in PATH; add it to your shell profile" >&2 ;;
esac
