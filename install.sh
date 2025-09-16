#!/usr/bin/env bash
set -euo pipefail

OWNER=gordonmurray
REPO=flycost
LATEST=${1:-$(curl -sSLI -o /dev/null -w %{url_effective} https://github.com/$OWNER/$REPO/releases/latest | awk -F/ '{print $NF}')}
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH=amd64;;
  aarch64|arm64) ARCH=arm64;;
  *) echo "unsupported architecture: $ARCH"; exit 1;;
esac

URL="https://github.com/$OWNER/$REPO/releases/download/$LATEST/flycost_${OS}_${ARCH}.tar.gz"
echo "Downloading flycost $LATEST for $OS/$ARCH..."

tmp=$(mktemp -d)
trap "rm -rf $tmp" EXIT

if command -v curl >/dev/null 2>&1; then
  curl -sSLf "$URL" | tar -xz -C "$tmp"
elif command -v wget >/dev/null 2>&1; then
  wget -qO- "$URL" | tar -xz -C "$tmp"
else
  echo "Error: curl or wget required"
  exit 1
fi

if [ -w /usr/local/bin ]; then
  mv "$tmp/flycost" /usr/local/bin/flycost
  echo "✅ Installed flycost $LATEST → /usr/local/bin/flycost"
else
  sudo mv "$tmp/flycost" /usr/local/bin/flycost
  echo "✅ Installed flycost $LATEST → /usr/local/bin/flycost (with sudo)"
fi

echo "🎉 Installation complete! Run 'flycost -h' to get started."