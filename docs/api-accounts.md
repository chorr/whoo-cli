Accounts
모든 항목리스트를 요청합니다. 항목에 관한 정보는 후잉 전반에 걸쳐 매우 빈번하게 호출되며 쉽게 수정되지 않는 정보이기 때문에, 한 번 요청하면 그 값을 컨슈머 내부에 저장을 하는 것을 추천합니다. 변경된 사항이 있을 때마다 컨슈머에 저장된 변수를 일부 수정하거나 재요청을 하여 갱신하면 됩니다.

기본적으로 모든 기간에 해당되는 항목이 반환되는데, 사용자의 거래수정이나 입력 화면에서 보여주는 항목은 open_date와 close_date가 해당 시점에 걸쳐있는 것만을 표시하길 권해드립니다. 특정 시점에 결치있는 항목만을 요청하시려면 specific_date 파라미터를 추가하여 주십시오.

Resource URL
https://whooing.com/api/accounts.format

Parameters(purple color is required)
section_id
섹션의 고유번호 Example Value : s199
start_date
특정 날짜 이후에만 보여야만 하는 항목으로 제한. 이 방법보다는 전체 기간의 항목을 불러와서 앱 내부적으로 시점 계산을 하는 방법을 추천. Example Value : 20110817
end_date
특정 날짜 이전에만 보여야만 하는 항목으로 제한. 이 방법보다는 전체 기간의 항목을 불러와서 앱 내부적으로 시점 계산을 하는 방법을 추천. Example Value : 20110916
Example Results(.json)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
"assets" : {
"x1" : {
"account_id" : "x1",
"type" : "group",
"title" : "유동자산",
"memo" : "바로쓸 수 있는 것들",
"open_date" : 20090511,
"close_date" : 20160101,
"category" : "",
},
"x2" : {
"account_id" : "x2",
"type" : "account",
"title" : "현금",
"memo" : "내 지갑 및 서랍에 있는 돈",
"open_date" : 20090511,
"close_date" : 20160101,
"category" : "normal",
}
},
"liabilities" : {
"x10" : {
"account_id" : "x10",
"type" : "account",
"title" : "신한카드",
"memo" : "월 목표 사용액 : 50만원",
"open_date" : 20110101,
"close_date" : 21000101,
"category" : "creditcard",
"opt_use_date" : "p1",
"opt_pay_date" : 25,
"opt_pay_account_id" : "x1"
}
},
"capital" : {
"x8" : {
"account_id" : "x8",
"type" : "account",
"title" : "초기설정",
"memo" : "기초자금 설정 및 자본수정",
"open_date" : 20100101,
"close_date" : 20100101,
"cetegory" : ""
}
},
"income" : {
"x21" : {
"account_id" : "x21",
"type" : "account",
"title" : "주수익",
"memo" : "월급 및 기타소득",
"open_date" : 20010101,
"close_date" : 21000101,
"category" : "steady"
}
},
"expenses" : {
"x23" : {
"account_id" : "x23",
"type" : "account",
"title" : "식비",
"memo" : "일반 생활식비",
"open_date" : 20010101,
"close_date" : 21000101,
"category" : "steady"
}
}
}
}
Example Results(.json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
"assets" : [
{
"account_id" : "x1",
"type" : "group",
"title" : "유동자산",
"memo" : "바로쓸 수 있는 것들",
"open_date" : 20090511,
"close_date" : 20160101,
"category" : "",
},
{
"account_id" : "x2",
"type" : "account",
"title" : "현금",
"memo" : "내 지갑 및 서랍에 있는 돈",
"open_date" : 20090511,
"close_date" : 20160101,
"category" : "normal",
}
],
"liabilities" : [
{
"account_id" : "x10",
"type" : "account",
"title" : "신한카드",
"memo" : "월 목표 사용액 : 50만원",
"open_date" : 20110101,
"close_date" : 21000101,
"category" : "creditcard",
"opt_use_date" : "p1",
"opt_pay_date" : 25,
"opt_pay_account_id" : "x1"
}
],
"capital" : [
{
"account_id" : "x8",
"type" : "account",
"title" : "초기설정",
"memo" : "기초자금 설정 및 자본수정",
"open_date" : 20100101,
"close_date" : 20100101,
"cetegory" : ""
}
],
"income" : [
{
"account_id" : "x21",
"type" : "account",
"title" : "주수익",
"memo" : "월급 및 기타소득",
"open_date" : 20010101,
"close_date" : 21000101,
"category" : "steady"
}
],
"expenses" : [
{
"account_id" : "x23",
"type" : "account",
"title" : "식비",
"memo" : "일반 생활식비",
"open_date" : 20010101,
"close_date" : 21000101,
"category" : "steady"
}
]
}
}
계정의 항목리스트를 요청합니다.

기본적으로 모든 기간에 해당되는 항목이 반환되는데, 사용자의 거래수정이나 입력 화면에서 보여주는 항목은 open_date와 close_date가 해당 시점에 걸쳐있는 것만을 표시하길 권해드립니다. 특정 시점에 결치있는 항목만을 요청하시려면 specific_date 파라미터를 추가하여 주십시오.

Resource URL
https://whooing.com/api/accounts/:account.format

Parameters(purple color is required)
:account
조회할 계정
'assets' : 자산
'liabilities' : 부채
'capital' : 순자산
'expenses' : 비용
'income' : 수익
Example Value : assets
section_id
섹션의 고유번호 Example Value : s199
start_date
특정 날짜 이후에만 보여야만 하는 항목으로 제한. 이 방법보다는 전체 기간의 항목을 불러와서 앱 내부적으로 시점 계산을 하는 방법을 추천. Example Value : 20110817
end_date
특정 날짜 이전에만 보여야만 하는 항목으로 제한. 이 방법보다는 전체 기간의 항목을 불러와서 앱 내부적으로 시점 계산을 하는 방법을 추천. Example Value : 20110916
Example Results(.json)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
"x1" : {
"account_id" : "x1",
"type" : "group",
"title" : "유동자산",
"memo" : "바로쓸 수 있는 것들",
"open_date" : 20090511,
"close_date" : 20160101,
"category" : "",
},
"x2" : {
"account_id" : "x2",
"type" : "account",
"title" : "현금",
"memo" : "내 지갑 및 서랍에 있는 돈",
"open_date" : 20090511,
"close_date" : 20160101,
"category" : "normal",
}
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
"account_id" : "x1",
"type" : "group",
"title" : "유동자산",
"memo" : "바로쓸 수 있는 것들",
"open_date" : 20090511,
"close_date" : 20160101,
"category" : "",
},
{
"account_id" : "x2",
"type" : "account",
"title" : "현금",
"memo" : "내 지갑 및 서랍에 있는 돈",
"open_date" : 20090511,
"close_date" : 20160101,
"category" : "normal",
}
]
}
항목의 정보를 요청합니다.

Resource URL
https://whooing.com/api/accounts/:account/:account_id.format

Parameters(purple color is required)
:account
조회할 계정
'assets' : 자산
'liabilities' : 부채
'capital' : 순자산
'expenses' : 비용
'income' : 수익
Example Value : assets
:account_id
항목의 고유번호 Example Value : x2
section_id
섹션의 고유번호 Example Value : s199
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
"account_id" : "x2",
"type" : "account",
"title" : "현금",
"memo" : "내 지갑 및 서랍에 있는 돈",
"open_date" : 20090511,
"close_date" : 20160101,
"category" : "normal",
}
}
거래내역에서 해당 항목을 이용한 거래가 있는지 여부를 조사합니다. 주로 항목을 삭제하기 전이나 종료하기 전에 확인 겸 사용합니다.

last_one은 현재 시점에서 해당 계정에서 마지막 남은 항목인지 여부를 나타냅니다. 즉, last_one이 'y'로 리턴되면 지우거나 종료할 수 없습니다.

Resource URL
https://whooing.com/api/accounts/:account/:account_id/exists.format

Parameters(purple color is required)
:account
조회할 계정 Example Value : assets
:account_id
항목의 고유번호 Example Value : x2
section_id
섹션의 고유번호 Example Value : s199
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
"count" : 0,
"minDate" : 20100102,
"maxDate" : 20111232,
"balance" : 0,
"last_one" : "n",
"close_date" : 20120812
}
}
항목을 추가합니다.

Resource URL
https://whooing.com/api/accounts/:account.format

Parameters(purple color is required)
:account
대상 계정
'assets' : 자산
'liabilities' : 부채
'capital' : 순자산
'expenses' : 비용
'income' : 수익
Example Value : assets
section_id
섹션의 고유번호 Example Value : s99
title
항목의 이름. 0~30 글자. 30글자가 넘어가는 경우 30글자까지만 입력. Example Value : 현금
type
그룹인지 항목인지 구분 Example Value : account
open_date
항목의 사용이 시작되는 날짜 Example Value : 20010101
close_date
항목의 사용이 종료되는 날짜. 29991231의 경우 종료되지 않는 항목이라고 인식. Example Value : 29991231
memo
항목의 설명. 0~255글자. Example Value : 내 지갑 및 서랍에 있는 돈
category
항목의 종류
'normal' : 일반항목(기본값)
'client' : 거래처 관리 항목
'creditcard' : 신용카드
'checkcard' : 체크카드
'steady' : 고정 수익/비용
'floating' : 유동 수익/비용
Example Value : normal
opt_use_date
신용카드인 경우 사용기간의 시작일. p 한 글자는 전달을 의미함. pp는 전전달. 범위는 pp1 ~ p31. Example Value : pp1
opt_pay_date
신용카드인 경우 대금결제일. 범위는 1~31. Example Value : 25
opt_pay_account_id
신용카드인 경우 대금결제하는 상대 자산항목. Example Value : x2
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
"account_id" : "x2",
"type" : "account",
"title" : "현금",
"memo" : "내 지갑 및 서랍에 있는 돈",
"open_date" : 20090511,
"close_date" : 20160101,
"category" : "normal",
}
}
항목의 정보를 수정합니다. 전체 필드를 전달해야합니다.

Resource URL
https://whooing.com/api/accounts/:account/:account_id.format

Parameters(purple color is required)
:account
대상 계정
'assets' : 자산
'liabilities' : 부채
'capital' : 순자산
'expenses' : 비용
'income' : 수익
Example Value : assets
:account_id
항목의 고유번호 Example Value : x2
section_id
섹션의 고유번호 Example Value : s99
title
항목의 이름. 0~30 글자. 30글자가 넘어가는 경우 30글자까지만 입력. Example Value : 현금
open_date
항목의 사용이 시작되는 날짜 Example Value : 20010101
close_date
항목의 사용이 종료되는 날짜. 29991231의 경우 종료되지 않는 항목이라고 인식. 항목 수정폼을 전송하기 이전에 GET accounts/:account/:account_id/exists.format로 balance와 maxDate값을 구하여 사용자에게 어떤 변화가 일어날지를 미리 공지하는 것이 좋음 Example Value : 29991231
memo
항목의 설명. 0~80글자. Example Value : 내 지갑 및 서랍에 있는 돈
category
항목의 종류
'normal' : 일반항목(기본값)
'client' : 거래처 관리 항목
'creditcard' : 신용카드
'checkcard' : 체크카드
'steady' : 고정 수익/비용
'floating' : 유동 수익/비용
Example Value : normal
opt_use_date
신용카드인 경우 사용기간의 시작일. p 한 글자는 전달을 의미함. pp는 전전달. 범위는 pp1 ~ p31. Example Value : pp1
opt_pay_date
신용카드인 경우 대금결제일. 범위는 1~31. Example Value : 25
opt_pay_account_id
신용카드인 경우 대금결제하는 상대 자산항목. Example Value : x2
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
"account_id" : "x2",
"type" : "account",
"title" : "현금",
"memo" : "내 지갑 및 서랍에 있는 돈",
"open_date" : 20090511,
"close_date" : 20160101,
"category" : "normal",
}
}
특정 항목을 삭제합니다. 항목을 삭제한다는 것은 해당 항목과 관련된 거래가 하나도 없다는 전제하에 진행됩니다.

삭제를 실시하기 전에 해당 항목을 포함한 거래가 있는 필히 검사하여야 합니다(위에 GET exist API를 이용). 만약 거래가 있는데도 불구하고 삭제를 강행하는 경우에는 해당 항목을 포함한 모든 거래내역의 항목이 'x0'으로 변환되며 보고서에서 표시되지 않게됩니다. 사용자가 고의적으로 삭제를 강행하는 경우에는 이에 대해서 상세히 알려주어야 합니다.

Resource URL
https://whooing.com/api/accounts/:account/:account_id/:section_id.format

Parameters
:account
대상 계정
'assets' : 자산
'liabilities' : 부채
'capital' : 순자산
'expenses' : 비용
'income' : 수익
Example Value : assets
:account_id
항목의 고유번호. 복수의 항목삭제는 지원하지 않음. Example Value : x2
:section_id
섹션의 고유번호 Example Value : s99
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4948,
"results" : {}
}

항목의 순서를 변경합니다. 인터페이스 상에 현재 시점의 항목만 나열을 했더라도, 전송하는 account_ids에는 비활성화 되어 있는 모든 항목의 고유번호도 포함을 하여야 합니다.

Resource URL
https://whooing.com/api/accounts/:account/sort.format

Parameters
:account
대상 계정 Example Value : assets
section_id
섹션의 고유번호 Example Value : s99
account_ids
정렬하려는 순서에 맞추어 전체 항목 고유번호들을 콤마(,)로 구분한 문자열 Example Value : x2,x4,x3,x5
Results(.json, .json_array)
{
'code' : 200,
"message" : "",
"error_parameters" : {},
'rest_of_api' : 2388,
'results' : [
"x2",
"x4",
"x3",
"x5"
]
}
