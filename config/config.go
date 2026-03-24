// config/config.go
// 설정 파일 저장/로드 기능을 제공하는 패키지

package config

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// 패키지 레벨 변수 - 빌드 시 ldflags로 주입 가능
// 예: go build -ldflags "-X whooing-cli/config.AppID=xxx -X whooing-cli/config.AppSecret=yyy" -o whooing .
var (
	// AppID는 후잉 앱 ID (빌드 시 -ldflags로 주입, 또는 환경변수)
	AppID string
	// AppSecret은 후잉 앱 Secret (빌드 시 -ldflags로 주입, 또는 환경변수)
	AppSecret string
)

// GetAppID는 AppID를 반환
// 1. 빌드 시 주입된 값이 있으면 그대로 사용
// 2. 비어있으면 환경변수 WHOOING_APP_ID에서 읽기
// 3. 둘 다 없으면 에러 반환
func GetAppID() (string, error) {
	// 빌드 시 주입된 값 확인
	if AppID != "" {
		return AppID, nil
	}

	// 환경변수에서 읽기
	id := os.Getenv("WHOOING_APP_ID")
	if id != "" {
		return id, nil
	}

	return "", fmt.Errorf("app_id가 설정되지 않았습니다. WHOOING_APP_ID 환경변수를 설정하거나, -ldflags로 빌드하세요")
}

// GetAppSecret는 AppSecret을 반환
// 1. 빌드 시 주입된 값이 있으면 그대로 사용
// 2. 비어있으면 환경변수 WHOOING_APP_SECRET에서 읽기
// 3. 둘 다 없으면 에러 반환
func GetAppSecret() (string, error) {
	// 빌드 시 주입된 값 확인
	if AppSecret != "" {
		return AppSecret, nil
	}

	// 환경변수에서 읽기
	secret := os.Getenv("WHOOING_APP_SECRET")
	if secret != "" {
		return secret, nil
	}

	return "", fmt.Errorf("app_secret이 설정되지 않았습니다. WHOOING_APP_SECRET 환경변수를 설정하거나, -ldflags로 빌드하세요")
}

// Config는 후잉 CLI의 설정을 저장하는 구조체
// config.json에는 token, token_secret, section_id 저장 (app_id, app_secret 제외)
type Config struct {
	Token       string `json:"token"`
	TokenSecret string `json:"token_secret"`
	SectionID   string `json:"section_id"`
}

// ComputeSigniture는 API 호출용 signiture를 계산
func (c *Config) ComputeSigniture() string {
	appSecret, err := GetAppSecret()
	if err != nil {
		return ""
	}
	if c.TokenSecret == "" {
		return ""
	}

	hash := sha1.New()
	hash.Write([]byte(appSecret + "|" + c.TokenSecret))
	return hex.EncodeToString(hash.Sum(nil))
}

// getConfigPath는 설정 파일 경로를 반환
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("홈 디렉토리 확인 실패: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "whooing-cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("설정 디렉토리 생성 실패: %w", err)
	}

	return filepath.Join(configDir, "config.json"), nil
}

// Load는 설정 파일에서 설정을 로드
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 설정 파일이 없으면 빈 설정 반환
			return &Config{}, nil
		}
		return nil, fmt.Errorf("설정 파일 읽기 실패: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("설정 파일 파싱 실패: %w", err)
	}

	return &cfg, nil
}

// Save는 현재 설정을 파일에 저장
func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("설정 직렬화 실패: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("설정 파일 저장 실패: %w", err)
	}

	return nil
}

// IsAuthenticated는 인증 완료 여부 확인
// token과 token_secret이 있어야 인증 완료로 판단
func (c *Config) IsAuthenticated() bool {
	return c.Token != "" && c.TokenSecret != ""
}

// LoadWithTestTokens는 설정을 로드하고 테스트 토큰이 있으면 적용
// 환경변수 WHOOING_TEST_TOKEN, WHOOING_TEST_TOKEN_SECRET가 있으면
// OAuth 없이 바로 API 호출 가능 (개발/테스트용)
func LoadWithTestTokens() (*Config, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	// 테스트 토큰 환경변수 확인
	testToken := os.Getenv("WHOOING_TEST_TOKEN")
	testTokenSecret := os.Getenv("WHOOING_TEST_TOKEN_SECRET")

	// 이미 저장된 토큰이 없고 테스트 토큰이 있으면 적용
	if !cfg.IsAuthenticated() && testToken != "" && testTokenSecret != "" {
		cfg.Token = testToken
		cfg.TokenSecret = testTokenSecret
	}

	return cfg, nil
}

