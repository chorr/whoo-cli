# whoo-cli

후잉(Whooing) 가계부 TUI + CLI 애플리케이션

## 설치

```bash
curl -fsSL https://raw.githubusercontent.com/chorr/whoo-cli/main/install.sh | bash
```

특정 버전 설치:

```bash
curl -fsSL https://raw.githubusercontent.com/chorr/whoo-cli/main/install.sh | bash -s -- v0.3.0
```

`~/.local/bin/whooing`에 설치됩니다.

## 사용법

```bash
# TUI 실행
whooing

# 인증/설정 상태 확인
whooing status

# CLI 모드 (JSON 출력)
whooing user          # 유저 정보
whooing sections      # 섹션 목록
whooing accounts      # 항목 목록
whooing entries       # 거래내역
whooing help          # 도움말
```

## 인증

최초 실행 시 OAuth PIN 인증이 필요합니다:

1. `whooing` 실행
2. 브라우저에서 인증 URL 접속
3. 후잉 계정으로 로그인 후 PIN 번호 확인
4. PIN 번호 입력

인증 토큰은 `~/.config/whooing-cli/config.json`에 저장됩니다.

## 지원 플랫폼

| OS | Architecture |
|----|-------------|
| Linux | amd64, arm64 |
| macOS | amd64, arm64 |

## 라이선스

MIT
