#!/usr/bin/env sh
# Install the latest platform-cli release for Linux/macOS.
set -eu
REPO="cypher682/platform-cli"
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac
tag=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')
[ -n "$tag" ] || { echo "could not determine latest release" >&2; exit 1; }
ver=${tag#v}
url="https://github.com/$REPO/releases/download/$tag/platform-cli_${ver}_${os}_${arch}.tar.gz"
tmp=$(mktemp -d)
curl -fsSL "$url" | tar -xz -C "$tmp"
dest=${INSTALL_DIR:-/usr/local/bin}
if [ -w "$dest" ]; then mv "$tmp/platform-cli" "$dest/"; else sudo mv "$tmp/platform-cli" "$dest/"; fi
echo "installed platform-cli $tag to $dest"
