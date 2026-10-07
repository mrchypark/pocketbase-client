#!/bin/sh
# Offline installer regression: verified download, mismatch, and missing checksum.
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
mkdir "$tmp/bin" "$tmp/work" "$tmp/assets"
asset=pocketbase-client_0.4.1_linux_amd64
printf 'verified binary\n' > "$tmp/assets/$asset"
(cd "$tmp/assets"; if command -v sha256sum >/dev/null; then sha256sum "$asset"; else shasum -a 256 "$asset"; fi) > "$tmp/assets/checksums.txt"
cat > "$tmp/bin/uname" <<'SH'
#!/bin/sh
case $1 in -s) echo Linux;; -m) echo x86_64;; *) exit 1;; esac
SH
cat > "$tmp/bin/curl" <<'SH'
#!/bin/sh
set -eu
while [ "$#" -gt 0 ]; do
    case $1 in -o) output=$2; shift 2;; -*) shift;; *) url=$1; shift;; esac
done
cp "$INSTALL_TEST_ASSETS/${url##*/}" "$output"
SH
chmod +x "$tmp/bin/uname" "$tmp/bin/curl"
export INSTALL_TEST_ASSETS="$tmp/assets"
export PATH="$tmp/bin:$PATH"
cd "$tmp/work"
sh "$root/install.sh" v0.4.1 >/dev/null
cmp pbc-gen "$tmp/assets/$asset"
test -x pbc-gen
printf 'existing installation\n' > pbc-gen
cp pbc-gen expected
printf 'tampered binary\n' > "$tmp/assets/$asset"
if sh "$root/install.sh" v0.4.1 >/dev/null 2>&1; then echo 'accepted checksum mismatch' >&2; exit 1; fi
cmp pbc-gen expected
: > "$tmp/assets/checksums.txt"
if sh "$root/install.sh" v0.4.1 >/dev/null 2>&1; then echo 'accepted missing checksum entry' >&2; exit 1; fi
cmp pbc-gen expected
echo 'installer checks passed'
