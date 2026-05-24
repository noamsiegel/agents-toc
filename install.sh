#!/usr/bin/env sh
# install.sh — fetch the latest agents-toc binary for this OS/arch.
# Usage:  curl -sSf https://raw.githubusercontent.com/noamsiegel/agents-toc/main/install.sh | sh
#
# Verifies the SHA-256 checksum from the release's checksums.txt before
# installing. Override the destination with PREFIX=/usr/local; default
# is $HOME/.local.
set -eu

REPO="noamsiegel/agents-toc"
BIN="agents-toc"
PREFIX="${PREFIX:-$HOME/.local}"
TARGET_DIR="$PREFIX/bin"

os="$(uname -s)"
arch="$(uname -m)"

case "$os" in
  Darwin) os_tag="Darwin" ;;
  Linux)  os_tag="Linux" ;;
  *)      printf 'unsupported OS: %s\n' "$os" >&2; exit 1 ;;
esac

case "$arch" in
  x86_64|amd64) arch_tag="x86_64" ;;
  arm64|aarch64) arch_tag="arm64" ;;
  *) printf 'unsupported arch: %s\n' "$arch" >&2; exit 1 ;;
esac

if command -v sha256sum >/dev/null 2>&1; then
  sha_cmd="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  sha_cmd="shasum -a 256"
else
  printf 'no sha256 tool found (need sha256sum or shasum)\n' >&2
  exit 1
fi

mkdir -p "$TARGET_DIR"

api="https://api.github.com/repos/$REPO/releases/latest"
version="$(curl -sSfL "$api" | sed -n 's/.*"tag_name":[[:space:]]*"v\([^"]*\)".*/\1/p' | head -1)"
if [ -z "$version" ]; then
  printf 'could not determine latest version from %s\n' "$api" >&2
  exit 1
fi

archive="agents-toc_${version}_${os_tag}_${arch_tag}.tar.gz"
base_url="https://github.com/$REPO/releases/download/v${version}"
url="$base_url/$archive"
sums_url="$base_url/checksums.txt"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

printf 'fetching %s\n' "$url"
curl -sSfL "$url" -o "$tmp/$archive"
curl -sSfL "$sums_url" -o "$tmp/checksums.txt"

expected="$(grep " $archive$" "$tmp/checksums.txt" | awk '{print $1}')"
if [ -z "$expected" ]; then
  printf 'no checksum line for %s in checksums.txt\n' "$archive" >&2
  exit 1
fi
actual="$(cd "$tmp" && $sha_cmd "$archive" | awk '{print $1}')"
if [ "$expected" != "$actual" ]; then
  printf 'checksum mismatch for %s:\n  expected %s\n  actual   %s\n' "$archive" "$expected" "$actual" >&2
  exit 1
fi

printf 'checksum verified (%s)\n' "$actual"
tar -xzf "$tmp/$archive" -C "$tmp"

install -m 0755 "$tmp/$BIN" "$TARGET_DIR/$BIN"
printf '%s installed at %s\n' "$BIN" "$TARGET_DIR/$BIN"
case ":$PATH:" in
  *":$TARGET_DIR:"*) ;;
  *) printf 'note: %s is not on PATH — add it to your shell rc.\n' "$TARGET_DIR" ;;
esac
