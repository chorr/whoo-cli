// auth/oauth.go
// Whooing OAuth PIN 방식 인증 플로우 처리

package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"whoo-cli/config"
)

const (
	baseURL = "https://whooing.com"
)

// OAuth는 후잉 OAuth 인증을 처리하는 구조체
type OAuth struct {
	config *config.Config
}

// NewOAuth는 새 OAuth 인스턴스 생성
func NewOAuth(cfg *config.Config) *OAuth {
	return &OAuth{config: cfg}
}

// RequestTokenResponse는 1단계 요청 토큰 응답
// 현재 Whooing API는 token만 반환한다. signiture는 과거 응답 호환용(선택).
type RequestTokenResponse struct {
	Token     string `json:"token"`
	Signiture string `json:"signiture"` // 원문 그대로. 미반환 시 빈 문자열
}

// AccessTokenResponse는 3단계 액세스 토큰 응답
// 실제 API는 token과 token_secret을 반환
type AccessTokenResponse struct {
	Token       string `json:"token"`
	TokenSecret string `json:"token_secret"` // API 응답 필드명
}

// RequestToken는 1단계: 요청 토큰 획득
// https://whooing.com/app_auth/request_token?app_id={}&app_secret={}
func (o *OAuth) RequestToken() (*RequestTokenResponse, error) {
	// AppID와 AppSecret 가져오기
	appID, err := config.GetAppID()
	if err != nil {
		return nil, err
	}
	appSecret, err := config.GetAppSecret()
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("app_id", appID)
	q.Set("app_secret", appSecret)
	reqURL := fmt.Sprintf("%s/app_auth/request_token?%s", baseURL, q.Encode())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("토큰 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("토큰 요청 실패: HTTP %d: %s", resp.StatusCode, truncateBody(body))
	}

	// API 오류 응답 체크 (code 필드가 있는 경우)
	if err := checkOAuthAPIError(body); err != nil {
		return nil, err
	}

	var result RequestTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w (body=%s)", err, truncateBody(body))
	}
	// token만 필수. signiture는 현재 API에서 반환하지 않음
	if result.Token == "" {
		return nil, fmt.Errorf("토큰 응답이 비어 있습니다: %s", truncateBody(body))
	}

	return &result, nil
}

// GetAuthorizationURL는 2단계: 사용자가 접속할 인증 URL 반환
func (o *OAuth) GetAuthorizationURL(token string) string {
	q := url.Values{}
	q.Set("token", token)
	return fmt.Sprintf("%s/app_auth/authorize?%s", baseURL, q.Encode())
}

// ExchangeToken는 3단계: PIN으로 최종 토큰 교환
// https://whooing.com/app_auth/access_token?app_id={}&app_secret={}&token={}&pin={}
// signiture는 과거 API 호환용 선택 파라미터
func (o *OAuth) ExchangeToken(tempToken, signiture, pin string) (*AccessTokenResponse, error) {
	// AppID와 AppSecret 가져오기
	appID, err := config.GetAppID()
	if err != nil {
		return nil, err
	}
	appSecret, err := config.GetAppSecret()
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("app_id", appID)
	q.Set("app_secret", appSecret)
	q.Set("token", tempToken)
	// 현재 request_token 응답에 signiture가 없으므로 있을 때만 전달
	if signiture != "" {
		q.Set("signiture", signiture)
	}
	q.Set("pin", pin)
	reqURL := fmt.Sprintf("%s/app_auth/access_token?%s", baseURL, q.Encode())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("토큰 교환 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("토큰 교환 실패: HTTP %d: %s", resp.StatusCode, truncateBody(body))
	}

	// API 오류 응답 체크
	if err := checkOAuthAPIError(body); err != nil {
		return nil, err
	}

	var result AccessTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w (body=%s)", err, truncateBody(body))
	}
	if result.Token == "" || result.TokenSecret == "" {
		return nil, fmt.Errorf("액세스 토큰 응답이 비어 있습니다 (PIN을 확인하세요): %s", truncateBody(body))
	}

	return &result, nil
}

// checkOAuthAPIError는 OAuth 엔드포인트의 {code, message} 오류 응답을 검사한다.
// code 필드가 없거나 200이면 nil.
func checkOAuthAPIError(body []byte) error {
	var apiCheck map[string]interface{}
	if err := json.Unmarshal(body, &apiCheck); err != nil {
		return nil
	}
	code, ok := apiCheck["code"].(float64)
	if !ok || code == 200 {
		return nil
	}
	message := ""
	if msg, ok := apiCheck["message"].(string); ok {
		message = msg
	}
	return fmt.Errorf("API 오류 (code=%d): %s", int(code), message)
}

// truncateBody는 에러 메시지용으로 응답 본문을 짧게 자른다.
func truncateBody(body []byte) string {
	const max = 200
	s := string(body)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// CompleteAuth는 전체 인증 플로우 완료 후 설정 저장
// token과 token_secret을 저장 (signiture는 ComputeSigniture()로 계산)
func (o *OAuth) CompleteAuth(token, tokenSecret string) error {
	if token == "" || tokenSecret == "" {
		return fmt.Errorf("빈 토큰은 저장할 수 없습니다")
	}
	o.config.Token = token
	o.config.TokenSecret = tokenSecret
	return o.config.Save()
}
