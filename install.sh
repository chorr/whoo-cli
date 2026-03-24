#!/usr/bin/env bash
# whoo-cli 설치 스크립트
# 사용법:
#   curl -fsSL https://raw.githubusercontent.com/chorr/whoo-cli/main/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/chorr/whoo-cli/main/install.sh | bash -s -- v0.3.0

set -euo pipefail

REPO="chorr/whoo-cli"
BINARY_NAME="whooing"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
TMP_DIR=""

# OS/아키텍처 감지
detect_platform() {
    local os arch

    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$os" in
        linux)  os="linux" ;;
        darwin) os="darwin" ;;
        *)
            echo "[오류] 지원하지 않는 OS: $os"
            exit 1
            ;;
    esac

    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64)  arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *)
            echo "[오류] 지원하지 않는 아키텍처: $arch"
            exit 1
            ;;
    esac

    echo "${os}_${arch}"
}

# 최신 버전 조회
get_latest_version() {
    local version
    version="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')"

    if [ -z "$version" ]; then
        echo "[오류] 최신 버전을 가져올 수 없습니다"
        exit 1
    fi
    echo "$version"
}

# 체크섬 검증
verify_checksum() {
    local file="$1" checksums_file="$2"
    local filename expected_line

    filename="$(basename "$file")"
    expected_line="$(grep "$filename" "$checksums_file")"

    if [ -z "$expected_line" ]; then
        echo "[오류] 체크섬 파일에 ${filename} 항목이 없습니다"
        return 1
    fi

    if command -v sha256sum &>/dev/null; then
        echo "$expected_line" | sha256sum -c --quiet
    elif command -v shasum &>/dev/null; then
        echo "$expected_line" | shasum -a 256 -c --quiet
    else
        echo "[경고] sha256sum/shasum을 찾을 수 없어 체크섬 검증을 건너뜁니다"
        return 0
    fi
}

main() {
    local version platform archive_name download_url checksums_url

    # 버전 결정
    version="${1:-}"
    if [ -z "$version" ]; then
        echo "[정보] 최신 버전 확인 중..."
        version="$(get_latest_version)"
    fi
    echo "[정보] 설치 버전: ${version}"

    # 플랫폼 감지
    platform="$(detect_platform)"
    echo "[정보] 플랫폼: ${platform}"

    # 다운로드 URL 구성
    local version_num="${version#v}"
    archive_name="whoo-cli_${version_num}_${platform}.tar.gz"
    download_url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"
    checksums_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"

    # 임시 디렉토리
    TMP_DIR="$(mktemp -d)"
    trap 'rm -rf "$TMP_DIR"' EXIT

    # 다운로드
    echo "[정보] 다운로드 중... ${archive_name}"
    if ! curl -fsSL "$download_url" -o "${TMP_DIR}/${archive_name}"; then
        echo "[오류] 다운로드 실패: ${download_url}"
        exit 1
    fi

    if ! curl -fsSL "$checksums_url" -o "${TMP_DIR}/checksums.txt"; then
        echo "[오류] 체크섬 파일 다운로드 실패"
        exit 1
    fi

    # 체크섬 검증
    echo "[정보] 체크섬 검증 중..."
    (cd "$TMP_DIR" && verify_checksum "$archive_name" "checksums.txt")

    # 압축 해제
    tar -xzf "${TMP_DIR}/${archive_name}" -C "$TMP_DIR"

    # 설치
    mkdir -p "$INSTALL_DIR"
    mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    echo "[완료] ${INSTALL_DIR}/${BINARY_NAME} 에 설치됨"

    # PATH 확인
    if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
        echo ""
        echo "[주의] ${INSTALL_DIR} 이 PATH에 없습니다"
        echo "  셸 설정에 다음을 추가하세요:"
        echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    fi
}

main "$@"
