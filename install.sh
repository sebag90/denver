#!/bin/sh
# Install the dv binary from GitHub releases into ~/.local/bin.
#
#   curl -fsSL https://raw.githubusercontent.com/sebag90/denver/main/install.sh | sh
#   curl -fsSL .../install.sh | sh -s -- v0.3.1        # pin a version
#   DV_INSTALL_DIR=~/bin sh install.sh                 # install elsewhere
set -eu

repo=sebag90/denver
version=${1:-${DV_VERSION:-latest}}
dir=${DV_INSTALL_DIR:-$HOME/.local/bin}

case "$(uname -s)-$(uname -m)" in
  Linux-x86_64)              asset=dv-linux-amd64 ;;
  Linux-aarch64|Linux-arm64) asset=dv-linux-arm64 ;;
  Linux-armv6l|Linux-armv7l) asset=dv-linux-arm ;;
  Darwin-x86_64)             asset=dv-darwin-amd64 ;;
  Darwin-arm64)              asset=dv-darwin-arm64 ;;
  FreeBSD-x86_64)            asset=dv-freebsd-amd64 ;;
  FreeBSD-arm64)             asset=dv-freebsd-arm64 ;;
  *) echo "unsupported platform: $(uname -s)-$(uname -m)" >&2; exit 1 ;;
esac

url="https://github.com/$repo/releases/$version/download/$asset"
echo "==> downloading $url"

tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
curl -fsSL -o "$tmp" "$url"
[ -s "$tmp" ] || { echo "download failed: empty response" >&2; exit 1; }

mkdir -p "$dir"
install -m 0755 "$tmp" "$dir/dv"
echo "==> installed $dir/dv ($version)"

case ":$PATH:" in
  *":$dir:"*) ;;
  *) echo "note: $dir is not in PATH; add it to your shell profile" >&2 ;;
esac
