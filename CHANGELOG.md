# Changelog

## 1.2.0 - 2026-09-16

### 주요 변경사항

- `whoo bs`의 자산·부채 잔액 조회 오류를 수정했습니다.
- `whoo balance --account-id <id>`로 특정 항목의 현재 잔액을 바로
  조회할 수 있습니다.
- 여러 계정을 함께 조회하는 `whoo report`의 안정성을 개선했습니다.

### 에이전트 및 자동화 지원

- `accounts`와 `entries`에 파싱하기 쉬운 `--flat` JSON 출력을
  추가했습니다.
- `accounts --with-balance`로 항목 정보와 현재 잔액을 함께 조회할 수
  있습니다.
- `entries search`에서 `--account-id`, `--l-id`, `--r-id`를 사용할 때
  계정 종류를 자동으로 확인합니다.
- `--section`과 `WHOO_SECTION`으로 실행할 섹션을 명확하게 지정할 수
  있습니다.
- JSON 응답에 실제 실행 대상인 `section_id`가 포함됩니다.

### 오류 및 도움말

- API 오류를 짧고 일관된 형식으로 표시하며, 서버의 HTML 응답은
  기본적으로 숨깁니다. 자세한 내용은 `--verbose`로 확인할 수 있습니다.
- `entries add --split`의 할부 금액 분배 규칙과 관련 예시를
  추가했습니다.

### 호환성

- 기존 명령어와 기본 JSON 구조는 유지됩니다.
- 기존 JSON 응답 최상위에 `section_id`가 추가됩니다.
