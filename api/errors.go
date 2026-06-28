// api/errors.go
// API 에러 타입 정의

package api

import (
	"encoding/json"
	"fmt"
)

// APIError는 후잉 API 에러 정보
type APIError struct {
	Code       int
	Message    string
	Parameters json.RawMessage
	Endpoint   string
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = "알 수 없는 오류"
	}
	if e.Endpoint != "" {
		return fmt.Sprintf("API 오류 (code=%d, endpoint=%s): %s", e.Code, e.Endpoint, msg)
	}
	return fmt.Sprintf("API 오류 (code=%d): %s", e.Code, msg)
}

// IsRateLimit는 429 응답인지 확인
func (e *APIError) IsRateLimit() bool { return e.Code == 429 }

// IsDailyLimit는 402 응답인지 확인
func (e *APIError) IsDailyLimit() bool { return e.Code == 402 }

// IsTokenExpired는 405 응답인지 확인 (토큰 만료)
func (e *APIError) IsTokenExpired() bool { return e.Code == 405 }
