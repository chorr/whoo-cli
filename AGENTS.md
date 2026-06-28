# AGENTS.md

## Project Overview

**whoo-cli** — Go 기반 후잉(Whooing) 가계부 TUI + CLI 애플리케이션 (MIT, 오픈소스)

TUI 모드(bubbletea views 패턴)와 CLI 모드(API JSON 출력)를 병행 지원합니다.
API 호출 로직(`api/`)을 공유하여 TUI/CLI 양쪽에서 재활용합니다.

## Project Structure

```
whoo-cli/
├── main.go                   # 진입점, 서브커맨드 디스패치 (TUI/CLI 분기)
├── cmd/
│   ├── app.go               # TUI 통합 앱 모델 (views 패턴)
│   ├── styles.go            # TUI 공통 스타일 정의
│   ├── *_sub.go             # TUI 화면별 서브 모델
│   ├── *.go                 # CLI 서브커맨드 핸들러
│   └── root.go              # 공통 헬퍼 (포맷, 클라이언트 생성 등)
├── api/                     # Whooing API 클라이언트 (TUI/CLI 공용)
├── auth/                    # OAuth PIN 인증
├── config/                  # 설정 관리
├── docs/                    # 공개 API·CLI 문서
├── .github/workflows/       # 릴리스 CI (GoReleaser)
├── .goreleaser.yml          # 크로스 플랫폼 빌드 설정
└── install.sh               # curl | bash 설치 스크립트
```

## Tech Stack

- Go 1.22+
- [bubbletea](https://github.com/charmbracelet/bubbletea) v0.25.0 — TUI 프레임워크
- [lipgloss](https://github.com/charmbracelet/lipgloss) v0.10.0 — 스타일링
- [bubbles](https://github.com/charmbracelet/bubbles) v0.18.0 — TUI 컴포넌트
- [GoReleaser](https://goreleaser.com/) — 크로스 플랫폼 릴리스

## Architecture

### Views 패턴

TUI는 `cmd/app.go`의 `appModel`이 상태 머신 역할을 하며, 화면별 서브 모델(`*_sub.go`)을 생성·소멸합니다.
화면 전환은 bubbletea 메시지(`authCompleteMsg`, `sectionSelectedMsg`, `menuSelectionMsg`, `backToMenuMsg` 등)로 처리합니다.

### 명령어 구조

```
whoo              → TUI 실행
whoo auth         → TUI 실행 (인증 플로우)
whoo status       → 인증/설정 상태 확인 (CLI)
whoo user         → 유저 정보 JSON 출력
whoo user_logs    → 유저 로그 JSON 출력
whoo sections, s  → 섹션 관리 (CLI)
whoo accounts, a  → 항목 관리 (CLI)
whoo entries, e   → 거래내역 조회 (CLI)
whoo frequent, f  → 자주 쓰는 거래 (CLI)
whoo monthly, m   → 월별 요약 (CLI)
whoo inout, io    → 수입/지출 (CLI)
whoo budget       → 예산 (CLI)
whoo bill, b      → 청구서 (CLI)
whoo checkcard, cc → 카드 확인 (CLI)
whoo help         → 도움말
```

### 패키지 의존 방향

```
main.go → cmd/app.go → cmd/*_sub.go → api/ → config/
       → cmd/*.go (CLI) → api/ → config/
       → auth/ → config/
```

## Code Conventions

- **언어**: Go 1.22+
- **주석**: 한국어 사용
- **에러 처리**: 명시적 반환, `panic` 사용 금지
- **TUI 출력**: 이모지 사용 금지
- **TUI 스타일**: `cmd/styles.go`에 정의, 서브모델에서 재선언 금지
- **API 클라이언트**: `NewClient(cfg)` 사용 (`api.NewWhooingClient` 직접 호출 금지)
- **API 응답 파싱**: `parseResponse()` 헬퍼 사용
- **CLI 플래그**: `flag.NewFlagSet` 사용
- **Whooing 용어**: 왼쪽/오른쪽/아이템/메모 (공식 CSV 용어 준수)

### Views 패턴 규칙

각 TUI 화면은 `Init/Update/View`를 갖는 서브 모델로 구현합니다.
목록·선택 화면은 vi 스타일 키(`j`/`k`)와 화살표 키를 동시 지원합니다.
도움말 표기: `[↑/↓/j/k] 이동` 형식

## API 규칙

- Base URL: `https://whooing.com/api/`
- 인증 헤더: `X-API-KEY: app_id={app_id},token={token},signiture={signiture},timestamp={unix_ms}`
- `signiture`는 Whooing API 원문 그대로 사용 (오타 아님)
- 응답 래퍼: `{"code":200,"error":{},"results":{...}}`

| 확장자 | 응답 형식 | 설명 |
|--------|-----------|------|
| `.json` | 객체 `{key: value}` | 키-값 형태 (map) |
| `.json_array` | 배열 또는 래핑된 객체 | endpoint마다 다름 |

상세: `docs/api-reference.md`, `docs/api-*.md`

## 복식부기 규칙

| 계정 | 코드 | 설명 |
|------|------|------|
| 자산 | `assets` | 왼쪽(l_account): 돈이 도착 |
| 부채 | `liabilities` | 오른쪽(r_account): 돈이 나감 |
| 자본 | `capital` | 오른쪽 |
| 비용 | `expenses` | 왼쪽 |
| 수익 | `income` | 오른쪽 |

## AI Guidelines

### 할 것 (DO)

- Go 표준 패턴 (`gofmt`) 준수
- 에러는 `fmt.Errorf("context: %w", err)` 형태로 반환
- TUI는 views 패턴(`cmd/app.go` 중심)으로 구현
- 화면 전환은 메시지로 처리
- `tea.WithAltScreen()` 전체 화면 모드 사용
- CLI API 메서드는 raw `[]byte` 반환 (TUI/CLI 재활용)
- CLI 커맨드 추가 시 `docs/dev-cli-guide.md` 절차 따를 것
- API 응답 파싱은 `parseResponse()` 또는 `parseEntryArrayResponse()` 사용

### 하지 말 것 (DON'T)

- `panic()` 사용 금지
- `os.Exit()`은 `main.go`, `cmd/root.go`의 RequireAuth/RequireSection, CLI 핸들러에서만 사용
- config·소스코드에 비밀값 하드코딩 금지
- TUI 출력에 이모지 사용 금지
- Whooing API 구조를 모르면 추측하지 말 것 — `docs/api-*.md` 참조
- 서브 모델 간 직접 참조 금지 (appModel을 통해 통신)
- CLI에서 API 응답을 별도 파싱하지 말 것 (raw JSON 출력)
- 서브모델 View()에서 공통 스타일 로컬 재선언 금지

## Build & Development

```bash
# 로컬 빌드 (후잉 앱 자격 증명은 환경변수로 제공)
export WHOOING_APP_ID=your_app_id
export WHOOING_APP_SECRET=your_app_secret
go build -o whoo .

# 검증
go vet ./...
go test ./...

# 실행
./whoo              # TUI
./whoo status       # 상태 확인
./whoo entries      # 거래내역 (JSON)
```

배포 바이너리는 GoReleaser 빌드 시 ldflags로 앱 자격 증명이 주입됩니다.
로컬 개발 시에는 환경변수 또는 `-ldflags`로 동일하게 설정할 수 있습니다.

## Release & CI

태그 push(`v*`) 시 GitHub Actions Release 워크플로가 GoReleaser를 실행합니다.

1. linux/darwin × amd64/arm64 바이너리 빌드
2. tar.gz 아카이브 + `checksums.txt` 생성
3. GitHub Releases에 업로드

릴리스 에셋 이름:

```
whoo-cli_{version}_{os}_{arch}.tar.gz
checksums.txt
```

워크플로 재실행 시 에셋 충돌을 방지하려면 `.goreleaser.yml`의 `replace_existing_artifacts: true`를 유지합니다.

## Environment Variables

| 환경변수 | 설명 | 필수 |
|----------|------|------|
| `WHOOING_APP_ID` | 후잉 앱 ID (로컬 빌드 시) | 로컬 빌드 시 |
| `WHOOING_APP_SECRET` | 후잉 앱 Secret (로컬 빌드 시) | 로컬 빌드 시 |
| `WHOOING_PIN` | OAuth PIN (자동 인증용, 선택) | 아니오 |

## External References

- [Whooing API Docs](https://whooing.com/#main/api)
- [Bubbletea](https://github.com/charmbracelet/bubbletea)
- [Lipgloss](https://github.com/charmbracelet/lipgloss)
- [Bubbletea Views Example](https://github.com/charmbracelet/bubbletea/tree/main/examples/views)

## Notes

- 설정 파일: `~/.config/whoo-cli/config.json`
- Config 저장 항목: `token`, `token_secret`, `section_id`
- OAuth PIN 방식: RequestToken → Authorize → ExchangeToken
- TUI 메인 메뉴: 거래내역, 거래 입력, 자산/부채, 섹션 변경, 사용자 정보, 섹션 관리, 항목 관리, 흐름 분석, 카드 관리, 예산·목표
- CLI 모드: API 응답 raw JSON pretty-print (`.json` 포맷 사용)
- CLI 구현 상세: `docs/dev-cli-guide.md`