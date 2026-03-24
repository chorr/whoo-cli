// auth/oauth.go
// Whooing OAuth PIN 방식 인증 플로우 처리

package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"whooing-cli/config"
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
type RequestTokenResponse struct {
	Token     string `json:"token"`
	Signiture string `json:"signiture"` // 원문 그대로
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

	url := fmt.Sprintf("%s/app_auth/request_token?app_id=%s&app_secret=%s",
		baseURL, appID, appSecret)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("토큰 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("토큰 요청 실패: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	var result RequestTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w", err)
	}

	return &result, nil
}

// GetAuthorizationURL는 2단계: 사용자가 접속할 인증 URL 반환
func (o *OAuth) GetAuthorizationURL(token string) string {
	return fmt.Sprintf("%s/app_auth/authorize?token=%s", baseURL, token)
}

// ExchangeToken는 3단계: PIN으로 최종 토큰 교환
// https://whooing.com/app_auth/access_token?app_id={}&app_secret={}&token={}&signiture={}&pin={}
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

	url := fmt.Sprintf("%s/app_auth/access_token?app_id=%s&app_secret=%s&token=%s&signiture=%s&pin=%s",
		baseURL, appID, appSecret, tempToken, signiture, pin)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("토큰 교환 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	// API 오류 응답 체크
	var apiCheck map[string]interface{}
	if err := json.Unmarshal(body, &apiCheck); err == nil {
		if code, ok := apiCheck["code"].(float64); ok && code != 200 {
			message := ""
			if msg, ok := apiCheck["message"].(string); ok {
				message = msg
			}
			return nil, fmt.Errorf("API 오류 (code=%d): %s", int(code), message)
		}
	}

	var result AccessTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w", err)
	}

	return &result, nil
}

// CompleteAuth는 전체 인증 플로우 완료 후 설정 저장
// token과 token_secret을 저장 (signiture는 ComputeSigniture()로 계산)
func (o *OAuth) CompleteAuth(token, tokenSecret string) error {
	o.config.Token = token
	o.config.TokenSecret = tokenSecret
	return o.config.Save()
}
