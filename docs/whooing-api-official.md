# Whooing Developer API Documentation

---

## 1. Getting Start

후잉의 거의 모든 기능은 OpenAPI를 통해서 직접 제어하실 수 있습니다. 이를통해 써드파티앱을 구현하여 배포하거나, 혹은 AI를 이용해 자신만의 가계부를 활용할 수도 있습니다.

### 관련 리소스

- whoochoo 기본 아이콘 셋트 ([Download icons](https://static.whooing.com/assets/img/whoochoo_icon.zip))
- 안드로이드 Whoochoo ([https://github.com/wisedog/Whoochoo](https://github.com/wisedog/Whoochoo))
- 안드로이드 후잉문자 자동입력기 ([https://github.com/wisedog/WhoochooAdvSms](https://github.com/wisedog/WhoochooAdvSms))
- OAuth 2.0 / API 테스트 도구 ([https://github.com/zidell/whooing-api-test](https://github.com/zidell/whooing-api-test))

---

## 2. Authorization

모든 리소스 접근은 보안프로토콜(https)을 이용하여 요청되어야 합니다.

### AI 도구에서 직접 사용

본인이 직접 외부 도구에서 사용하려는 목적이라면 키발급만 진행하여 연동이 가능합니다. 후잉 홈페이지의 [우측 상단에서 계정 > 비밀번호 및 보안]으로 이동 후에 [+AI 연동] 버튼을 눌러 인증키를 발급받습니다. 발급받은 인증키로 아래와 같이 API 호출을 진행하면 됩니다.

**curl 예시:**

```bash
curl -X GET "https://whooing.com/api/sections.json" \
  -H "X-API-KEY: 발급된_인증키"
```

### 배포용앱

#### 1. App 등록

[My Apps](https://whooing.com/forum/developer/ko/apps)에서 기본적인 정보를 입력하여 App을 등록하고 고유의 **app_id**와 **app_secret**을 부여받습니다. 수익을 추구하거나 App 자체에서 무제한의 API를 이용하려는 특수한 경우에는 등록 후에 support@whooing.com으로 연락을 주십시오.

OAuth 2.0을 사용하는 앱은 등록 시 **redirect_uri**도 함께 등록해야 합니다. 여러 개를 등록할 경우 줄바꿈으로 구분합니다.

**redirect_uri 등록 유형:**

| 유형 | 등록 예시 | 설명 |
| ---- | --------- | ---- |
| 웹 앱 | `https://example.com/callback` | 정확히 일치하는 URL만 허용 |
| 로컬 개발 | `http://localhost` | 포트 무관하게 모든 localhost 허용 (RFC 8252) |
| 모바일/데스크탑 앱 | `myapp://` | 해당 Custom URI Scheme으로 시작하는 URI 허용 |

#### 2. Token 발급 및 User 인증

User가 App을 통해 후잉의 리소스에 접근하기 위해서는 Access Token이 필요합니다. 에러가 있는 경우에는 `{ "error": "error_code", "error_description": "설명" }` 형태로 응답합니다.

---

##### OAuth 2.0 (권장)

OAuth 2.0 Authorization Code Flow with PKCE를 사용합니다. 서버 메타데이터는 아래 주소에서 자동으로 탐색할 수 있습니다.

```
GET https://whooing.com/.well-known/oauth-authorization-server
```

###### 단계 1: 사용자 인증 페이지로 이동

사용자 브라우저를 아래 주소로 리다이렉트합니다.

**요청주소:** `https://whooing.com/oauth2/authorize`

**파라미터:**

| 파라미터                | 필수 | 설명                                                       |
| ----------------------- | ---- | ---------------------------------------------------------- |
| `response_type`         | ✓    | `code` 고정                                                |
| `client_id`             | ✓    | 발급받은 app_id                                            |
| `redirect_uri`          | ✓    | 앱 등록 시 등록한 redirect URI. 웹은 `https://...`, 로컬 개발은 `http://localhost:포트/경로`, 앱은 `myapp://경로` |
| `scope`                 |      | 요청 권한 (콤마 구분). 예: `read,write`                    |
| `state`                 |      | CSRF 방지용 임의 문자열 (권장)                             |
| `code_challenge`        |      | PKCE code_verifier를 SHA-256 해시 후 Base64URL 인코딩한 값 |
| `code_challenge_method` |      | `S256` (PKCE 사용 시 필수)                                 |

**사용 가능한 scope:**

| scope     | 설명             |
| --------- | ---------------- |
| `read`    | 데이터 읽기      |
| `write`   | 데이터 쓰기/수정 |
| `post_it` | 포스트잇 접근    |
| `messages` | 메시지 접근      |
| `bbs`     | 게시판 접근      |

사용자가 승인하면 등록된 `redirect_uri`로 리다이렉트됩니다.

```
https://example.com/callback?code=a3f...&state=xyz
```

사용자가 거부하면:

```
https://example.com/callback?error=access_denied&state=xyz
```

###### 단계 2: Access Token 발급

**요청주소:** `POST https://whooing.com/oauth2/token`

**파라미터 (application/x-www-form-urlencoded):**

| 파라미터        | 필수 | 설명                                   |
| --------------- | ---- | -------------------------------------- |
| `grant_type`    | ✓    | `authorization_code`                   |
| `code`          | ✓    | 단계 1에서 발급받은 code               |
| `redirect_uri`  | ✓    | 단계 1과 동일한 redirect_uri           |
| `client_id`     | ✓    | 발급받은 app_id                        |
| `code_verifier` |      | PKCE 사용 시 원본 code_verifier 문자열 |

**응답:**

```json
{
	"access_token": "a3f9b2c1d4e5f6...",
	"token_type": "Bearer",
	"expires_in": 31536000,
	"refresh_token": "e5f6a3f9b2c1d4...",
	"scope": "read,write"
}
```

###### 단계 3: 리소스 요청

발급받은 `access_token`을 `Authorization` 헤더에 담아 요청합니다.

```bash
curl -X GET "https://whooing.com/api/sections.json" \
  -H "Authorization: Bearer a3f9b2c1d4e5f6..."
```

###### Access Token 갱신

`access_token`은 **1년** 유효하며, API 호출 시 만료까지 **6개월 이내**이면 자동으로 1년 연장됩니다. 정상적으로 사용 중인 토큰은 만료될 일이 없습니다.

`refresh_token`은 **영구** 유효합니다. `access_token`이 만료된 경우 `refresh_token`으로 갱신합니다.

- **사전 갱신**: `expires_in` 값을 저장해두고, 만료 임박 시점에 미리 갱신
- **지연 갱신**: API 호출 후 `401` 응답을 받았을 때 갱신 후 재요청

**요청주소:** `POST https://whooing.com/oauth2/token`

```
grant_type=refresh_token
refresh_token=e5f6a3f9b2c1d4...
client_id=45
```

응답 형식은 단계 2와 동일합니다. 새 `refresh_token`이 발급되고 기존 것은 폐기됩니다.

###### Token 폐기

**요청주소:** `POST https://whooing.com/oauth2/revoke`

```
token=a3f9b2c1d4e5f6...
```

---

##### OAuth 1.0a (레거시)

> 기존에 OAuth 1.0a로 개발된 앱은 계속 사용할 수 있습니다. 신규 개발에는 OAuth 2.0을 권장합니다.

###### 단계 1: Request Token 발급

**요청주소:** `https://whooing.com/app_auth/request_token`

**파라미터:**

- **app_id** (required): 발급받은 App ID
- **app_secret** (required): 발급받은 app_secret
- callbackuri: 토큰 발급완료 후 되돌아갈 페이지.

**응답 - 어플리케이션 (callbackuri 미지정):**

```json
{
	"token": "7580361ddabad0a55ab34ee880be438a0e5dc294"
}
```

**응답 - 웹서비스 (callbackuri 지정):** 해당 주소로 token 파라미터와 함께 리다이렉트됩니다.

###### 단계 2: 사용자 인증 및 PIN 발급

**요청주소:** `https://whooing.com/app_auth/authorize`

**파라미터:**

- **token** (required): 발급받은 token
- callbackuri: 인증완료 후 되돌아갈 페이지.
- no_register: 애플 등 가입화면을 제외해야 하는 경우 `no_register=y`

###### 단계 3: Access Token 발급

**요청주소:** `https://whooing.com/app_auth/access_token`

**파라미터:**

- **app_id** (required)
- **app_secret** (required)
- **token** (required)
- **pin** (required)

**응답:**

```json
{
	"code": 200,
	"token": "4ee880be438a0e5dc294ddabad07580361a55ab3",
	"token_secret": "1d51fcc7c53f4fefda07f68c73e208815c85d103",
	"user_id": 2399
}
```

###### 리소스 요청 (OAuth 1.0a)

아래 파라미터를 `=`와 `,`로 묶어 요청 Header의 **X-API-KEY**에 담습니다.

```bash
# signature = sha1(app_secret + '|' + token_secret)

curl -X GET "https://whooing.com/api/sections.json" \
  -H "X-API-KEY: app_id=45,token=7580361ddabad0a55ab34ee880be438a0e5dc294,signature=f0bd0e904282ea22a252b0a32d9638dc4c700e2,nounce=dc5584feec085bbda9a26e013f702c9b77692625,timestamp=1305875257"
```

---

##### Onetime PIN

써드파티앱에서 복잡한 인증 구현 없이, 사용자가 후잉 웹사이트에서 1회용 토큰을 발급받아 입력하면 바로 Access Token을 받을 수 있는 방식입니다.

**요청주소:** `https://whooing.com/app_auth/access_token_by_onetime`

**파라미터:**

- **app_id** (required)
- **app_secret** (required)
- **onetime_pin** (required): 유저가 입력한 1회용 토큰번호 (1분 이내 유효)

> 사용자에게 [후잉 > 계정 > 비밀번호 및 보안 > 1회용 인증생성]에서 토큰을 생성하도록 안내해야 합니다.

#### 3. 샘플코드

- [Javascript. whooing-api-test](https://whooing-api-test.vercel.app)

---

## 3. Rate Limiting

서버 안정성을 위해 API 호출 횟수에 제한이 있습니다.

### 일일 한도

응답 바디의 `rest_of_api` 필드로 오늘 남은 호출 가능 횟수를 확인할 수 있습니다.

| 유형 | 한도     |
| ---- | -------- |
| 일일 | 20,000회 |

### 단기 요청 제한

일일 한도와 별개로, 짧은 시간 내에 요청이 집중될 경우 WAF(방화벽)에 의해 일시적으로 차단될 수 있습니다. **분당 20회 이하**로 요청 속도를 유지하는 것을 권장합니다. 차단 시 `429 Too Many Requests`가 반환되며, 잠시 후 재시도하면 됩니다.

### AI Agent / MCP 이용 가이드

AI Agent나 MCP(Model Context Protocol) 도구에서 API를 사용할 때는 아래 사항을 반드시 준수해 주세요.

- **분당 20회 이하**로 요청 속도를 유지할 것
- **`429 Too Many Requests`** 응답을 받으면 일정시간 대기 후 재요청
- 불필요한 반복 요청을 자제하고 응답 결과를 적극적으로 캐싱할 것(sections, accounts 등 기반정보)

---

## 4. API Reference

모든 API는 개발자가 원하시는 대로 요청할 수 있습니다만, 염두에 두어야할 것이 몇 가지 있습니다.

### 필수 안내 사항

#### 리턴코드 안내

| 코드 | 내용                                                                                                                                                        |
| ---- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 200  | 모든 데이터가 정상적으로 반환됨(레코드가 없을 수 있음)                                                                                                      |
| 204  | 레코드가 없음                                                                                                                                               |
| 400  | 요청한 파라미터 중 누락되었거나 잘못되었을 수 있음. 개선하여야할 파라미터는 error_parameters에 배열로 반환됨                                                |
| 401  | 권한을 벗어난 요청                                                                                                                                          |
| 402  | 하루 요청횟수를 초과하였을 때. 업그레이드 안내 페이지로 이동시켜야함                                                                                        |
| 405  | 인증토큰이 만료되거나 잘못 되었을 때. ※ 언제든지 사용자가 후잉에서 인증토큰의 권한을 취소할 수 있으므로 항상 이 값을 확인하여 재인증을 요구하는 구조여야 함 |
| 500  | 서버측에서 발생한 에러. 잠시 후 다시 시도하면 됨                                                                                                            |

#### 요청횟수 제한

일일 요청 API는 개발시에도 이용할 때와 마찬가지로 유료결제 여부에 따라 제한이 됩니다. 만약 여러분이 엄청나게 좋은 앱을 개발하는데 후잉의 도움이 필요하시다면 support@whooing.com으로 연락을 주십시오.

가이드:

- 모든 API 응답의 `rest_of_api` 부분을 체크하여 사용자에게 남은 API 횟수를 항상 알려줍니다.
- 응답코드가 402로 온 경우에는 '하루 할당 API가 종료되었으니 프리미엄으로 업그레이드 하시거나 다음날 이용해 주시기 바랍니다'라는 메시지를 띄워주십시오.

#### 캐싱하여야할 API들

아래의 API들은 최초에 1회만 불러온 이후에 재요청없이 나머지 API 결과들과 결합(혹은 사전 조합)하여 사용하면 됩니다. 일종의 환경 정보라 생각하면 됩니다.

| API 종류 | 설명                                                         |
| -------- | ------------------------------------------------------------ |
| user     | 유저정보이며 유저속성이 바뀌지 않는한 유효                   |
| sections | 섹션리스트로 섹션이 바뀌지 않는한 유효                       |
| accounts | 항목리스트로, 섹션이 바뀌거나 항목변동이 생기지 않는 한 유효 |

#### 정확성보다는 효율

후잉은 사용자들에게 정확성보다는 속도와 효율적인 환경을 만들고자 더욱 노력합니다. API를 통해서 데이터를 수정/삭제할 때에 결과 code가 200으로 반환이 되었어도 수정이나 삭제가 되지 않을 수도 있습니다. 하지만 그러한 일은 0.1%의 확률로 일어날 것이므로, API를 요청하고 code가 200으로 반환되었으면 완료된 것으로 처리하여도 상관없습니다.

#### Rate Limit

요청수는 기본 하루 20000 건이지만, 짧은시간 안에 여러번의 요청이 있는경우 방화벽에 의해 차단될 수 있습니다. 다수의 API 호출이 필요한 경우 동시요청보다는 순차적으로 호출한 이후에 앱내에서 결과를 조합하는 방식으로 진행하십시오.

#### 조회 기간 제한

기간을 받는 보고서·거래내역 등 대부분의 API는 단일 요청에서 **최대 10년(120개월)**까지만 조회할 수 있습니다. 이를 초과하면 에러가 반환됩니다. 더 긴 기간을 조회할 경우 기간을 나눠 여러 번 요청하십시오.

---

### User - 사용자

#### GET user.json

유저에 관한 기본 정보를 요청합니다.

**Resource URL:** `https://whooing.com/api/user.json`

파라미터 없음.

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"user_id": 4,
		"username": "Helloman",
		"last_ip": "192.168.0.1",
		"last_login_timestamp": 1322448931,
		"created_timestamp": 1321448931,
		"modified_timestamp": 1321448931,
		"language": "ko",
		"level": "1",
		"expire": 1321448931,
		"timezone": "Asia/Seoul",
		"currency": "KRW",
		"country": "KR",
		"image_url": "https://static.whooing.com/profiles/p14.jpg",
		"mileage": 230
	}
}
```

#### PUT user.json

항목의 정보를 수정합니다. 수정이 되는 파라미터들만 전송하면 됩니다.

**Resource URL:** `https://whooing.com/api/user.json`

**파라미터:**

- username: 사용자의 별명 (예: 흥반장)
- country: 사용자의 국가 (예: KR)
- language: 사용자의 언어 (예: ko)
- timezone: 사용자의 타임존 (예: Asia/Seoul)
- currency: 사용자의 기본 통화단위 (예: KRW)

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"user_id": 4,
		"username": "Helloman",
		"language": "ko",
		"level": "1",
		"expire": 1321448931,
		"timezone": "Asia/Seoul",
		"currency": "KRW",
		"country": "KR"
	}
}
```

#### GET user_logs.json

유저의 로그리스트를 요청합니다.

**Resource URL:** `https://whooing.com/api/user_logs.json`

**파라미터:**

- max: id가 max보다 작은 로그로만 제한 (예: 1297)
- limit: 표시할 로그의 갯수 (예: 20)

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": [
		{
			"id": 483915,
			"contents": "`Def 234` 섹션을 수정",
			"datetime": 1618497070,
			"ip": "182.172.164.88",
			"segment0": "sections",
			"segment1": "",
			"writer": "user"
		}
	]
}
```

#### GET user_point_logs.json

유저의 포인트 로그리스트를 요청합니다.

**Resource URL:** `https://whooing.com/api/user_point_logs.json`

**파라미터:**

- max: point_id가 max보다 작은 포인트 로그로만 제한 (예: 1297)
- type: 포인트의 종류. 기본은 all, 추천인을 통한 가입 리스트는 affiliate (예: all)
- limit: 표시할 포인트 로그의 갯수 (예: 20)

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": [
		{
			"point_id": 144968,
			"datetime": 1621760410,
			"description": "일일 방문",
			"point": 50,
			"writer": "user"
		}
	]
}
```

---

### Sections - 섹션

#### GET sections.json

유저의 섹션 리스트를 요청합니다. isolation값이 y인 경우에는 섹션리스트에서 나타나지 않길 원하는 비자금일 경우도 있으므로 컨슈머의 목적에 따라 별도로 처리하여 줍니다.

**Resource URL:** `https://whooing.com/api/sections.json`

파라미터 없음.

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": [
		{
			"section_id": "s123",
			"title": "유동성 자산",
			"memo": "자주접근하는 자산만 관리",
			"currency": "KRW",
			"isolation": "n",
			"total_assets": 2982799.0,
			"total_liabilities": 23910.0,
			"skin_id": 0,
			"decimal_places": 2,
			"date_format": "YMD",
			"webhook_token": "xxxx-xxxx-xxxx-xxxx-xxxx"
		}
	]
}
```

#### GET sections/:section_id.json

특정 섹션의 정보를 요청합니다.

**Resource URL:** `https://whooing.com/api/sections/:section_id.json`

**파라미터:**

- **:section_id** (required): 섹션의 고유번호 (예: s122)

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"section_id": "s123",
		"title": "유동성 자산",
		"currency": "KRW",
		"isolation": "n",
		"total_assets": 2982799.0,
		"total_liabilities": 23910.0,
		"skin_id": 1,
		"decimal_places": 2,
		"date_format": "YMD",
		"webhook_token": "xxxx-xxxx-xxxx-xxxx-xxxx",
		"ui": {
			"billOrder": "asc",
			"insertSlot": "2",
			"width": "normal"
		}
	}
}
```

#### POST sections.json

섹션을 추가합니다. 각 사용자당 섹션은 최대 9개까지 추가할 수 있습니다.

**Resource URL:** `https://whooing.com/api/sections.json`

**파라미터:**

- **title** (required): 섹션의 제목. 1~30 글자
- **currency** (required): 섹션의 통화단위 (예: USD)
- memo: 섹션 보조설명. 0~80 글자
- skin_id: 등록된 스킨의 고유번호
- decimal_places: 소수점 이하로 몇자리 까지 표시할 것인지. 0~3
- date_format: 날짜를 표시하는 방법 (예: YMD)
- start_year: 섹션 생성시 항목의 open_date를 설정하는 시작연도. 2000년~현재연도 범위 (예: 2020)
- template_id: 초기 항목 구성 템플릿 ID

#### PUT sections/:section_id.json

특정 섹션에 관한 정보를 수정합니다.

**Resource URL:** `https://whooing.com/api/sections/:section_id.json`

**파라미터:**

- **:section_id** (required): 섹션의 고유번호
- title: 섹션의 제목
- currency: 섹션의 통화단위
- memo: 섹션 보조설명
- skin_id: 스킨의 고유번호
- decimal_places: 소수점 이하 자릿수. 0~3
- date_format: 날짜를 표시하는 방법
- `ui[{key}]`: UI 설정 개별 키-값 업데이트 (PHP 배열 표기). `title` 없이 `ui[*]`만 보내도 동작. 예: `ui[budgetLong]=y`
- `ui`: UI 설정을 JSON 문자열로 한번에 업데이트. 예: `ui={"budgetLong":"y","mainIndex":"main/insert"}`. 알려진 키: `budgetLong`(장기목표 표시 y/n), `mainIndex`(기본 화면), `insertMethod`(입력 방식), `cashflowSalesAccs`(현금흐름 영업계정 목록 JSON 배열)

#### DELETE sections/:section_id.json

특정 섹션을 삭제합니다.

**Resource URL:** `https://whooing.com/api/sections/:section_id.json`

**파라미터:**

- **:section_id** (required): 섹션의 고유번호. 복수의 섹션들을 한꺼번에 삭제하려면 콤마(,)로 이어붙인 문자열을 전송 (예: s129,s118,s199)

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4948
}
```

#### GET sections/default.json

기본섹션을 조회합니다. 컨슈머의 초기화면에 섹션 리스트를 제공하지 않는 경우에는 기본섹션을 바로 출력하기 위해 요청합니다.

**Resource URL:** `https://whooing.com/api/sections/default.json`

#### PUT sections/sort.json

섹션의 순서를 변경합니다.

**Resource URL:** `https://whooing.com/api/sections/sort.json`

**파라미터:**

- **section_ids** (required): 정렬하려는 순서에 맞추어 전체 섹션 고유번호들을 콤마(,)로 구분한 문자열 (예: s99,s72,s78,s52)

```json
{
	"code": 200,
	"rest_of_api": 2388,
	"results": ["s99", "s72", "s78", "s52"]
}
```

---

### Accounts - 항목

항목에 관한 정보는 후잉 전반에 걸쳐 매우 빈번하게 호출되며 쉽게 수정되지 않는 정보이기 때문에, 한 번 요청하면 그 값을 컨슈머 내부에 저장을 하는 것을 추천합니다.

계정 타입:

- `assets`: 자산
- `liabilities`: 부채
- `capital`: 순자산
- `expenses`: 비용
- `income`: 수익

#### GET accounts.json

모든 항목리스트를 요청합니다.

**Resource URL:** `https://whooing.com/api/accounts.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호 (예: s199)
- start_date: 특정 날짜 이후에만 보여야만 하는 항목으로 제한 (예: 20130101)
- end_date: 특정 날짜 이전에만 보여야만 하는 항목으로 제한 (예: 20131231)

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"assets": [
			{
				"account_id": "x1",
				"type": "group",
				"title": "유동자산",
				"memo": "바로쓸 수 있는 것들",
				"open_date": 20090511,
				"close_date": 20160101,
				"category": ""
			},
			{
				"account_id": "x2",
				"type": "account",
				"title": "현금",
				"memo": "내 지갑 및 서랍에 있는 돈",
				"open_date": 20090511,
				"close_date": 20160101,
				"category": "normal"
			}
		],
		"liabilities": [
			{
				"account_id": "x10",
				"type": "account",
				"title": "신한카드",
				"category": "creditcard",
				"opt_use_date": "p1",
				"opt_pay_date": 25,
				"opt_pay_account_id": "x1"
			}
		],
		"capital": [],
		"income": [],
		"expenses": []
	}
}
```

#### GET accounts/:account.json

계정의 항목리스트를 요청합니다.

**Resource URL:** `https://whooing.com/api/accounts/:account.json`

**파라미터:**

- **:account** (required): 조회할 계정 (assets/liabilities/capital/expenses/income)
- **section_id** (required): 섹션의 고유번호
- start_date: 특정 날짜 이후 항목으로 제한
- end_date: 특정 날짜 이전 항목으로 제한

#### GET accounts/:account/:account_id.json

항목의 정보를 요청합니다.

**Resource URL:** `https://whooing.com/api/accounts/:account/:account_id.json`

**파라미터:**

- **:account** (required): 조회할 계정
- **:account_id** (required): 항목의 고유번호 (예: x2)
- **section_id** (required): 섹션의 고유번호

#### GET accounts/:account/:account_id/exists.json

거래내역에서 해당 항목을 이용한 거래가 있는지 여부를 조사합니다. 주로 항목을 삭제하기 전이나 종료하기 전에 확인 겸 사용합니다.

**Resource URL:** `https://whooing.com/api/accounts/:account/:account_id/exists.json`

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"count": 0,
		"minDate": 20100102,
		"maxDate": 20111232,
		"balance": 0,
		"last_one": "n",
		"close_date": 20120812
	}
}
```

#### POST accounts/:account.json

항목을 추가합니다.

**Resource URL:** `https://whooing.com/api/accounts/:account.json`

**파라미터:**

- **:account** (required): 대상 계정
- **section_id** (required): 섹션의 고유번호
- title: 항목의 이름. 0~30 글자
- **type** (required): 그룹인지 항목인지 구분 (예: account)
- **open_date** (required): 항목의 사용이 시작되는 날짜 (예: 20010101)
- **close_date** (required): 항목의 사용이 종료되는 날짜. 29991231의 경우 종료되지 않는 항목
- memo: 항목의 설명. 0~255글자
- category: 항목의 종류 (normal/client/creditcard/checkcard/steady/floating)
- opt_use_date: 신용카드인 경우 사용기간의 시작일 (pp1 ~ p31)
- opt_pay_date: 신용카드인 경우 대금결제일 (1~31)
- opt_pay_account_id: 신용카드인 경우 대금결제하는 상대 자산항목

#### PUT accounts/:account/:account_id.json

항목의 정보를 수정합니다. 전체 필드를 전달해야합니다.

**Resource URL:** `https://whooing.com/api/accounts/:account/:account_id.json`

#### DELETE accounts/:account/:account_id/:section_id.json

특정 항목을 삭제합니다. 항목을 삭제하기 전에 exists API로 거래 존재 여부를 반드시 검사하여야 합니다.

**Resource URL:** `https://whooing.com/api/accounts/:account/:account_id/:section_id.json`

**파라미터:**

- **:account** (required): 대상 계정
- **:account_id** (required): 항목의 고유번호. 복수의 항목삭제는 지원하지 않음
- **:section_id** (required): 섹션의 고유번호

#### PUT accounts/:account/sort.json

항목의 순서를 변경합니다. 비활성화 되어 있는 모든 항목의 고유번호도 account_ids에 포함을 하여야 합니다.

**Resource URL:** `https://whooing.com/api/accounts/:account/sort.json`

**파라미터:**

- **:account** (required): 대상 계정
- **section_id** (required): 섹션의 고유번호
- **account_ids** (required): 정렬하려는 순서에 맞추어 전체 항목 고유번호들을 콤마(,)로 구분한 문자열 (예: x2,x4,x3,x5)

---

### Entries - 거래내역

#### GET entries.json

조건에 맞는 거래내역을 조회합니다. l_account_id나 r_account_id가 x0으로 반환되는 것들은 삭제되었지만 기록상 남겨둔 것입니다.

**Resource URL:** `https://whooing.com/api/entries.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- start_date: 조회 시작 날짜 (예: 20110123). 생략 시 한 달 전 날짜로 자동 설정
- end_date: 조회 종료 날짜. 조회시작~종료까지는 최대 1년 (예: 20110603). 생략 시 오늘 날짜로 자동 설정. start_date와 end_date 중 하나만 생략하면 에러
- max: 특정 entry_date 이전으로만 제한 (예: 20110203.0034)
- limit: 조회할 거래의 수. 기본 20, 최대 100 (후잉 공식 앱은 최대 5000)
- account: 왼쪽이나 오른쪽에 해당 계정이 나타나는 경우로 제한
- account_id: 왼쪽이나 오른쪽에 해당 계정과 항목이 나타나는 경우로 제한 (account 필수)
- l_account: 왼쪽에 해당 계정이 나타나는 경우로 제한
- l_account_id: 왼쪽에 해당 계정과 항목이 나타나는 경우로 제한 (l_account 필수)
- r_account: 오른쪽에 해당 계정이 나타나는 경우로 제한
- r_account_id: 오른쪽에 해당 계정과 항목이 나타나는 경우로 제한 (r_account 필수)
- item: 아이템 or 거래처 검색. 동작 방식은 아래 **검색 상세** 참고
- money_from: 특정 금액부터 시작하는 거래로 제한
- money_to: 특정 금액까지의 거래로 제한
- memo: 메모 검색. 동작 방식은 아래 **검색 상세** 참고
- sort_column: 정렬 기준 컬럼. entry_date(기본값), item, money, total, l_account_id, r_account_id
- sort_order: 정렬 순서. desc(기본값), asc

**검색 상세:**

**item 검색**

- 기본은 정확 일치 (예: `item=커피` → item이 정확히 '커피'인 것만)
- `*`를 와일드카드로 사용 가능
    - `커피*` → '커피'로 시작하는 것
    - `*커피` → '커피'로 끝나는 것
    - `*커피*` → '커피'를 포함하는 것
- 괄호로 detail(괄호메모) 필터링 가능: `item=아이템(detail검색어)`
    - detail은 부분 포함 검색
    - 공백으로 구분하면 AND 조건 (예: `아이템(서울 강남)` → '서울' AND '강남' 포함)
    - `!` prefix로 제외 검색 (예: `아이템(서울 !강남)` → '서울' 포함 AND '강남' 미포함)
    - `()`만 쓰면 detail이 비어있는 항목 검색 (예: `아이템()`)

**memo 검색**

- 부분 포함 검색 (기본)
- 공백으로 구분하면 AND 조건 (예: `카페 라떼` → '카페' AND '라떼' 모두 포함)
- `!` prefix로 제외 검색 (예: `카페 !라떼` → '카페' 포함 AND '라떼' 미포함)
- 영문은 대소문자 구분 없음

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"reports": [],
		"rows": [
			{
				"entry_id": 1352827,
				"entry_date": 20110817.0001,
				"l_account": "expenses",
				"l_account_id": "x20",
				"r_account": "assets",
				"r_account_id": "x4",
				"item": "후원(과장학금)",
				"money": 10000,
				"total": 840721.99,
				"memo": "",
				"app_id": 0,
				"attachments": [
					{
						"uuid": "810cbdb1b-7486jvk57",
						"src": "https://static.whooing.com/get/810cbdb1b-7486jvk57",
						"filename": "example.jpg",
						"mimeType": "image/jpeg",
						"size": 28098
					}
				]
			}
		]
	}
}
```

#### GET entries/:entry_id.json

거래 정보를 조회합니다.

**Resource URL:** `https://whooing.com/api/entries/:entry_id.json`

**파라미터:**

- **:entry_id** (required): 조회할 거래의 고유번호 (예: 1352827)
- **section_id** (required): 섹션의 고유번호

#### POST entries.json

거래를 추가합니다. 단건 또는 최대 300개의 거래를 일괄 입력할 수 있습니다.

**Resource URL:** `https://whooing.com/api/entries.json`

**파라미터 (단건):**

- **section_id** (required): 섹션의 고유번호
- entry_date: 거래가 일어난 날짜 (예: 20110812)
- **l_account** (required): 왼쪽의 계정 (예: expenses)
- **l_account_id** (required): 왼쪽의 항목 고유번호 (예: x20)
- **r_account** (required): 오른쪽의 계정 (예: assets)
- **r_account_id** (required): 오른쪽의 항목 고유번호 (예: x4)
- item: 아이템 or 거래처. 괄호메모나 반복·분할 명령어 포함 가능 (아래 **item 명령어** 참고). 반복·분할 명령어는 POST(입력) 전용이며 PUT(수정)에서는 동작하지 않는다.
- **money** (required): 거래액 (예: 10000)
- memo: 거래에 들어가는 보충 메모
- attachment_ids: 파일 uuid를 콤마로 구분한 문자열

**item 명령어 — 반복·분할 입력:**

거래를 여러 달에 걸쳐 자동 입력할 때 item 뒤에 명령어를 붙인다. 날짜는 입력일 기준으로 매월 자동 생성된다.

| 기호 | 형식        | 동작                           | 예시 (30만원)              | 총액    |
| ---- | ----------- | ------------------------------ | -------------------------- | ------- |
| `//` | `아이템//n` | 금액을 n등분해 n개월 분할 기록 | `노트북//3` → 10만원×3건   | 30만원  |
| `**` | `아이템**n` | 동일 금액을 n개월 반복 기록    | `구독료**12` → 30만원×12건 | 360만원 |

- `//`는 할부처럼 총액을 나눠 기록할 때 사용. `**`는 월정액처럼 매달 같은 금액이 나갈 때 사용.
- 할부 수수료율이 있으면 개월 수 뒤에 `(n%)`를 붙인다: `노트북//3(28%)`

**파라미터 (일괄입력):**

- **section_id** (required): 섹션의 고유번호
- data_type: entries를 문자열로 만든 방식. 기본 json
- **entries** (required): 아래의 배열을 json 문자열로 직렬화, 최대 300개

```json
[
	{
		"entry_date": 20110812,
		"l_account": "expenses",
		"l_account_id": "x20",
		"r_account": "assets",
		"r_account_id": "x4",
		"item": "후원(과장학금)**2", // **2: 동일 금액 2개월 반복
		"money": 10000,
		"memo": "오늘도 어김없이 빠져나갔다"
	}
]
```

#### PUT entries/:entry_id.json

거래를 수정합니다. 수정이 필요한 필드만 전송합니다.

**Resource URL:** `https://whooing.com/api/entries/:entry_id.json`

**파라미터:**

- **:entry_id** (required): 수정할 거래의 고유번호
- **section_id** (required): 섹션의 고유번호
- entry_date, l_account, l_account_id, r_account, r_account_id, item, money, memo, attachment_ids (모두 선택)

#### PUT entries/:entry_ids/:section_id.json

복수개의 거래를 수정합니다.

**Resource URL:** `https://whooing.com/api/entries/:entry_ids/:section_id.json`

**파라미터:**

- **:entry_ids** (required): 수정할 거래의 고유번호들. 콤마(,)로 이은 문자열. 최대 100개 (예: 1352827,1352828,1352829)
- **:section_id** (required): 섹션의 고유번호
- entry_date, l_account, l_account_id, r_account, r_account_id, item, money, memo (모두 선택)
- ~~attachment_ids~~: 복수의 거래에서는 지원하지 않습니다

#### DELETE entries/:entry_ids/:section_id.json

특정 거래를 삭제합니다.

**Resource URL:** `https://whooing.com/api/entries/:entry_ids/:section_id.json`

**파라미터:**

- **:entry_ids** (required): 거래의 고유번호. 복수개의 경우에는 콤마로 구분. 최대 100개
- **section_id** (required): 섹션의 고유번호

#### GET entries/latest.json

최근에 입력된 거래내역을 조회합니다.

**Resource URL:** `https://whooing.com/api/entries/latest.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- max: 특정 entry_id 이전으로만 제한
- limit: 조회할 거래의 수. 기본 20

#### GET entries/latest_items.json

최근 60일 이내에 입력된 거래내역을 중복없이 불러옵니다. 주로 거래입력 Suggest 기능에 활용됩니다.

**Resource URL:** `https://whooing.com/api/entries/latest_items.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호

#### GET entries/flow_of_account.json

특정 계정과 모든 계정/항목의 상대적인 증가/감소를 조회합니다. GET entries를 호출할 때의 동일한 파라미터를 사용합니다.

**Resource URL:** `https://whooing.com/api/entries/flow_of_account.json`

#### GET entries/flow_of_account_id.json

특정 항목과 모든 계정/항목의 상대적인 증가/감소를 조회합니다.

**Resource URL:** `https://whooing.com/api/entries/flow_of_account_id.json`

#### GET entries/changes_of_account_id.json

특정 항목의 일일 변동내역을 표시합니다.

**Resource URL:** `https://whooing.com/api/entries/changes_of_account_id.json`

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"aggregate": {
			"in": 1010002,
			"out": 298933
		},
		"rows_type": "day",
		"rows": [
			{ "date": "20110616", "money": 0 },
			{ "date": "20110617", "money": 230 }
		]
	}
}
```

#### GET entries/changes_of_client.json

특정 거래처의 일일 변동내역을 표시합니다.

**Resource URL:** `https://whooing.com/api/entries/changes_of_client.json`

#### GET entries/changes_of_item.json

특정 아이템의 일일 발생내역을 표시합니다.

**Resource URL:** `https://whooing.com/api/entries/changes_of_item.json`

#### GET entries/account_ids_of_account.json

계정의 항목별 금액을 조회합니다.

**Resource URL:** `https://whooing.com/api/entries/account_ids_of_account.json`

#### GET entries/clients_of_account_id.json

항목의 거래처별 금액을 조회합니다.

**Resource URL:** `https://whooing.com/api/entries/clients_of_account_id.json`

#### GET entries/items_of_account_id.json

항목의 아이템별 금액을 조회합니다.

**Resource URL:** `https://whooing.com/api/entries/items_of_account_id.json`

#### POST entries/outside.json

외부데이터들을 파싱하고 총 인식한 건수를 반환합니다. code가 400으로 반환되면 지원하지 않는 형식입니다.

**Resource URL:** `https://whooing.com/api/entries/outside.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- **rows** (required): 외부데이터 내용 (예: 은행 SMS 문자 내용)

#### POST entries/outside_report.json

방금 외부입력을 시도했다가 인식되지 않은 것에 대해서 소스를 보고합니다.

**Resource URL:** `https://whooing.com/api/entries/outside_report.json`

**파라미터:**

- **source** (required): 데이터의 소스명

---

### Frequent Items - 자주입력 거래들

#### GET frequent_items.json

모든 자주입력거래 리스트를 요청합니다.

**Resource URL:** `https://whooing.com/api/frequent_items.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"slot1": [
			{
				"item_id": "f4",
				"item": "생필품",
				"money": 40000,
				"l_account": "expenses",
				"l_account_id": "x12",
				"r_account": "assets",
				"r_account_id": "x5"
			}
		]
	}
}
```

#### GET frequent_items/:slot.json

특정 슬롯의 자주입력거래 리스트를 요청합니다.

**Resource URL:** `https://whooing.com/api/frequent_items/:slot.json`

**파라미터:**

- **:slot** (required): 특정 슬롯. slot1~slot3까지 존재함. 어떤 슬롯까지 사용자가 활용하는지는 `GET sections/:section_id.json > ui > insertSlot`에 명시
- **section_id** (required): 섹션의 고유번호

#### GET frequent_items/:slot/:item_id.json

특정 자주입력거래의 정보를 요청합니다.

**Resource URL:** `https://whooing.com/api/frequent_items/:slot/:item_id.json`

**파라미터:**

- **:slot** (required): 특정 슬롯 (예: slot1)
- **:item_id** (required): 자주입력거래 고유번호 (예: f4)
- **section_id** (required): 섹션의 고유번호

#### POST frequent_items/:slot.json

자주입력거래를 새로 등록합니다. 해당 슬롯의 가장 마지막에 추가됩니다.

**Resource URL:** `https://whooing.com/api/frequent_items/:slot.json`

**파라미터:**

- **:slot** (required): 특정 슬롯
- **section_id** (required): 섹션의 고유번호
- **item** (required): 아이템 (예: 생필품)
- money: 자주나오는 금액
- **l_account** (required): 왼쪽의 계정
- l_account_id: 왼쪽의 항목 고유번호
- **r_account** (required): 오른쪽의 계정
- r_account_id: 오른쪽의 항목 고유번호

#### PUT frequent_items/:slot/:item_id.json

자주입력거래의 정보를 수정합니다. 수정되는 파라미터만 전송하여도 됩니다. 슬롯 이동은 "기존 슬롯에서 삭제" → "다른 슬롯에 추가" → "다른 슬롯에서 순서 재설정" 3단계로 처리합니다.

**Resource URL:** `https://whooing.com/api/frequent_items/:slot/:item_id.json`

#### DELETE frequent_items/:slot/:item_id/:section_id.json

특정 자주입력거래를 삭제합니다.

**Resource URL:** `https://whooing.com/api/frequent_items/:slot/:item_id/:section_id.json`

**파라미터:**

- **:slot** (required): 특정 슬롯
- **:item_id** (required): 자주입력거래의 고유번호. 복수 삭제시 콤마(,)로 구분 (예: f28,f12,f15)
- **:section_id** (required): 섹션의 고유번호

#### PUT frequent_items/:slot/sort.json

자주입력거래의 순서를 변경합니다.

**Resource URL:** `https://whooing.com/api/frequent_items/:slot/sort.json`

**파라미터:**

- **:slot** (required): 특정 슬롯
- **section_id** (required): 섹션의 고유번호
- **item_ids** (required): 자주입력거래의 고유번호들을 콤마(,)로 구분한 문자열 (예: f21,f22,f23,f11)

---

### Monthly Items - 매월입력 거래들

#### GET monthly_items.json

모든 월별입력거래 리스트를 요청합니다. count는 빨리 결제를 진행하여야 할 갯수입니다.

**Resource URL:** `https://whooing.com/api/monthly_items.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"count": 0,
		"slot1": [
			{
				"item_id": "m4",
				"item": "통신비",
				"money": 79200,
				"l_account": "expenses",
				"l_account_id": "x12",
				"r_account": "assets",
				"r_account_id": "x5",
				"skip_holiday": "after",
				"pay_date": 27,
				"due_date": "20120327",
				"d_day": 1,
				"paid_date": "20120227"
			}
		]
	}
}
```

#### GET monthly_items/slot1.json

특정 슬롯의 월별입력거래 리스트를 요청합니다.

**Resource URL:** `https://whooing.com/api/monthly_items/slot1.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호

#### GET monthly_items/slot1/:item_id.json

특정 월별입력거래의 정보를 요청합니다.

**Resource URL:** `https://whooing.com/api/monthly_items/slot1/:item_id.json`

**파라미터:**

- **:item_id** (required): 월별입력거래 고유번호 (예: m4)
- **section_id** (required): 섹션의 고유번호

#### POST monthly_items/slot1.json

월별입력거래를 새로 등록합니다. 해당 슬롯의 가장 마지막에 추가됩니다.

**Resource URL:** `https://whooing.com/api/monthly_items/slot1.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- **item** (required): 아이템 (예: 생필품)
- money: 금액
- l_account: 왼쪽의 계정
- l_account_id: 왼쪽의 항목 고유번호
- r_account: 오른쪽의 계정
- r_account_id: 오른쪽의 항목 고유번호
- **pay_date** (required): 입력해야할 일자 (예: 27)
- skip_holiday: 휴일 건너뛰기 여부. none, before, after 중 하나 (예: after)

#### PUT monthly_items/slot1/:item_id.json

월별입력거래의 정보를 수정합니다. 수정되는 파라미터만 전송하여도 됩니다.

**Resource URL:** `https://whooing.com/api/monthly_items/slot1/:item_id.json`

**파라미터:**

- **:item_id** (required): 월별입력거래의 고유번호 (예: m4)
- **section_id** (required): 섹션의 고유번호
- item, money, l_account, l_account_id, r_account, r_account_id, **pay_date**, skip_holiday (모두 선택)

#### DELETE monthly_items/slot1/:item_id/:section_id.json

특정 월별입력거래를 삭제합니다.

**Resource URL:** `https://whooing.com/api/monthly_items/slot1/:item_id/:section_id.json`

**파라미터:**

- **:item_id** (required): 월별입력거래의 고유번호. 복수 삭제시 콤마(,)로 구분 (예: m28,m12,m15)
- **:section_id** (required): 섹션의 고유번호

---

### Budget - 예산

#### 예산 구조 개요

후잉의 예산은 세 레이어로 구성됩니다:

1. **budget_goal (장기목표 설정)**: 목표 기간(base_ym~goal_ym), 목표 자산(goal_money), 연간 수입·지출 예산, 월별 배분 비율을 정의합니다. 섹션 하나에 하나의 설정만 존재합니다.
2. **goal (월별 자본 목표)**: budget_goal에서 파생된 월별 자본 도달 목표값입니다. 산 차트의 목표 점선으로 표시됩니다.
3. **budget (월별 예산)**: 계정(expenses/income)별, 항목(account_id)별 월 예산 금액입니다.

**budgetLong 연동 여부** (`sections.list`의 `ui.budgetLong` 값):

- `"y"`: 장기목표 연동 활성. 월별 예산(budget)을 수정하면 **해당 월부터 goal_ym까지의 모든 goal 금액이 차이만큼 자동 조정**됩니다. budget_goal의 goal_money도 함께 갱신됩니다.
- `"n"`: 장기목표 미사용. 월별 예산만 독립적으로 관리됩니다.

---

#### GET budget/:account.json

해당 계정의 기간내 월 단위 예산 대비 실적을 조회합니다. budget은 꼭 :account를 동반하여야 합니다.

**Resource URL:** `https://whooing.com/api/budget/:account.json`

**파라미터:**

- **:account** (required): 조회할 계정. expenses, income 중 택1
- **section_id** (required): 섹션의 고유번호
- start_date: 조회 시작 년월 (예: 201101). 생략 시 한 달 전 날짜로 자동 설정
- end_date: 조회 종료 년월 (예: 201106). 생략 시 오늘 날짜로 자동 설정. start_date와 end_date 중 하나만 생략하면 에러

**응답 필드 설명:**

- `aggregate.total`: 전체 기간 합산 예산/실적/잔여
- `aggregate.total_steady`: 고정 항목(steady) 합산
- `aggregate.total_floating`: 변동 항목(floating) 합산
- `aggregate.accounts`: 항목별 예산/실적/잔여 배열
- `aggregate.misc.standard`: 오늘 날짜 기준으로 예산을 선형 배분했을 때의 기준 집행액 (오늘까지 이 정도는 써야 정상 페이스)
- `aggregate.misc.possibility`: 예산 달성 가능성 지수 (0~100). 100에 가까울수록 예산 내 달성 가능성 높음
- `aggregate.misc.daily_remains`: 오늘 기준 일일 사용 가능 잔여액
- `aggregate.misc.weekly_remains`: 오늘 기준 주간 사용 가능 잔여액
- `rows`: 월별 상세 내역

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"aggregate": {
			"total": {
				"budget": 9929000,
				"money": 2399000,
				"remains": 7530000
			},
			"total_steady": {
				"budget": 2983023,
				"money": 182000,
				"remains": 2304000
			},
			"total_floating": {
				"budget": 6945977,
				"money": 234000,
				"remains": 44320023
			},
			"accounts": [
				{
					"account_id": "x12",
					"budget": 120000,
					"money": 234002,
					"remains": -59203
				}
			],
			"misc": {
				"standard": 2939,
				"possibility": 29.239892,
				"daily_remains": 23983,
				"weekly_remains": 82983,
				"today": { "budget": 2300, "money": 300, "remains": 2000 }
			}
		},
		"rows_type": "month",
		"rows": {
			"201106": {
				"date": "201106",
				"total": {},
				"accounts": [],
				"misc": {}
			}
		}
	}
}
```

#### PUT budget/:account.json

특정 월의 계정 항목별 예산을 수정합니다. 변경할 항목만 전송하면 됩니다.

**⚠️ 주의 (budgetLong 연동):** `ui.budgetLong="y"` 섹션에서는 예산 총액이 변경되면 해당 월부터 장기목표 종료월(goal_ym)까지의 모든 월별 자본 목표(goal)가 자동으로 조정됩니다. 예산 수정 전 사용자에게 이 사이드이펙트를 반드시 안내하세요.

**Resource URL:** `https://whooing.com/api/budget/:account.json`

**파라미터:**

- **:account** (required): 조회할 계정. expenses, income 중 택1
- **section_id** (required): 섹션의 고유번호
- **target_ym** (required): 수정 년월 (예: 201106)
- x12, x13, ...: 항목 ID를 키로, 예산 금액을 값으로 전송 (변경할 항목만 전송 가능)

#### PUT budget/:account/basic_total.json

장기목표 설정 후 연간 예산을 월별로 배분할 때 사용합니다. 최소 1년치 이상의 기간, 1~12월 12개 파라미터를 모두 전송해야 합니다.

**Resource URL:** `https://whooing.com/api/budget/:account/basic_total.json`

**파라미터:**

- **:account** (required): 조회할 계정
- **section_id** (required): 섹션의 고유번호
- **start_date** (required): 수정 시작 년월
- **end_date** (required): 수정 종료 년월
- **1~12** (required): 각 월 번호를 키로, 해당 월 예산 총액을 값으로 전송 (12개 모두 필수)

#### DELETE budget/:account.json

해당 계정의 기간내 예산을 리셋합니다.

**Resource URL:** `https://whooing.com/api/budget/:account.json`

**파라미터:**

- **:account** (required): 조회할 계정
- **section_id** (required): 섹션의 고유번호
- **start_date** (required): 삭제 시작 년월
- **end_date** (required): 삭제 종료 년월

---

### Budget Goal - 장기 예산목표

장기목표의 기간, 목표 자산, 연간 예산, 월별 배분 비율을 관리합니다. 섹션당 하나의 설정만 존재합니다. 설정을 저장하면 goal_ym 이후의 월별 자본 목표(goal)가 자동 삭제됩니다.

#### GET budget_goal.json

장기목표 설정을 조회합니다. 설정이 없으면 `set_id=0`의 기본값을 반환합니다.

**Resource URL:** `https://whooing.com/api/budget_goal.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호

**응답 필드 설명:**

- `set_id`: 0이면 설정 없음
- `base_ym`: 장기목표 시작 년월 (YYYYMM)
- `goal_ym`: 장기목표 종료 년월 (YYYYMM)
- `base_money`: 시작 시점의 자산 금액
- `goal_money`: 목표 자산 금액
- `base_income`: 연간 수입 예산
- `base_expenses`: 연간 지출 예산
- `each_months`: 월별 배분 비율 배열 `[[수입 12개월], [지출 12개월]]`
- `split_type`: 배분 방식 (`auto`=과거 데이터 기반 자동, `equal`=균등, `manual`=직접입력)

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"set_id": 1,
		"last_modified": 1711234567,
		"base_ym": 202601,
		"goal_ym": 202612,
		"base_money": 10000000,
		"goal_money": 50000000,
		"base_income": 36000000,
		"base_expenses": 24000000,
		"each_months": [[300, 280, ...], [200, 210, ...]],
		"split_type": "auto"
	}
}
```

#### PUT budget_goal.json

장기목표 설정을 저장합니다. 설정이 없으면 새로 생성하고, 있으면 업데이트합니다. `goal_ym` 이후의 월별 자본 목표(goal)는 자동으로 삭제됩니다.

**Resource URL:** `https://whooing.com/api/budget_goal.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- **base_ym** (required): 장기목표 시작 년월 (예: 202601)
- **goal_ym** (required): 장기목표 종료 년월. 현재월 기준 최소 1년 이후, 최대 10년 이내
- **goal_money** (required): 목표 자산 금액
- base_money: 시작 시점 자산 금액
- base_income: 연간 수입 예산
- base_expenses: 연간 지출 예산
- each_months: 월별 배분 비율 JSON 문자열 `[[수입 12개월 배열], [지출 12개월 배열]]`. 생략 시 균등 배분
- split_type: 배분 방식. `auto`, `equal`, `manual` 중 택1 (기본값: `auto`)

#### DELETE budget_goal.json

장기목표 설정과 해당 섹션의 모든 월별 목표(goal) 및 예산(budget) 데이터를 초기화합니다. **되돌릴 수 없습니다.**

**Resource URL:** `https://whooing.com/api/budget_goal.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호

---

### Bill - 신용카드

#### GET bill.json

해당 기간내의 월별 카드 청구내역을 조회합니다.

**Resource URL:** `https://whooing.com/api/bill.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- start_date: 조회 시작 날짜 (예: 201101). 생략 시 한 달 전 날짜로 자동 설정
- end_date: 조회 종료 날짜 (예: 201106). 생략 시 오늘 날짜로 자동 설정. start_date와 end_date 중 하나만 생략하면 에러

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"aggregate": {
			"total": 5336000,
			"accounts": [
				{
					"account_id": "x10",
					"money": 2938000,
					"start_use_date": 20110420,
					"end_use_date": 20110519,
					"pay_date": 25,
					"pay_account_id": "X5"
				}
			]
		},
		"rows_type": "month",
		"rows": {
			"201106": {
				"total": 126000,
				"accounts": []
			}
		}
	}
}
```

#### GET bill/:account_id.json

해당 기간내의 특정 항목의 월별 카드 청구내역을 조회합니다.

**Resource URL:** `https://whooing.com/api/bill/:account_id.json`

**파라미터:**

- **:account_id** (required): 항목의 고유번호 (예: x11)
- **section_id** (required): 섹션의 고유번호
- start_date, end_date: 생략 시 각각 한 달 전/오늘로 자동 설정. 하나만 생략하면 에러

---

### Checkcard - 체크카드

#### GET checkcard.json

해당 기간내의 월별 체크카드 사용내역을 조회합니다.

**Resource URL:** `https://whooing.com/api/checkcard.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- start_date: 조회 시작 날짜 (예: 201101). 생략 시 한 달 전 날짜로 자동 설정
- end_date: 조회 종료 날짜 (예: 201106). 생략 시 오늘 날짜로 자동 설정. start_date와 end_date 중 하나만 생략하면 에러

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"aggregate": {
			"total": 5336000,
			"accounts": [
				{
					"account_id": "x10",
					"money": 2938000,
					"pay_account_id": "X5"
				}
			]
		},
		"rows_type": "month",
		"rows": {
			"201106": { "total": 126000, "accounts": [] }
		}
	}
}
```

#### GET checkcard/:account_id.json

해당 기간내의 특정 항목의 월별 체크카드 사용내역을 조회합니다.

**Resource URL:** `https://whooing.com/api/checkcard/:account_id.json`

**파라미터:**

- **:account_id** (required): 항목의 고유번호 (예: x11)
- **section_id** (required): 섹션의 고유번호
- start_date, end_date: 생략 시 각각 한 달 전/오늘로 자동 설정. 하나만 생략하면 에러

---

### In/Out - 자금증감

#### GET in_out.json

해당 기간내의 절대증감량을 조회합니다.

**Resource URL:** `https://whooing.com/api/in_out.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- start_date: 조회 시작 날짜 (예: 20110304). 생략 시 한 달 전 날짜로 자동 설정
- end_date: 조회 종료 날짜 (예: 20110604). 생략 시 오늘 날짜로 자동 설정. start_date와 end_date 중 하나만 생략하면 에러

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"assets": {
			"total": { "in": 1066000, "out": 421833, "margin": 644167 },
			"accounts": [
				{
					"account_id": "x1",
					"in": 1066000,
					"out": 421833,
					"margin": 644167
				}
			]
		},
		"liabilities": {}
	}
}
```

#### GET in_out/:account.json

해당 기간 내의 자산이나 부채 중 한 계정만을 조회합니다.

**Resource URL:** `https://whooing.com/api/in_out/:account.json`

**파라미터:**

- **:account** (required): 조회할 자산이나 부채 중 택1 (예: assets)
- **section_id** (required), start_date, end_date: 생략 시 각각 한 달 전/오늘로 자동 설정. 하나만 생략하면 에러

#### GET in_out/:account/:account_id.json

해당 기간내의 특정 항목에 대해서 조회를 합니다.

**Resource URL:** `https://whooing.com/api/in_out/:account/:account_id.json`

**파라미터:**

- **:account** (required), **:account_id** (required)
- **section_id** (required), start_date, end_date: 생략 시 각각 한 달 전/오늘로 자동 설정. 하나만 생략하면 에러

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": { "in": 1066000, "out": 421833, "margin": 644167 }
}
```

---

### Calendar - 캘린더

#### GET calendar.json

해당 기간내의 월별/일별 수익·비용·기타거래의 액수를 표시합니다. day는 일요일부터 0~6의 값입니다.

**Resource URL:** `https://whooing.com/api/calendar.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- start_date: 조회 시작 날짜 (예: 201101). 생략 시 한 달 전 날짜로 자동 설정
- end_date: 조회 종료 날짜 (예: 201106). 생략 시 오늘 날짜로 자동 설정. start_date와 end_date 중 하나만 생략하면 에러
- account: 왼쪽이나 오른쪽에 해당 계정이 나타나는 경우로 제한
- account_id: 왼쪽이나 오른쪽에 해당 계정과 항목이 나타나는 경우로 제한 (account 필수)
- l_account: 왼쪽에 해당 계정이 나타나는 경우로 제한
- l_account_id: 왼쪽에 해당 계정과 항목이 나타나는 경우로 제한 (l_account 필수)
- r_account: 오른쪽에 해당 계정이 나타나는 경우로 제한
- r_account_id: 오른쪽에 해당 계정과 항목이 나타나는 경우로 제한 (r_account 필수)
- item: 아이템 or 거래처에 특정 값을 가지는 경우로 제한 (\*를 와일드카드로 쓸 수 있음)

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"aggregate": {
			"income": 192300,
			"expenses": 203000,
			"etc": 23040
		},
		"rows_type": "month",
		"rows": {
			"201106": [
				{
					"date": "20110601",
					"day": 0,
					"count": 3,
					"income": 9390000,
					"expenses": 293000,
					"etc": 0
				}
			]
		}
	}
}
```

---

### Report - 자산/부채/비용/수익 보고서

> **통합 보고서 API**로, 아래 레거시 엔드포인트들을 대체합니다.
>
> | 대체 전 (deprecated)                             | 대체 방법                                                                       |
> | ------------------------------------------------ | ------------------------------------------------------------------------------- |
> | `pl.json` — 기간 손익 합계                       | `report/expenses,income.json?rows_type=none`                                    |
> | `bs.json` — 특정 시점 잔액                       | `report/assets,liabilities.json?rows_type=none`                                 |
> | `daily_pl.json` — 기간별 손익 추이               | `report_summary/expenses,income.json?rows_type=day\|month\|quarter`             |
> | `zigzag.json` — 월별 손익 항목변동               | `report/expenses,income/:account_id.json?rows_type=month`                       |
> | `zigzag_mountain.json` — 월별 자산/부채 항목변동 | `report/assets,liabilities/:account_id.json?rows_type=month`                    |
> | `mountain.json` — 총자산 변동 추이               | `report/assets,liabilities.json?rows_type=month` (목표는 `goal.json` 별도 조회) |

> **whoo-cli 호환성 메모 (2026-09-16):**
> 공식 대체 경로인 `/report/assets,liabilities.json`은 현재 콤마가 포함된
> 경로에서 HTTP 403을 반환한다. 단일 계정 report 경로는 정상 응답하므로,
> CLI는 deprecated된 `bs.json` 대신 자산·부채 report를 각각 조회한 결과를
> 병합한다.

#### GET report.json

통합 보고서 조회 API입니다. 항상 `{total, accounts}` 구조를 반환합니다. 기간별로 다양한 계정 타입(자산, 부채, 수익, 비용)의 데이터를 조회할 수 있으며, 일별, 월별, 분기별, 연도별 등 다양한 단위로 데이터를 그룹화하여 반환합니다.

**Resource URL:** `https://whooing.com/api/report.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- account_id: 항목의 고유번호 (예: x123). 생략 시 전체 항목 조회 (''와 동일)
- start_date: 조회 시작 날짜 (Ymd 또는 Ym 형식, 예: 20240101 또는 202401). 생략 시 한 달 전 날짜로 자동 설정
- end_date: 조회 종료 날짜 (Ymd 또는 Ym 형식, 예: 20241231 또는 202412). 생략 시 오늘 날짜로 자동 설정. start_date와 end_date 중 하나만 생략하면 에러. 조회 기간은 최대 10년(120개월)
- rows_type: 조회 단위. day, month, quarter, year, none. 생략 시 month
    - `day`: 일별로 그룹화 (최대 12개월)
    - `month`: 월별로 그룹화
    - `quarter`: 분기별로 그룹화
    - `year`: 연도별로 그룹화
    - `none`: 전체 기간 합계만 반환
- item: 항목명 필터링 (와일드카드 지원)
    - `*항목*`: 항목명에 "항목"이 포함된 모든 항목
    - `*항목`: "항목"으로 끝나는 모든 항목
    - `항목*`: "항목"으로 시작하는 모든 항목
    - `항목`: 정확히 "항목"과 일치하는 항목

**응답 구조** (항상 `{total, accounts}` 형식):

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"rows_type": "month",
		"rows": {
			"202401": {
				"date": "202401",
				"assets": {
					"total": 3000000,
					"accounts": { "x10": 2000000, "x11": 1000000 }
				},
				"liabilities": {
					"total": 500000,
					"accounts": { "x20": 500000 }
				},
				"capital": { "total": 2500000 }
			}
		},
		"aggregate": {
			"assets": { "total": 3500000, "accounts": { "x10": 2500000, "x11": 1000000 } },
			"liabilities": { "total": 600000, "accounts": { "x20": 600000 } },
			"capital": { "total": 2900000 }
		}
	}
}
```

**주요 특징:**

- 자산과 부채를 함께 조회하면 자본(capital)이 자동으로 계산됩니다
- 수익과 비용을 함께 조회하면 순이익(net_income)이 자동으로 계산됩니다
- 자산과 부채의 경우 시작일 이전의 잔액을 기초로 하여 누적 잔액을 계산합니다
- `aggregate`: 전체 조회 기간의 요약. 자산/부채/자본은 마지막 기간의 누적 잔액, 수익/비용은 전체 기간 합산
- `account_id`를 지정하면 `accounts`에 해당 항목만 포함됩니다

**주의사항:**

- `rows_type=day`인 경우 최대 12개월까지만 조회 가능합니다

#### GET report/:account.json

특정 계정 타입만 조회합니다.

**Resource URL:** `https://whooing.com/api/report/:account.json`

**파라미터:**

- **:account** (required): 조회할 계정 타입 (path parameter, 콤마로 구분하여 여러 개 지정 가능)
    - `assets`, `liabilities`, `expenses`, `income`, `all`
    - 예: `expenses` 또는 `expenses,income` 또는 `all`
- account_id: 항목의 고유번호 (예: x123). 생략 시 전체 항목 조회
- **section_id** (required), rows_type (생략 시 month)
- start_date, end_date: 생략 시 각각 한 달 전/오늘로 자동 설정. 하나만 생략하면 에러
- item: 항목명 필터링

#### GET report/:account/:account_id.json

특정 항목에 대해서 조회합니다.

**Resource URL:** `https://whooing.com/api/report/:account/:account_id.json`

**파라미터:**

- **:account** (required): 조회할 계정 타입 (예: expenses)
- **:account_id** (required): 항목의 고유번호 (예: x123)
- **section_id** (required), rows_type (생략 시 month)
- start_date, end_date: 생략 시 각각 한 달 전/오늘로 자동 설정. 하나만 생략하면 에러
- item: 항목명 필터링

#### GET report_summary.json

기간별 손익/자산 요약입니다. 항상 **flat 숫자** 응답을 반환합니다. 항목별 분해 없이 계정 타입별 합계만 필요할 때 사용합니다.

> `report.json`과의 차이: `report.json`은 `{total, accounts}` 구조로 항목별 상세를 반환하지만, `report_summary.json`은 계정 타입별 숫자만 반환합니다. BS 계정(자산/부채/자본)의 값 의미는 동일합니다 — 누적 잔액(해당 시점의 잔액 합계)을 반환합니다. PL 계정(수익/비용)은 기간 합산입니다.

**Resource URL:** `https://whooing.com/api/report_summary.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- start_date, end_date, rows_type, item: `report.json`과 동일
- account: `report.json`과 동일 (기본값: expenses,income)

**응답 구조** (항상 flat 숫자):

```json
{
	"code": 200,
	"message": "",
	"results": {
		"rows_type": "month",
		"rows": {
			"202401": {
				"date": "202401",
				"expenses": 500000,
				"income": 3000000,
				"net_income": 2500000
			}
		},
		"aggregate": {
			"expenses": 6000000,
			"income": 36000000,
			"net_income": 30000000
		}
	}
}
```

#### GET report_summary/:account.json

특정 계정 타입만 요약 조회합니다.

**Resource URL:** `https://whooing.com/api/report_summary/:account.json`

**파라미터:** `report_summary.json`과 동일 + **:account** (path parameter)

---

### Report Customs - 사용자 정의 보고서 행

재무상태표(report_bs) 또는 손익계산서(report_pl) 하단에 사용자가 직접 정의한 합산 행입니다.
`plus`/`minus`에 나열된 항목들의 금액을 더하고 빼서 행의 기본값을 구하고, `addminus` 공식으로 추가 변환을 적용합니다.

**plus/minus 원소 형식**: `"<account>_<account_id 또는 total>"`. 예: `"assets_x11"`, `"liabilities_total"`

**addminus 공식**: `x`를 plus-minus 합산값으로 하는 사칙연산 식. 예: `"x*0.1"`, `"x+1000"`. 변환이 없으면 `"x"`.

#### GET main/report_customs.json

사용자 정의 보고서 행 목록 또는 단건을 조회합니다.

**Resource URL:** `https://whooing.com/api/main/report_customs.json`

**파라미터:**

- **section_id** (required): 섹션 ID. 예: `s18`
- **report** (required): `report_bs` 또는 `report_pl`
- **action** (required): `list` (전체 목록) 또는 `info` (단건)
- **customId**: `action=info` 일 때 필수. 행 ID. 예: `12`

**응답 예시 (action=list):**

```json
{
  "status": "done",
  "rows": [
    {
      "id": "12",
      "report": "report_bs",
      "title": "현금성 자산",
      "plus": ["assets_x11", "assets_x12"],
      "minus": [],
      "addminus": "x",
      "money": 0
    }
  ]
}
```

#### POST main/report_customs.json

사용자 정의 보고서 행을 생성/수정/삭제합니다.

**Resource URL:** `https://whooing.com/api/main/report_customs.json`

**파라미터:**

- **section_id** (required): 섹션 ID
- **report** (required): `report_bs` 또는 `report_pl`
- **action** (required): `post` (생성/수정), `delete`, `sort`, `clean_disabled`
- **row**: `action=post` 시 필수. JSON 문자열 `{"id"?: number, "title": string, "plus": string[], "minus": string[], "addminus": string}`. id 없으면 신규 생성, 있으면 수정.
- **customId**: `action=delete` 시 필수. 삭제할 행 ID.
- **sorted**: `action=sort` 시 필수. 새 순서의 customId 배열.

---

### Goal - 월별 자본 목표

장기목표 설정(budget_goal)에서 파생된 월별 자본 도달 목표값입니다. `ui.budgetLong="y"` 섹션에서만 값이 존재합니다. 산 차트(mountain)의 목표 점선으로 시각화됩니다.

#### GET goal.json

해당 기간내의 월별 자본 목표값을 조회합니다. mountain API에도 포함되어 있지만, 목표값만 별도로 조회할 때 사용합니다. `ui.budgetLong="n"` 섹션은 항상 null을 반환합니다.

**Resource URL:** `https://whooing.com/api/goal.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- start_date: 조회 시작 년월 (예: 201101). 생략 시 3개월 전 날짜로 자동 설정
- end_date: 조회 종료 년월 (예: 201106). 생략 시 오늘 날짜로 자동 설정. start_date와 end_date 중 하나만 생략하면 에러

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": [
		{ "date": "201101", "money": 12300000 },
		{ "date": "201102", "money": 15200000 }
	]
}
```

#### PUT goal.json

월별 자본 목표값을 직접 수정합니다. 지정한 월 사이는 선형 보간으로 자동 채워집니다. 모든 월을 지정할 필요 없이 변경할 월만 전송하면 됩니다.

**Resource URL:** `https://whooing.com/api/goal.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- 202601: 목표 자본금액 (예: 21000000) — 각 년월(YYYYMM)을 키로, 목표금액을 값으로 전송
- 202606: 목표 자본금액 (예: 35000000)
- ...: (변경할 월만 전송. 지정한 월 사이는 자동 보간)

---

### Post-it - 포스트잇

#### GET post_it.json

모든 포스트잇 리스트를 요청합니다.

**Resource URL:** `https://whooing.com/api/post_it.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- page: 특정 페이지내의 포스트잇만 호출. 앞에 \_가 붙은 것은 후잉의 기본 페이지 (예: \_main/index, \_main/insert, \_main/entries)

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": [
		{
			"post_it_id": 13,
			"page": "_main/index",
			"everywhere": "n",
			"contents": "포스트잇의 내용1"
		}
	]
}
```

#### GET post_it/:post_it_id.json

특정 포스트잇의 정보를 요청합니다.

**Resource URL:** `https://whooing.com/api/post_it/:post_it_id.json`

**파라미터:**

- **:post_it_id** (required): 포스트잇 고유번호 (예: 2932)
- **section_id** (required): 섹션의 고유번호

#### POST post_it.json

포스트잇을 추가합니다.

**Resource URL:** `https://whooing.com/api/post_it.json`

**파라미터:**

- **section_id** (required): 섹션의 고유번호
- **page** (required): 위치할 페이지 (예: \_main/index, \_main/entries, \_main/income_budget)
- **contents** (required): 포스트잇 내용
- everywhere: 모든 페이지에 표시 여부 (예: n)
- color: 색상의 RGB 16진수값 (예: ffbd94)

#### PUT post_it/:post_it_id.json

포스트잇의 정보를 수정합니다. 변경사항이 있는 파라미터만 전송하면 됩니다.

**Resource URL:** `https://whooing.com/api/post_it/:post_it_id.json`

**파라미터:**

- **:post_it_id** (required): 포스트잇의 고유번호 (예: 3203)
- **section_id** (required): 섹션의 고유번호
- page, contents, everywhere (모두 선택)

#### DELETE post_it/:section_id/:post_it_id.json

특정 포스트잇을 삭제합니다.

**Resource URL:** `https://whooing.com/api/post_it/:section_id/:post_it_id.json`

**파라미터:**

- **:section_id** (required): 섹션의 고유번호
- **:post_it_id** (required): 포스트잇의 고유번호. 복수 삭제시 콤마(,)로 구분 (예: 2328,9812,1108)

---

### Messages - 쪽지

#### GET messages.json

최근의 유저별 쪽지리스트를 요청합니다.

**Resource URL:** `https://whooing.com/api/messages.json`

**파라미터:**

- max: 특정 timestamp(소수점 둘째자리까지 포함) 이전으로만 제한 (예: 1320754854.31)
- limit: 조회할 쪽지리스트의 수. 기본 20

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": [
		{
			"opponent_user_id": 2932344,
			"opponent_username": "helloman",
			"opponent_image_url": "https://static.whooing.com/profiles/p0.jpg",
			"timestamp": 130029832,
			"timestamp_id": "1320764066.23",
			"read": "y",
			"summary": "내일 굴다리 밑으로 꼭 와라..."
		}
	]
}
```

#### GET messages/:opponent_user_id.json

특정 유저와의 쪽지들을 요청합니다.

**Resource URL:** `https://whooing.com/api/messages/:opponent_user_id.json`

**파라미터:**

- **:opponent_user_id** (required): 상대의 고유번호 (예: 229932)
- max: 특정 message_id 이전으로만 제한
- since: 특정 message_id 이후로 새로 들어온 쪽지리스트를 불러들임 (limit 무효화)
- limit: 조회할 쪽지의 수. 기본 20

#### POST messages.json

특정 유저에게 쪽지를 전송합니다.

**Resource URL:** `https://whooing.com/api/messages.json`

**파라미터:**

- **opponent_user_ids** (required): 쪽지를 받는 상대방. 콤마로 구분하여 여러명에게 전송 가능 (예: 293)
- **message** (required): 쪽지의 내용. **최대 2,000자**이며 초과하면 `400`과 함께 `error_parameters.message`로 사유가 반환됩니다
- attachment_ids: 업로드된 파일의 attachment_id를 콤마(,)로 구분한 문자열

#### DELETE messages/:opponent_user_id.json

특정 유저와의 쪽지들을 모두 삭제합니다.

**Resource URL:** `https://whooing.com/api/messages/:opponent_user_id.json`

**파라미터:**

- **:opponent_user_id** (required): 특정 유저의 고유번호 (예: 292328)

#### GET messages/unread.json

읽지 않은 쪽지리스트의 수를 구합니다.

**Resource URL:** `https://whooing.com/api/messages/unread.json`

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4948,
	"results": 4
}
```

---

### BBS - 게시판

#### GET bbs.json

모든 게시물리스트를 요청합니다.

**Resource URL:** `https://whooing.com/api/bbs.json`

**파라미터:**

- language: 특정 언어로 제한 (예: ko)
- group: 특정 그룹의 게시물 (예: humor)
- q: 제목과 내용에 제한되는 검색어
- page: 표시될 페이지. max와 배타적으로 쓰이며 page 파라미터가 높은 우선순위를 가짐
- max: bbs_id가 max보다 작은 게시물로만 제한
- since: bbs_id가 since보다 큰 게시물로만 제한
- limit: 표시할 게시물의 갯수

#### GET bbs/:category.json

카테고리 내 게시물리스트를 조회합니다.

**Resource URL:** `https://whooing.com/api/bbs/:category.json`

**파라미터:**

- **:category** (required): 조회할 카테고리 (예: free)
- language, group, q, page, max, since, limit (모두 선택)

#### GET bbs/:category/:bbs_id.json

특정 게시물의 정보를 요청합니다. comment의 rows에 있는 대댓글들은 최신 3개만 표시합니다.

**Resource URL:** `https://whooing.com/api/bbs/:category/:bbs_id.json`

**파라미터:**

- **:category** (required): 조회할 카테고리 (예: free)
- **:bbs_id** (required): 게시물의 고유번호 (예: 2932)
- max: comment_id가 max보다 작은 코멘트들로만 제한
- since: comment_id가 since보다 큰 코멘트들로만 제한
- limit: 한 번에 표시할 코멘트들의 갯수. maximum: 31. limit=1로 요청하면 원 소스만 전송
- order: 코멘트 정렬 순서, desc or asc

#### POST bbs/:category.json

게시물을 추가합니다.

**Resource URL:** `https://whooing.com/api/bbs/:category.json`

**파라미터:**

- **:category** (required): 등록할 카테고리 (예: free)
- group: 등록할 그룹
- **subject** (required): 게시물의 제목
- **contents** (required): 게시물의 내용. 5000자까지 가능
- language: 언어
- attachment_ids: 파일 uuid를 콤마로 구분한 문자열

#### PUT bbs/:category/:bbs_id.json

게시물의 정보를 수정합니다. 변경사항이 있는 파라미터만 전송하면 됩니다.

**Resource URL:** `https://whooing.com/api/bbs/:category/:bbs_id.json`

**파라미터:**

- **:category** (required), **:bbs_id** (required)
- group, subject, contents, attachment_ids (모두 선택)

#### DELETE bbs/:category/:bbs_id.json

특정 게시물을 삭제합니다.

**Resource URL:** `https://whooing.com/api/bbs/:category/:bbs_id.json`

**파라미터:**

- **:category** (required), **:bbs_id** (required): 게시물의 고유번호. 복수 삭제시 콤마(,)로 구분

#### GET bbs/:category/:bbs_id/:comment_id.json

특정 코멘트의 정보(대댓글포함)를 요청합니다. 수정 목적이면 limit=1 포함.

**Resource URL:** `https://whooing.com/api/bbs/:category/:bbs_id/:comment_id.json`

#### POST bbs/:category/:bbs_id.json (코멘트 추가)

게시물에 코멘트를 추가합니다.

**Resource URL:** `https://whooing.com/api/bbs/:category/:bbs_id.json`

**파라미터:**

- **:category** (required), **:bbs_id** (required)
- **contents** (required): 코멘트의 내용. 500자까지 가능
- attachment_ids: 업로드 파일 id를 콤마(,)로 구분한 문자열

#### PUT bbs/:category/:bbs_id/:comment_id.json

코멘트의 정보를 수정합니다.

**Resource URL:** `https://whooing.com/api/bbs/:category/:bbs_id/:comment_id.json`

**파라미터:**

- **:category** (required), **:bbs_id** (required), **:comment_id** (required)
- contents, attachment_ids (선택)

#### DELETE bbs/:category/:bbs_id/:comment_id.json

특정 코멘트를 삭제합니다.

**Resource URL:** `https://whooing.com/api/bbs/:category/:bbs_id/:comment_id.json`

**파라미터:**

- **:category** (required), **:bbs_id** (required)
- **:comment_id** (required): 코멘트의 고유번호. 복수 삭제시 콤마(,)로 구분 (예: c2936,c198,c9117)

#### POST bbs/:category/:bbs_id/:comment_id.json (대댓글 추가)

코멘트에 대댓글을 추가합니다.

**Resource URL:** `https://whooing.com/api/bbs/:category/:bbs_id/:comment_id.json`

**파라미터:**

- **:category** (required), **:bbs_id** (required), **:comment_id** (required)
- **contents** (required): 대댓글의 내용. 255자까지 가능

#### DELETE bbs/:category/:bbs_id/:comment_id/:addition_id.json

특정 대댓글을 삭제합니다.

**Resource URL:** `https://whooing.com/api/bbs/:category/:bbs_id/:comment_id/:addition_id.json`

**파라미터:**

- **:category** (required), **:bbs_id** (required), **:comment_id** (required)
- **:addition_id** (required): 대댓글의 고유번호. 복수 삭제시 콤마(,)로 구분 (예: 2843,92837,0173)

#### PUT bbs/recommandation.json

게시물이나 코멘트를 추천/혹은 비추천합니다. 추천은 한 번 밖에 안되며, 중복으로 추천된다면 서버에서 스스로 제외합니다.

**Resource URL:** `https://whooing.com/api/bbs/recommandation.json`

**파라미터:**

- **bbs_id** (required): 원본글의 고유번호 (예: 12445)
- comment_id: 추천대상이 코멘트라면 코멘트의 고유번호 (예: c382445)

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4948,
	"results": "recommanded"
}
```

---

### Upload - 파일업로드

파일 첨부는 2단계로 진행됩니다: (1) presigned URL 취득 → (2) S3에 바이너리 업로드.

> **2026-08 변경**: 예전에는 (3) POST로 완료를 알리는 단계가 **필수**였고, 이 호출을 하지 않으면 파일이 최종 저장소로 옮겨지지 않아 첨부가 깨진 상태로 남았습니다. 이제 업로드가 처음부터 최종 저장소로 직접 들어가므로 **(2)번 PUT이 성공하면 그것으로 완료**입니다. POST 단계는 선택이며, 호출해도 그대로 동작합니다(기존 클라이언트 수정 불필요).

#### GET upload.json

게시판이나 거래에 파일을 첨부하기 위해 바이너리를 전송할 주소를 받습니다. 발급 후 **5분간만** 유효하므로 즉시 업로드를 시작하여야 합니다.

AWS S3 presigned 업로드 방법은 [aws s3 presigned upload client](https://www.google.com/search?q=aws+s3+presigend+upload+client)를 참고하십시오.

**Resource URL:** `https://whooing.com/api/upload.json`

**파라미터:**

- **name** (required): 실제 파일명 (예: example.jpg)
- **mimeType** (required): 파일의 [mime 타입](https://developer.mozilla.org/ko/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types) (예: image/jpeg)
- **size** (required): 파일의 사이즈. 바이트 단위. 최대 20메가 (예: 2809872)

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4948,
	"results": {
		"url": "https://whooing-upload-files.s3.ap-northeast-1.amazonaws.com/810cbdb1b-7486jvk57?X-Amz-Algorithm=...",
		"file_info": {
			"uuid": "810cbdb1b-7486jvk57",
			"mimeType": "image/jpeg",
			"size": 28098,
			"name": "example.jpg",
			"src": "https://static.whooing.com/get/810cbdb1b-7486jvk57",
			"filename": "example.jpg",
			"size_mb": 0.03
		}
	}
}
```

`results.url`로 파일 바이너리를 **PUT** 하면 업로드가 끝납니다. 이때 `Content-Type` 헤더는 위에서 넘긴 `mimeType`과 **동일해야** 합니다(서명에 포함되어 있어 다르면 403).

`src`·`filename`·`size_mb`는 화면 표시용으로 함께 내려주는 추가 필드입니다. 기존 `name`·`size`(바이트)는 그대로 유지됩니다.

얻어진 각 파일의 uuid는 거래나 게시글 등록시 `attachment_ids`에 콤마로 직렬화하여 전달합니다.

#### POST upload.json

**선택 단계입니다.** 업로드가 최종 저장소에 정상 반영되었는지 확인하고 첨부 정보를 돌려받습니다. 몇 번을 호출해도 결과가 같습니다.

파일이 실제로 저장되지 않은 상태(PUT 실패·미수행)에서 호출하면 `code: 400`을 반환하므로, 업로드 성공 여부를 서버에 되물어 확인하는 용도로 쓸 수 있습니다.

**Resource URL:** `https://whooing.com/api/upload.json`

**파라미터:**

- **uuid** (required): 위에서 부여받은 파일의 고유번호 (예: 810cbdb1b-7486jvk57)

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4948,
	"results": {
		"uuid": "810cbdb1b-7486jvk57",
		"src": "https://static.whooing.com/get/810cbdb1b-7486jvk57",
		"filename": "example.jpg",
		"mimeType": "image/jpeg",
		"size": 28098
	}
}
```

---

### Notifications - 알림

#### GET notifications.json

모든 알람을 요청합니다. 최근 2주이내의 알람중 최근 50개만 반환합니다.

써드파티 컨슈머에서는 RESTful로 끌어와서 확인하기 때문에 실시간을 보장하기는 어렵습니다. 사용자가 컨슈머를 활성화했을 때 백그라운드 작업으로 주기적으로 새로운 알람을 확인하는 구성으로 하시면 됩니다. 이 시간은 **최소 5분**으로 권장하고 있으며 너무 잦게 요청하는 경우 후잉 서버에 무리가 가므로 제한될 수 있습니다.

**Resource URL:** `https://whooing.com/api/notifications.json`

**파라미터:**

- section_id: 섹션의 고유번호. 이 값이 있으면 섹션의 임시저장소 갯수를 추가 반환 (예: s99)

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"account": 0,
		"messages": 0,
		"outside": 0,
		"payment": 0,
		"notification": {
			"lastTimestamp": 1609929187,
			"badgeCount": 0,
			"rows": [
				{
					"noti_id": 1680607,
					"contents": "<b>테스트</b>님으로부터 쪽지 도착",
					"isNewest": true,
					"link": "messages/who/1",
					"nav": "messages",
					"summary": "궁금한점이 있습니다",
					"timestamp": 1619418434
				}
			]
		}
	}
}
```

#### PUT notifications.json

알람리스트를 열어본 시간을 체크하여, 신규 알람의 뱃지를 리셋합니다. 사용자가 UI에서 명시적으로 리스트를 열어본 시점에서 발생시킵니다. 이 작업을 통해 후잉 API를 쓰는 모든 써드파티에서 일관된 뱃지 카운트를 보장합니다.

**Resource URL:** `https://whooing.com/api/notifications.json`

파라미터 없음.

```json
{
	"code": 200,
	"message": "",
	"error_parameters": {},
	"rest_of_api": 4988,
	"results": {
		"lastTimestamp": 1609929187
	}
}
```

---

## 5. 개발자 약관

### 제 1 조 (목적)

이 약관은 후잉(whooing)(이하 "회사")과 참여 개발자(이하 "서드파티")간에 서비스(이하 "서비스")의 개발과 관련하여 서드파티의 제한 사항 등 기타 필요한 사항을 규정함을 목적으로 합니다.

### 제 2 조 (범위)

이 약관은 후잉서비스 이용약관의 하위 약관으로 적용하며, 명시하지 않은 사항은 후잉서비스 이용약관에 따릅니다.

### 제 3 조 (수익)

서드파티는 독립적으로 수익모델을 구성할 수 있으며, 회사는 서드파티 서비스의 소유권 및 수익에 대하여 관여하지 않습니다.

### 제 4 조 (제한)

회사는 다음 각호에 해당한다고 판단되는 경우에는 사전동의 없이 중단시킬 수 있습니다.

1. 서드파티 서비스가 회사에서 허용한 범위를 벗어날 경우
2. 회사의 저작권, 회원의 저작권 등 홈페이지 내의 기타 권리를 침해할 경우
3. 회원의 정보를 동의 없이 수집하거나, 수집된 정보를 목적과 달리 활용할 경우
4. 서드파티 서비스와 관련된 직간접적인 상황으로 인하여 회사에 피해가 있을 경우
5. 회원의 민원이나 신고가 지속되거나 사안이 중대하다고 판단될 경우
