# CLI 구현 가이드

## 개요

whoo-cli는 TUI 모드와 CLI 모드를 병행 지원한다.
- **TUI**: 인터랙티브 화면 (bubbletea 기반, 인자 없이 실행)
- **CLI**: API 호출 → JSON 출력 (서브커맨드 기반)

## CLI 설계 원칙

### 1. 출력 형식

- API 응답을 **파싱하지 않고** JSON 원본을 그대로 pretty-print 출력
- API 요청 시 `.json` 포맷 사용 (`.json_array` 아님)
- 출력 예시:
```json
{
  "code": 200,
  "message": "",
  "error_parameters": {},
  "rest_of_api": 4988,
  "results": {
    "user_id": 4,
    "username": "Helloman",
    ...
  }
}
```

### 2. API 로직 재활용

API 호출 로직은 `api/whooing.go`의 `WhooingClient`에 구현한다.
TUI와 CLI 모두 동일한 API 메서드를 사용한다.

```
api/whooing.go      ← API 메서드 (공통)
  ├── cmd/*_sub.go  ← TUI 서브모델에서 호출
  └── cmd/*.go      ← CLI 핸들러에서 호출
```

- API 메서드는 `[]byte` (raw JSON)를 반환하여 CLI에서 바로 출력 가능
- TUI에서 구조체 파싱이 필요하면 `parseResponse()` 활용

### 3. 커맨드 라우팅

`main.go`의 switch 문에서 서브커맨드를 분기한다.

인자가 없고 TTY가 없으면 TUI를 열지 않고 stderr 안내 후 종료한다.

### 3.1 help 예약어

`help`, `--help`, `-h`는 모든 깊이에서 예약어다.

- 이 경로에서는 `RequireAuth` / `RequireSection` / API / TUI를 실행하지 않는다
- TTY가 없어도 exit 0으로 한국어 사용법을 출력한다
- Go `flag` 기본 영어 Usage를 그대로 쓰지 않는다

### 4. CLI 핸들러 파일 규칙

| 항목 | 규칙 |
|------|------|
| 파일 위치 | `cmd/{커맨드명}.go` |
| 함수명 | `Run{커맨드명}(cfg *config.Config)` |
| 인증 확인 | `RequireAuth(cfg)` 호출 |
| API 클라이언트 | `NewClient(cfg)` 사용 |
| 에러 출력 | `PrintError()` 사용 후 `os.Exit(1)` |
| JSON 출력 | `json.Indent`로 pretty-print |

### 5. 새 CLI 커맨드 추가 절차

1. `api/whooing.go`에 API 메서드 추가 (없는 경우)
2. `cmd/{커맨드명}.go` 파일 생성
3. `main.go` switch에 case 추가

### 6. 에러 처리

- 인증 미완료: `RequireAuth()`가 안내 메시지 출력 후 종료
- API 오류: `PrintError()`로 stderr 출력 후 `os.Exit(1)`
- JSON indent 실패: 원본 출력으로 fallback

## 구현 현황

| 커맨드 | 상태 | 설명 |
|--------|------|------|
| `whoo user` | 완료 | 유저 정보 조회/수정 (`edit`) |
| `whoo user_logs` | 완료 | 유저 로그 조회 (`--max`/`--limit`) |
| `whoo user_point_logs` | 완료 | 유저 포인트 로그 조회 (`--max`/`--type`/`--limit`) |
| `whoo sections` | 완료 | 섹션 관리 (add/edit/delete/sort, `--ui` JSON) |
| `whoo accounts` | 완료 | 항목 관리 (add/edit/delete/sort, 잔액 없음) |
| `whoo bs` | 완료 | 자산/부채 잔액 (내부적으로 report API 사용) |
| `whoo report` | 완료 | 통합 보고서 / `summary` (bs/pl/zigzag 등 대체) |
| `whoo inout` | 완료 | 자금증감 (raw JSON, 구조체 파싱 없음) |
| `whoo entries` | 완료 | 거래내역 (add/batch/update/delete/search/agg/outside 등) |
| `whoo frequent` | 완료 | 자주입력 (list/get/add/edit/delete/sort/use) |
| `whoo monthly` | 완료 | 월별입력 (list/get/add/edit/delete/sort/pay) |
| `whoo budget` | 완료 | 예산 (get/set/basic_total/reset) |
| `whoo budget-goal` | 완료 | 장기 예산목표 (get/set/reset) |
| `whoo goal` | 완료 | 자본 목표 (get/set) |
| `whoo bill` | 완료 | 신용카드 청구 |
| `whoo checkcard` | 완료 | 체크카드 연동 |
| `whoo auth` | 완료 | TUI 또는 `--url`/`--pin` 헤드리스 |
| `whoo status` | 완료 | 인증/설정 상태 확인 |
| `whoo help` | 완료 | 도움말 표시 |

## 참고

- 공식 API 문서: `docs/whooing-api-official.md`
- API 문서: `docs/api-*.md`
- API 레퍼런스: `docs/api-reference.md`
