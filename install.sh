#!/bin/sh
# install.sh
#
# Installs the pbc-gen binary.
#
# Usage:
#   - To install the latest version: ./install.sh
#   - To install a specific version: ./install.sh v0.1.2

set -e

# --- Helper Functions ---
get_latest_release() {
    # Get the latest release tag name using GitHub API.
    # While jq would provide more stable parsing, we use grep/sed for compatibility.
    curl --silent "https://api.github.com/repos/mrchypark/pocketbase-client/releases/latest" |
        grep '"tag_name":' |
        sed -E 's/.*"([^"]+)".*/\1/'
}

# --- Main Logic ---

# Use the first argument as version. If empty, query the latest release.
VERSION=${1:-$(get_latest_release)}

if [ -z "$VERSION" ]; then
    echo "Error: Could not determine the release version to install."
    echo "Please specify a version tag, for example: ./install.sh v0.1.2"
    exit 1
fi

echo "Installing version: ${VERSION}"

# Auto-detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64 | arm64) ARCH="arm64" ;;
    *)
        echo "Error: Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Construct download URL according to GoReleaser naming format
# `LATEST_RELEASE#v` -> Remove 'v' prefix from version tag (e.g., v0.1.3-rc -> 0.1.3-rc).
FILENAME="pocketbase-client_${VERSION#v}_${OS}_${ARCH}"
DOWNLOAD_URL="https://github.com/mrchypark/pocketbase-client/releases/download/${VERSION}/${FILENAME}"

echo "Downloading pbc-gen from ${DOWNLOAD_URL}..."

# Download binary with curl and grant execution permission
curl -L -o "pbc-gen" "$DOWNLOAD_URL"
chmod +x "pbc-gen"

echo ""
echo "✅ pbc-gen (${VERSION}) has been downloaded successfully to the current directory."
echo "   You can move it to a directory in your PATH, for example:"
echo "   sudo mv ./pbc-gen /usr/local/bin/pbc-gen"