User
유저에 관한 기본 정보를 요청합니다.

Resource URL
https://whooing.com/api/user.format

Parameters(purple color is required)

-

Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
"user_id" : 4,
"username" : "Helloman",
"last_ip" : "192.168.0.1",
"last_login_timestamp" : 1322448931,
"created_timestamp" : 1321448931,
"modified_timestamp" : 1321448931,
"language" : "ko",
"level" : "1",
"expire" : 1321448931,
"timezone" : "Asia/Seoul",
"currency" : "KRW",
"country" : "KR",
"image_url" : "https://static.whooing.com/profiles/p14.jpg",
"mileage" : 230
}
}
항목의 정보를 수정합니다. 수정이 되는 파라미터들만 전송하면 됩니다.

Resource URL
https://whooing.com/api/user.format

Parameters(purple color is required)
username
사용자의 별명 Example Value : 흥반장
country
사용자의 국가 Example Value : KR
language
사용자의 언어 Example Value : ko
timezone
사용자의 타임존 Example Value : Asia/Seoul
currency
사용자의 기본 통화단위 Example Value : KRW
Example Results(.json, .json_array)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : {
"user_id" : 4,
"username" : "Helloman",
"last_ip" : "192.168.0.1",
"last_login_timestamp" : 1322448931,
"created_timestamp" : 1321448931,
"modified_timestamp" : 1321448931,
"language" : "ko",
"level" : "1",
"expire" : 1321448931,
"timezone" : "Asia/Seoul",
"currency" : "KRW",
"country" : "KR"
}
}
유저의 로그리스트를 요청합니다.

Resource URL
https://whooing.com/api/user_logs.format

Parameters(purple color is required)
max
id가 max보다 작은 로그로만(즉, max보다 이전으로만) 제한 Example Value : 1297
limit
표시할 로그의 갯수 Example Value : 20
Example Results(JSON)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : [
{
"id": 483915,
"contents": "`Def 234` 섹션을 수정",
"datetime": 1618497070,
"ip": "182.172.164.88",
"segment0": "sections",
"segment1": "",
"writer": "user",
},
.
.
.
]
}
유저의 포인트 로그리스트를 요청합니다.

Resource URL
https://whooing.com/api/user_point_logs.format

Parameters(purple color is required)
max
point_id가 max보다 작은 포인트 로그로만(즉, max보다 이전으로만) 제한 Example Value : 1297
type
포인트의 종류, 기본은 all, 추천인을 통한 가입 리스트는 affiliate. Example Value : all
limit
표시할 포인트 로그의 갯수 Example Value : 20
Example Results(JSON)
{
"code" : 200,
"message" : "",
"error_parameters" : {},
"rest_of_api" : 4988,
"results" : [
{
"point_id": 144968,
"datetime": 1621760410,
"description": "일일 방문",
"point": 50,
"writer": "user",
},
.
.
.
]
}
