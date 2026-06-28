Sections
유저의 섹션 리스트를 요청합니다. isolation값이 y인 경우에는 섹션리스트에서 나타나지 않길 원하는 비자금일 경우도 있으므로 컨슈머의 목적에 따라 별도로 처리하여 줍니다.

Resource URL
https://whooing.com/api/sections.format

Parameters(purple color is required)

-

Example Results(.json)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
"s123" : {
"section_id" : "s123",
"title" : "유동성 자산",
"memo" : "자주접근하는 자산만 관리",
"currency" : "KRW",
"isolation" : "n",
"total_assets" : 2982799.00,
"total_liabilities" : 23910.00,
"skin_id" : 0,
"decimal_places" : 2,
"date_format" : "YMD"
"webhook_token" : "xxxx-xxxx-xxxx-xxxx-xxxx"
},
"s283" : {
"section_id" : "s283",
"title" : "부동자산",
"memo" : "규모가 큰 자산들",
"currency" : "KRW",
"isolation" : "n",
"total_assets" : 1929882838.00,
"total_liabilities" : 2328862.00,
"skin_id" : 1,
"decimal_places" : 2,
"date_format" : "YMD",
"webhook_token" : "xxxx-xxxx-xxxx-xxxx-xxxx"
},
}
}
Example Results(.json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : [
{
"section_id" : "s123",
"title" : "유동성 자산",
"memo" : "자주접근하는 자산만 관리",
"currency" : "KRW",
"isolation" : "n",
"total_assets" : 2982799.00,
"total_liabilities" : 23910.00,
"skin_id" : 0,
"decimal_places" : 2,
"date_format" : "YMD",
"webhook_token" : "xxxx-xxxx-xxxx-xxxx-xxxx"
},
{
"section_id" : "s283",
"title" : "부동자산",
"memo" : "규모가 큰 자산들",
"currency" : "KRW",
"isolation" : "n",
"total_assets" : 1929882838.00,
"total_liabilities" : 2328862.00,
"skin_id" : 1,
"decimal_places" : 2,
"date_format" : "YMD",
"webhook_token" : "xxxx-xxxx-xxxx-xxxx-xxxx"
},
]
}
특정 섹션의 정보를 요청합니다. 기본섹션으로 지정되어 있는 것을 불러오기 위해서는 https://whooing.com/api/sections/default.format로 호출해주시면 됩니다.

Resource URL
https://whooing.com/api/sections/:section_id.format

Parameters(purple color is required)
:section_id
섹션의 고유번호 Example Value : s122
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
"section_id" : "s123",
"title" : "유동성 자산",
"memo" : "자주접근하는 자산만 관리",
"currency" : "KRW",
"isolation" : "n",
"total_assets" : 2982799.00,
"total_liabilities" : 23910.00,
"skin_id" : 1,
"decimal_places" : 2,
"date_format" : "YMD",
"webhook_token" : "xxxx-xxxx-xxxx-xxxx-xxxx",
"ui" : {
"billOrder" : "asc"
"billPreset" : "p_l_7"
"bsPreset" : "p_l_5"
"budgetOrder" : "asc"
"budgetPreset" : "p_m_12"
"budgetStyle" : "percent"
"daily_plPreset" : "p_l_3"
"entriesMethod" : "out"
"entriesPreset" : "p_l_5"
"in_outPreset" : "p_l_6"
"insertDateUp" : "n"
"insertMethod" : "0"
"insertMoneyUnit" : "1000"
"insertSlot" : "2"
"insertViewGroup" : "y"
"mainIndex" : "index"
"mountainPreset" : "p_l_3"
"plPreset" : "p_l_5"
"viewCapital" : "1"
"width" : "normal"
"zigzagPreset" : "p_l_3"
}
}
}
섹션을 추가합니다. 각 사용자당 섹션은 최대 9개까지 추가할 수 있습니다.

Resource URL
https://whooing.com/api/sections.format

Parameters(purple color is required)
title
섹션의 제목. 1~30 글자. Example Value : 부동산 모음'
currency
섹션의 통화단위 Example Value : USD
memo
섹션 보조설명. 0~80 글자. Example Value : 단위가 큰 자산유형만 별도 관리
skin_id
등록된 스킨의 고유번호 Example Value : 2
decimal_places
소수점 이하로 몇자리 까지 표시할 것인지 결정. 0~3. Example Value : 2
date_format
날짜를 표시하는 방법 Example Value : YMD
start_year
섹션 생성시 항목(accounts)의 open_date를 설정하는 시작연도. 2000년 ~ 현재연도 범위만 가능. 기본값은 2000년. 예를 들어 2020년을 입력하면 생성되는 항목들의 open_date가 20200101로 설정됩니다. Example Value : 2020
template_id
초기 항목 구성 템플릿 ID. 섹션 생성시 어떤 항목들을 기본으로 생성할지 결정합니다. Example Value : 1
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4984,
"results" : {
"section_id" : "s123",
"title" : "유동성 자산",
"memo" : "자주접근하는 자산만 관리",
"currency" : "KRW",
"isolation" : "n",
"total_assets" : 2982799.00,
"total_liabilities" : 23910.00,
"skin_id" : 1,
"decimal_places" : 2,
"date_format" : "YMD"
}
}
특정 섹션에 관한 정보를 수정합니다. 전체 필드를 입력해야합니다.

Resource URL
https://whooing.com/api/sections/:section_id.format

Parameters(purple color is required)
:section_id
섹션의 고유번호 Example Value : s129
title
섹션의 제목 Example Value : 부동산 모음
currency
섹션의 통화단위 Example Value : USD
memo
섹션 보조설명 Example Value : 단위가 큰 자산유형만 별도 관리
skin_id
스킨의 고유번호 Example Value : 2
decimal_places
소수점 이하로 몇자리 까지 표시할 것인지 결정. 0~3. Example Value : 2
date_format
날짜를 표시하는 방법 Example Value : YMD
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4984,
"results" : {
"section_id" : "s123",
"title" : "유동성 자산",
"memo" : "자주접근하는 자산만 관리",
"currency" : "KRW",
"isolation" : "n",
"total_assets" : 2982799.00,
"total_liabilities" : 23910.00,
"skin_id" : 2,
"decimal_places" : 2,
"date_format" : "YMD"
}
}
특정 섹션을 삭제합니다.

Resource URL
https://whooing.com/api/sections/:section_id.format

Parameters
:section_id
섹션의 고유번호. 복수의 섹션들을 한꺼번에 삭제하려면 콤마(,)로 이어붙인 문자열을 전송. Example Value : s129 or Example Value : s129,s118,s199
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4948
}

기본섹션을 조회합니다. 컨슈머의 초기화면에 섹션 리스트를 제공하지 않는 경우에는 기본섹션을 바로 출력하기 위해 요청합니다.

Resource URL
https://whooing.com/api/sections/default.format

Parameters

-

Results(.json, .json_array)
{
"code" : 200,
"error" : "",
"rest_of_api" : 2388,
"results" : {
"section_id" : "s123",
"title" : "유동성 자산",
"memo" : "자주접근하는 자산만 관리",
"currency" : "KRW",
"isolation" : "n",
"total_assets" : 2982799.00,
"total_liabilities" : 23910.00,
"skin_id" : 2,
"decimal_places" : 2,
"date_format" : "YMD"
}
}
섹션의 순서를 변경합니다.

Resource URL
https://whooing.com/api/sections/sort.format

Parameters
section_ids
정렬하려는 순서에 맞추어 전체 섹션 고유번호들을 콤마(,)로 구분한 문자열 Example Value : s99,s72,s78,s52
Results(.json, .json_array)
{
"code" : 200,
"error" : "",
"rest_of_api" : 2388,
"results" : [
"s99",
"s72",
"s78",
"s52"
]
}
