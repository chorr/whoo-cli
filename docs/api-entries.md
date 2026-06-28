Entries
조건에 맞는 거래내역을 조회합니다. 만약 연관보고서(reports)가 있는 경우 연관보고서를 요청할 때 동일한 파마미터들을 전송해야하므로, 이곳에 쓰인 파라미터는 메모리에 저장하고 있는 것이 유리합니다. results값 중에서 l_account_id나 r_account_id가 x0으로 반환되는 것들은 삭제되었지만 기록상 남겨둔 것입니다. 리스트를 출력할 때에 x0의 값을 가지는 것은 별도로 처리하여 '삭제된 항목'과 같은 식으로 표현해 주십시오.

Resource URL
https://whooing.com/api/entries.format

Parameters(purple color is required)
section_id
섹션의 고유번호 Example Value : s199
start_date
조회 시작 날짜 Example Value : 20110123
end_date
조회 종료 날짜(조회시작 ~ 종료까지는 최대 1년) Example Value : 20110603
max
특정 entry_date 이전으로만 제한. 기본 없음. 날짜뒤의 소수점은 내부적으로 사용되는 동일 날짜에서의 소팅을 위해서 이용됩니다. Example Value : 20110203.0034
limit
조회할 거래의 수. 기본 20. Example Value : 20
account
왼쪽이나 오른쪽에 해당 계정이 나타나는 경우로 제한 Example Value : assets
account_id
왼쪽이나 오른쪽에 해당 계정과 항목이 나타나는 경우로 제한(account 필수) Example Value : x39
l_account
왼쪽에 해당 계정이 나타나는 경우로 제한. Example Value : expenses
l_account_id
왼쪽에 해당 계정과 항목이 나타나는 경우로 제한(l_account 필수) Example Value : x239
r_account
오른쪽에 해당 계정이 나타나는 경우로 제한. Example Value : assets
r_account_id
오른쪽에 해당 계정과 항목이 나타나는 경우로 제한(r_account 필수) Example Value : x239
item
아이템 or 거래처에 특정 값을 가지는 경우로 제한(*를 와이들카드로 쓸 있음). 아이템은 와일드카드를 쓰지 않으면 정확히 일치하는 값을 검색하며, 괄호내의 문자는 기본적으로 해당 키워드가 포함한 문자를 검색. Example Value : 배고파(잇힝)
money_from
특정 금액부터 시작하는 거래로 제한. Example Value : 3000
money_to
특정 금액까지의 거래로 제한. Example Value : 80000
memo
특정 값을 포함하는 메모로 조건을 제한. 기본적으로 *memo\* 검색. Example Value : 배고파서 빵을
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
"reports" : [],
"rows" : [
{
"entry_id" : 1352827,
"entry_date" : 20110817.0001,
"l_account" : "expenses",
"l_account_id" : "x20",
"r_account" : "assets",
"r_account_id" : "x4",
"item" : "후원(과장학금)",
"money" : 10000,
"total" : 840721.99
"memo" : "",
"app_id" : 0,
"attachments" : [
{
"uuid": "810cbdb1b-7486jvk57",
"src": "https://static.whooing.com/get/810cbdb1b-7486jvk57",
"filename": "example.jpg",
"mimeType": "image/jpeg",
"size": 28098,
},
...
]
},
{
"entry_id" : 1352823,
"entry_date" : 20110813.0001,
"l_account" : "assets",
"l_account_id" : "x3",
"r_account" : "assets",
"r_account_id" : "x4",
"item" : "계좌이체",
"money" : 10000,
"total" : 840721.99
"memo" : "",
"app_id" : 0,
"attachments" : []
}
]
}
}
거래 정보를 조회합니다. 특정 거래를 수정할 때 기존에 불러진 정보를 이용하여도 되고, 새로 정보를 요청해서 수정 폼을 보여줄 수도 있습니다. results값 중에서 l_account_id나 r_account_id가 x0으로 반환되는 것들은 삭제되었지만 기록상 남겨둔 것입니다. 리스트를 출력할 때에 x0의 값을 가지는 것은 별도로 처리하여 '삭제된 항목'과 같은 식으로 표현해 주십시오.

Resource URL
https://whooing.com/api/entries/:entry_id.format

Parameters(purple color is required)
:entry_id
조회할 거래의 고유번호 Example Value : 1352827
section_id
섹션의 고유번호 Example Value : s199
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
"entry_id" : 1352827,
"entry_date" : 20110817.0001,
"l_account" : "expenses",
"l_account_id" : "x20",
"r_account" : "assets",
"r_account_id" : "x4",
"item" : "후원(과장학금)",
"money" : 10000,
"total" : "",
"memo" : "",
"app_id" : 0,
"attachments" : [
{
"uuid": "810cbdb1b-7486jvk57",
"src": "https://static.whooing.com/get/810cbdb1b-7486jvk57",
"filename": "example.jpg",
"mimeType": "image/jpeg",
"size": 28098,
},
...
]
}
}
거래를 추가합니다.

사용자에게 입력폼을 제공할 때, 왼쪽과 오른쪽에 보여지는 항목들은 사용자가 기입력한 거래입력 날짜에 활성화가 되어있는 항목들면 표시되어야 합니다. 즉, 거래입력 날짜의 값에 따라 실시간으로 왼쪽과 오른쪽이 변화되어야한다는 것입니다.

Resource URL
https://whooing.com/api/entries.format

Parameters(purple color is required)
section_id
섹션의 고유번호 Example Value : s99
entry_date
거래가 일어난 날짜 Example Value : 20110812
l_account
왼쪽의 계정 Example Value : expenses
l_account_id
왼쪽의 항목 고유번호 Example Value : x20
r_account
오른쪽의 계정 Example Value : assets
r_account_id
오른쪽의 항목 고유번호 Example Value : x4
item
아이템 or 거래처. 괄호메모나 명령어도 포함. Example Value : 후원(과장학금)**2
money
거래액 Example Value : 10000
memo
거래에 들어가는 보충 메모. 일기. 이 값으로는 검색할 수 없음. Example Value : 오늘도 어김없이 빠져나갔다
attachment_ids
별도의 API를 이용하여 미리 각 파일의 업로드와 고유주소를 취득하고, 이 고유번호를 콤마로 구분하여 직렬화한 값입니다. Example Value : 18abbd321-917kvo2j3, 10b23906c-547oytlc5
or
section_id
섹션의 고유번호 Example Value : s99
data_type
entries를 문자열로 만든 방식. 기본 json Example Value : json
entries
아래의 배열(object)을 json이나 serialize한 문자열, 최대 300개의 거래.
[
{
"entry_date" : 20110812,
"l_account" : "expenses",
"l_account_id" : "x20",
"r_account" : "assets",
"r_account_id" : "x4",
"item" : "후원(과장학금)**2",
"money" : 10000,
"memo" : "오늘도 어김없이 빠져나갔다"
}
]
Example Value : [{"entry_date":20110812,"l_account":"expenses","l_account-id":"x20","l_category":"floating","l_opt_pay_account_id":"","r_account":"assets","r_account_id":"x4","r_category":"normal","r_opt_pay_account_id":"","item":"\ud6c4\uc6d0(\uacfc\uc7a5\ud559\uae08)**2","money":10000,"memo":"\uc624\ub298\ub3c4 \uc5b4\uae40\uc5c6\uc774 \ube60\uc838\ub098\uac14\ub2e4"}]
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : [
{
"entry_id" : 1352827,
"entry_date" : 20110812.0001,
"l_account" : "expenses",
"l_account_id" : "x20",
"r_account" : "assets",
"r_account_id" : "x4",
"item" : "후원(과장학금)",
"money" : 10000,
"total" : "",
"memo" : "오늘도 어김없이 빠져나갔다",
"app_id" : 0,
"attachments" : [
{
"uuid": "810cbdb1b-7486jvk57",
"src": "https://static.whooing.com/get/810cbdb1b-7486jvk57",
"filename": "example.jpg",
"mimeType": "image/jpeg",
"size": 28098,
},
...
]
},
{
"entry_id" : 1352827,
"entry_date" : 20110912.0001,
"l_account" : "expenses",
"l_account_id" : "x20",
"r_account" : "assets",
"r_account_id" : "x4",
"item" : "후원(과장학금)",
"money" : 10000,
"total" : "",
"memo" : "오늘도 어김없이 빠져나갔다",
"app_id" : 0,
"attachments" : [
{
"uuid": "810cbdb1b-7486jvk57",
"src": "https://static.whooing.com/get/810cbdb1b-7486jvk57",
"filename": "example.jpg",
"mimeType": "image/jpeg",
"size": 28098,
},
...
]
}
]
}
거래를 수정합니다. 수정이 필요한 필드만 전송합니다.

사용자에게 입력폼을 제공할 때, 왼쪽과 오른쪽에 보여지는 항목들은 사용자가 기입력한 거래입력 날짜에 활성화가 되어있는 항목들만 표시되어야 합니다. 즉, 거래입력 날짜의 값에 따라 실시간으로 왼쪽과 오른쪽이 변화되어야한다는 것입니다.

Resource URL
https://whooing.com/api/entries/:entry_id.format

Parameters(purple color is required)
:entry_id
수정할 거래의 고유번호. Example Value : 1352827
section_id
섹션의 고유번호 Example Value : s99
entry_date
거래가 일어난 날짜 Example Value : 20110812
l_account
왼쪽의 계정 Example Value : expenses
l_account_id
왼쪽의 항목 고유번호 Example Value : x20
r_account
오른쪽의 계정 Example Value : assets
r_account_id
오른쪽의 항목 고유번호 Example Value : x4
item
아이템. 괄호메모도 포함. 명령어는 제외. Example Value : 후원(과장학금)
money
거래액 Example Value : 10000
memo
거래에 들어가는 보충 메모. 일기. Example Value : 오늘도 어김없이 빠져나갔다
attachment_ids
별도의 API를 이용하여 미리 각 파일의 업로드와 고유주소를 취득하고, 이 고유번호를 콤마로 구분하여 직렬화한 값입니다. Example Value : 18abbd321-917kvo2j3, 10b23906c-547oytlc5
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
'entry_id' : 1352827,
'entry_date' : 20110812.0001,
'l_account' : 'expenses',
'l_account_id' : 'x20',
'r_account' : 'assets',
'r_account_id' : 'x4',
'item' : '후원(과장학금)',
'money' : 10000,
'total' : '',
'memo' : '오늘도 어김없이 빠져나갔다',
'app_id' : 0,
"attachments" : [
{
"uuid": "810cbdb1b-7486jvk57",
"src": "https://static.whooing.com/get/810cbdb1b-7486jvk57",
"filename": "example.jpg",
"mimeType": "image/jpeg",
"size": 28098,
},
...
]
}
}
복수개의 거래를 수정합니다.

이 요청은 다른 API와는 좀 예외적인 것으로, 정확히 수정에 관련된 파라미터와 값에 대해서만 요청을 하시면 됩니다.

Resource URL
https://whooing.com/api/entries/:entry_ids/:section_id.format

Parameters(purple color is required)
:entry_ids
수정할 거래의 고유번호들. 콤마(,)로 이은 문자열. 최대 100개까지 가능. Example Value : 1352827,1352828,1352829
:section_id
섹션의 고유번호 Example Value : s99
entry_date
거래날짜를 일괄 수정할 경우 이 값을 전송합니다. Example Value : 20110812
l_account
거래들의 왼쪽의 계정을 수정할 경우 이 값을 전송합니다. Example Value : expenses
l_account_id
위의 파라미터와 같이 전송되어야 합니다. Example Value : x20
r_account
거래들의 오른쪽의 계정을 수정할 경우 이 값을 전송합니다. Example Value : assets
r_account_id
위의 파라미터와 같이 전송되어야 합니다. Example Value : x4
item
거래들의 아이템들을 수정할 경우에 이 값을 전송합니다. Example Value : 후원(과장학금)
money
거래들의 거래액을 수정할 경우에 이 값을 전송합니다. Example Value : 10000
memo
거래들의 메모를 수정할 경우 이 값을 전송합니다. Example Value : 오늘도 어김없이 빠져나갔다
attachment_ids
복수의 거래에서는 지원하지 않습니다.
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 348,
"results" : {
2352827 : {
'entry_id' : 1352827,
'entry_date' : 20110812.0001,
'l_account' : 'expenses',
'l_account_id' : 'x20',
'r_account' : 'assets',
'r_account_id' : 'x4',
'item' : '후원(과장학금)',
'money' : 10000,
'total' : '',
'memo' : '오늘도 어김없이 빠져나갔다',
'app_id' : 0,
"attachments": []
},
.
.
.
}
}
특정 거래를 삭제합니다.

Resource URL
https://whooing.com/api/entries/:entry_ids/:section_id.format

Parameters
:entry_ids
거래의 고유번호(복수개의 경우에는 콤마로 구분한 문자열), 최대 100개까지 가능. Example Value : 18377288
section_id
섹션의 고유번호 Example Value : s199
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4948
}

최근에 입력된 거래내역을 조회합니다. 주로 거래입력의 하단에 참고용으로 표시됩니다. results값 중에서 l_account_id나 r_account_id가 x0으로 반환되는 것들은 삭제되었지만 기록상 남겨둔 것입니다. 리스트를 출력할 때에 x0의 값을 가지는 것은 별도로 처리하여 '삭제된 항목'과 같은 식으로 표현해 주십시오.

Resource URL
https://whooing.com/api/entries/latest.format

Parameters(purple color is required)
section_id
섹션의 고유번호 Example Value : s199
max
특정 entry_id 이전으로만 제한. 기본 없음. Example Value : 8285331
limit
조회할 거래의 수. 기본 20. Example Value : 20
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : [
{
"entry_id" : 1352827,
"entry_date" : 20110817.0001,
"l_account" : "expenses",
"l_account_id" : "x20",
"r_account" : "assets",
"r_account_id" : "x4",
"item" : "후원(과장학금)",
"money" : 10000,
"total" : "",
"memo" : "",
"app_id" : 0,
"attachments" : [
{
"uuid": "810cbdb1b-7486jvk57",
"src": "https://static.whooing.com/get/810cbdb1b-7486jvk57",
"filename": "example.jpg",
"mimeType": "image/jpeg",
"size": 28098,
},
...
]
},
{
"entry_id" : 1352823,
"entry_date" : 20110813.0001,
"l_account" : "assets",
"l_account_id" : "x3",
"r_account" : "assets",
"r_account_id" : "x4",
"item" : "계좌이체",
"money" : 10000,
"total" : "",
"memo" : "",
"app_id" : 0,
"attachments" : [
{
"uuid": "810cbdb1b-7486jvk57",
"src": "https://static.whooing.com/get/810cbdb1b-7486jvk57",
"filename": "example.jpg",
"mimeType": "image/jpeg",
"size": 28098,
},
...
]
}
]
}
최근 60일 이내에 입력된 거래내역을 중복없이 불러옵니다. 주로 사용자가 거래입력을 할 때 Suggest기능을 위해 활용될 수 있습니다. 이 값의 일부는 자주 변동되지만 시스템 자원이용 차원에서 클라이언트 내부에 저장을 한 후에, 변동 내역이 있을 때마다 직접 변경을 해주는 식으로 응용하길 추천합니다.

Resource URL
https://whooing.com/api/entries/latest_items.format

Parameters(purple color is required)
section_id
섹션의 고유번호 Example Value : s199
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : [
{
"l_account" : "expenses",
"l_account_id" : "x20",
"r_account" : "assets",
"r_account_id" : "x4",
"item" : "후원",
"money" : 10000
},
{
"l_account" : "assets",
"l_account_id" : "x3",
"r_account" : "assets",
"r_account_id" : "x4",
"item" : "계좌이체",
"money" : 23880
}
]
}

특정 계정과 모든 계정/항목의 상대적인 증가/감소를 조회합니다.

Resource URL
https://whooing.com/api/entries/flow_of_account.format

Parameters(purple color is required)

- GET entries를 호출할 때의 동일한 파라미터
  Example Results(.json)
  {
  "code" : 200,
  "message" : "",
  "error_parameters" : {},
  "rest_of_api" : 4988,
  "results" : {
  "assets" : {
  "total" : {
  "from" : 324000,
  "to" : 4000,
  "margin" : 320000
  },
  "accounts" : {
  "x1" : {
  "from" : 124000,
  "to" : 0,
  "margin" : 124000
  },
  "x2" : {
  "from" : 24000,
  "to" : 2000,
  "margin" : 22000
  }
  }
  },
  "libilities" : {
  .
  .
  .
  },
  "capital" : {
  .
  .
  .
  },
  "income" : {
  .
  .
  .
  },
  "expenses" : {
  .
  .
  .
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
  "assets" : {
  "total" : {
  "from" : 324000,
  "to" : 4000,
  "margin" : 320000
  },
  "accounts" : [
  {
  "from" : 124000,
  "to" : 0,
  "margin" : 124000
  },
  {
  "from" : 24000,
  "to" : 2000,
  "margin" : 22000
  }
  ]
  },
  "libilities" : {
  .
  .
  .
  },
  "capital" : {
  .
  .
  .
  },
  "income" : {
  .
  .
  .
  },
  "expenses" : {
  .
  .
  .
  }
  }
  }
  특정 항목과 모든 계정/항목의 상대적인 증가/감소를 조회합니다.

Resource URL
https://whooing.com/api/entries/flow_of_account_id.format

Parameters(purple color is required)

- GET entries를 호출할 때의 동일한 파라미터
  Example Results(.json)
  {
  "code" : 200,
  "message" : "",
  "error_parameters" : {},
  "rest_of_api" : 4988,
  "results" : {
  "assets" : {
  "total" : {
  "from" : 324000,
  "to" : 4000,
  "margin" : 320000
  },
  "accounts" : {
  "x1" : {
  "from" : 124000,
  "to" : 0,
  "margin" : 124000
  },
  "x2" : {
  "from" : 24000,
  "to" : 2000,
  "margin" : 22000
  }
  }
  },
  "libilities" : {
  .
  .
  .
  },
  "capital" : {
  .
  .
  .
  },
  "income" : {
  .
  .
  .
  },
  "expenses" : {
  .
  .
  .
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
  "assets" : {
  "total" : {
  "from" : 324000,
  "to" : 4000,
  "margin" : 320000
  },
  "accounts" : [
  {
  "from" : 124000,
  "to" : 0,
  "margin" : 124000
  },
  {
  "from" : 24000,
  "to" : 2000,
  "margin" : 22000
  }
  ]
  },
  "libilities" : {
  .
  .
  .
  },
  "capital" : {
  .
  .
  .
  },
  "income" : {
  .
  .
  .
  },
  "expenses" : {
  .
  .
  .
  }
  }
  }
  특정 항목의 일일 변동내역을 표시합니다.

Resource URL
https://whooing.com/api/entries/changes_of_account_id.format

Parameters(purple color is required)

- GET entries를 호출할 때의 동일한 파라미터
  Example Results(.json)
  {
  "code" : 200,
  "message" : "",
  "error_parameters" : {},
  "rest_of_api" : 4988,
  "results" : {
  "aggregate" : {
  "in" : 1010002,
  "out" : 298933
  },
  "rows_type" : "day",
  "rows" : {
  "20110616" : 0,
  "20110617" : 230,
  "20110618" : -2230,
  "20110619" : 2300,
  "20110620" : -40,
  "20110621" : 23230,
  .
  .
  .
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
  "aggregate" : {
  "in" : 1010002,
  "out" : 298933
  },
  "rows_type" : "day",
  "rows" : [
  {
  "date" : "20110616",
  "money" : 0
  },
  {
  "date" : "20110617",
  "money" : 230
  }
  .
  .
  .
  ]
  }
  }
  특정 거래처의 일일 변동내역을 표시합니다.

Resource URL
https://whooing.com/api/entries/changes_of_client.format

Parameters(purple color is required)

- GET entries를 호출할 때의 동일한 파라미터
  Example Results(.json)
  {
  "code" : 200,
  "message" : "",
  "error_parameters" : {},
  "rest_of_api" : 4988,
  "results" : {
  "aggregate" : {
  "in" : 1010002,
  "out" : 298933
  },
  "rows_type" : "day",
  "rows" : {
  "20110616" : 0,
  "20110617" : 230,
  "20110618" : -2230,
  "20110619" : 2300,
  "20110620" : -40,
  "20110621" : 23230,
  .
  .
  .
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
  "aggregate" : {
  "in" : 1010002,
  "out" : 298933
  },
  "rows_type" : "day",
  "rows" : [
  {
  "date" : "20110616",
  "money" : 0
  },
  {
  "date" : "20110617",
  "money" : 230
  }
  .
  .
  .
  ]
  }
  }
  특정 아이템의 일일 발생내역을 표시합니다.

Resource URL
https://whooing.com/api/entries/changes_of_item.format

Parameters(purple color is required)

- GET entries를 호출할 때의 동일한 파라미터
  Example Results(.json)
  {
  "code" : 200,
  "message" : "",
  "error_parameters" : {},
  "rest_of_api" : 4988,
  "results" : {
  "aggregate" : {
  "in" : 101002,
  "out" : 0
  },
  "rows_type" : "day",
  "rows" : {
  "20110616" : 0,
  "20110617" : 230,
  "20110618" : 0,
  "20110619" : 2300,
  "20110620" : 2340,
  "20110621" : 97670,
  .
  .
  .
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
  "aggregate" : {
  "in" : 101002,
  "out" : 0
  },
  "rows_type" : "day",
  "rows" : [
  {
  "date" : "20110616",
  "money" : 0
  },
  {
  "date" : "20110617",
  "money" : 230
  }
  .
  .
  .
  ]
  }
  }
  계정의 항목별 금액을 조회합니다.

Resource URL
https://whooing.com/api/entries/account_ids_of_account.format

Parameters(purple color is required)

- GET entries를 호출할 때의 동일한 파라미터
  Example Results(.json, .json_array)
  {
  "code" : 200,
  "message" : "",
  "error_parameters" : {},
  "rest_of_api" : 4988,
  "results" : [
  {
  "name" : "데이트비",
  "money" : 865030
  },
  {
  "name" : "유흥비",
  "money" : 23998
  }
  .
  .
  .
  ]
  }
  항목의 거래처별 금액을 조회합니다.

Resource URL
https://whooing.com/api/entries/clients_of_account_id.format

Parameters(purple color is required)

- GET entries를 호출할 때의 동일한 파라미터
  Example Results(.json, .json_array)
  {
  "code" : 200,
  "message" : "",
  "error_parameters" : {},
  "rest_of_api" : 4988,
  "results" : [
  {
  "name" : "김유리",
  "money" : 230000
  },
  {
  "name" : "김성민",
  "money" : 10000
  }
  ]
  }
  항목의 아이템별 금액을 조회합니다.

Resource URL
https://whooing.com/api/entries/items_of_account_id.format

Parameters(purple color is required)

- GET entries를 호출할 때의 동일한 파라미터
  Example Results(.json, .json_array)
  {
  "code" : 200,
  "message" : "",
  "error_parameters" : {},
  "rest_of_api" : 4988,
  "results" : [
  {
  "name" : "주식",
  "money" : 2977200
  },
  {
  "name" : "간식",
  "money" : 330000
  }
  ]
  }

외부데이터들을 파싱하고 총 인식한 건수를 반환합니다. code가 400으로 반환되면 지원하지 않는 형식이므로 사용자에게 알리고 아래의 outside_report API를 이용하여 보고해주시면 됩니다.

Resource URL
https://whooing.com/api/entries/outside.format

Parameters(purple color is required)
section_id
섹션의 고유번호 Example Value : s199
rows
외부데이터 내용 Example Value :
우리 11/13 12:22
*29372
지급 17,000원
주）흥반점
잔액 922,918원
우리 11/21 14:13
*99172
지급 12,220원
AMAZON INC
잔액 110,028,100원
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
"cnt" : 2
}
}
방금 외부입력을 시도했다가 인식되지 않은 것에 대해서 소스를 보고합니다.

Resource URL
https://whooing.com/api/entries/outside_report.format

Parameters(purple color is required)
source
데이터의 소스명
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : "success"
}
