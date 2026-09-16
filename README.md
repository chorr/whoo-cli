# whoo-cli

후잉(Whooing) 가계부 TUI + CLI 애플리케이션

## 설치

```bash
curl -fsSL https://raw.githubusercontent.com/chorr/whoo-cli/main/install.sh | bash
```

특정 버전 설치:

```bash
curl -fsSL https://raw.githubusercontent.com/chorr/whoo-cli/main/install.sh | bash -s -- v1.2.0
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
whoo accounts --flat
whoo bs            # 자산/부채 잔액
whoo balance --account-id x2
whoo report        # 통합 보고서 (기간별 손익/잔액)
whoo inout         # 기간 자금증감
whoo entries       # 거래내역
whoo help          # 도움말
```

### 에이전트/비대화형 사용

모든 API 커맨드는 JSON을 stdout에 출력하며, 실제 조회·기록 대상인
`section_id`를 응답에 포함합니다. 섹션 우선순위는
`--section` → `WHOO_SECTION` → `~/.config/whoo/config.json`입니다.

```bash
export WHOO_SECTION=s36258

# 한 항목의 현재 잔액
whoo balance --account-id x2

# 파싱하기 쉬운 항목/거래 배열
whoo accounts --flat
whoo accounts assets --flat --with-balance
whoo entries search --account-id x2 \
  --from 20260301 --to 20260916 --flat

# 커맨드 앞뒤 어디서든 전역 섹션 지정 가능
whoo --section s36258 bs
whoo bs --section s36258
```

기본 출력은 기존 후잉 API 응답 구조를 유지합니다. `--flat`은 중첩된
래퍼를 제거한 compact 배열이며 각 행에 `section_id`가 있습니다.

다른 섹션을 명시해 쓰기 커맨드를 실행하면 stderr에 경고합니다.
의도한 자동화라면 `--allow-section-override`로 경고를 숨길 수 있습니다.

API 오류는 기본적으로 `code`, `endpoint`, `reason`과 짧은 설명만
stderr에 표시합니다. 서버가 반환한 HTML 등 원문은 `--verbose`에서만
확인할 수 있습니다.

### 잔액과 보고서

`whoo bs`와 `whoo balance`는 공식 report API의 `report/assets.json`과
`report/liabilities.json`을 각각 호출해 병합합니다. 따라서
`/report/assets,liabilities.json` 콤마 경로의 HTTP 403 영향을 받지
않으며 deprecated된 `bs.json`에도 의존하지 않습니다.

`whoo report assets,liabilities`는 계정별 report API를 각각 호출한 뒤
기존 다중 계정 JSON 구조로 병합합니다.

### 할부 금액 분배

`entries add --split N`은 아이템에 후잉의 `//N` 명령어를 붙여 서버로
전송합니다. 후잉 서버는 카드 항목의 ‘할부입력시 처리방식’ 단위
(일반적으로 1원 또는 100원)에 맞춰 이후 회차 금액을 절삭하고,
나머지를 첫 회차에 더합니다.

예를 들어 100원 단위 설정에서 2,290,000원을 12개월로 나누면
첫 회차 191,200원, 이후 11회는 190,800원입니다. CLI가 로컬에서
회차별 금액을 다시 계산하지 않으므로 최종 규칙은 후잉 항목 설정을
따릅니다.

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
# WHOOING_APP_ID, WHOOING_APP_SECRET 입력 (WHOOING_PIN은 선택)

make dev
make check
```

`make`와 실행 시 `.env`를 읽어 자격 증명을 채웁니다. `.env`와 로컬 빌드 결과물은 Git에서 제외되므로 실제 값과 인증 토큰을 커밋하지 마세요.

## 지원 플랫폼

| OS | Architecture |
|----|-------------|
| Linux | amd64, arm64 |
| macOS | amd64, arm64 |

## 라이선스

MIT