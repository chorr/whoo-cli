# whoo-cli

후잉(Whooing) 가계부 TUI + CLI 애플리케이션

## 설치

```bash
curl -fsSL https://raw.githubusercontent.com/chorr/whoo-cli/main/install.sh | bash
```

특정 버전 설치:

```bash
curl -fsSL https://raw.githubusercontent.com/chorr/whoo-cli/main/install.sh | bash -s -- v1.0.1
```

`~/.local/bin/whoo`에 설치됩니다.

## 사용법

```bash
# TUI 실행
whoo

# 인증/설정 상태 확인
whoo status
whoo version

# CLI 모드 (JSON 출력)
whoo user          # 유저 정보
whoo sections      # 섹션 목록
whoo accounts      # 항목 메타 (잔액 없음)
whoo bs            # 자산/부채 잔액
whoo inout         # 기간 자금증감
whoo entries       # 거래내역
whoo help          # 도움말
```

## 인증

최초 실행 시 OAuth PIN 인증이 필요합니다.

TTY가 있는 터미널:

```bash
whoo auth
```

에이전트/헤드리스:

```bash
whoo auth --url          # 인증 URL을 stdout에 출력
whoo auth --pin <PIN>    # PIN으로 토큰 교환
# 또는 WHOOING_PIN=<PIN> whoo auth
```

인증 토큰은 `~/.config/whoo/config.json`에 저장됩니다.

## 개발

Go 1.22 이상이 필요합니다.

```bash
cp .env.example .env
# .env에 본인의 Whooing 앱 자격 증명 입력

make dev
make check
```

`.env`와 로컬 빌드 결과물은 Git에서 제외됩니다. 실제 앱 자격 증명과 인증 토큰을 커밋하지 마세요.

## 지원 플랫폼

| OS | Architecture |
|----|-------------|
| Linux | amd64, arm64 |
| macOS | amd64, arm64 |

## 라이선스

MIT