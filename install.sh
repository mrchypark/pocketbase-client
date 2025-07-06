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
    # GitHub API를 사용해 최신 릴리스의 태그 이름을 가져옵니다.
    # jq가 설치되어 있으면 더 안정적으로 파싱할 수 있지만, 호환성을 위해 grep/sed를 사용합니다.
    curl --silent "https://api.github.com/repos/mrchypark/pocketbase-client/releases/latest" |
        grep '"tag_name":' |
        sed -E 's/.*"([^"]+)".*/\1/'
}

# --- Main Logic ---

# 첫 번째 인자를 버전으로 사용합니다. 비어있으면 최신 릴리스를 조회합니다.
VERSION=${1:-$(get_latest_release)}

if [ -z "$VERSION" ]; then
    echo "Error: Could not determine the release version to install."
    echo "Please specify a version tag, for example: ./install.sh v0.1.2"
    exit 1
fi

echo "Installing version: ${VERSION}"

# OS 및 아키텍처 자동 탐지
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

# GoReleaser 이름 형식에 맞춰 다운로드 URL 구성
# `LATEST_RELEASE#v` -> 버전 태그에서 'v' 접두사를 제거합니다 (e.g., v0.1.3-rc -> 0.1.3-rc).
FILENAME="pocketbase-client_${VERSION#v}_${OS}_${ARCH}"
DOWNLOAD_URL="https://github.com/mrchypark/pocketbase-client/releases/download/${VERSION}/${FILENAME}"

echo "Downloading pbc-gen from ${DOWNLOAD_URL}..."

# curl로 바이너리 다운로드 및 실행 권한 부여
curl -L -o "pbc-gen" "$DOWNLOAD_URL"
chmod +x "pbc-gen"

echo ""
echo "✅ pbc-gen (${VERSION}) has been downloaded successfully to the current directory."
echo "   You can move it to a directory in your PATH, for example:"
echo "   sudo mv ./pbc-gen /usr/local/bin/pbc-gen"