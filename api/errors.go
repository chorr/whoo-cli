// api/errors.go
// API 에러 타입 정의

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// APIError는 후잉 API 에러 정보
type APIError struct {
	Code       int             `json:"code"`
	Endpoint   string          `json:"endpoint,omitempty"`
	Reason     string          `json:"reason"`
	Message    string          `json:"message"`
	Parameters json.RawMessage `json:"parameters,omitempty"`
	Details    string          `json:"-"`
	Verbose    bool            `json:"-"`
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = "알 수 없는 오류"
	}
	reason := e.Reason
	if reason == "" {
		reason = reasonForStatus(e.Code)
	}

	prefix := fmt.Sprintf("API 오류 (code=%d", e.Code)
	if e.Endpoint != "" {
		prefix += fmt.Sprintf(", endpoint=%s", e.Endpoint)
	}
	if reason != "" {
		prefix += fmt.Sprintf(", reason=%s", reason)
	}
	prefix += "): " + msg

	// 토큰 만료 시 재인증 안내
	if e.IsTokenExpired() {
		prefix += " — whoo auth 로 재인증하세요"
	}
	if e.Verbose && e.Details != "" {
		prefix += "\n응답 본문: " + e.Details
	}
	return prefix
}

// IsRateLimit는 429 응답인지 확인
func (e *APIError) IsRateLimit() bool { return e.Code == 429 }

// IsDailyLimit는 402 응답인지 확인
func (e *APIError) IsDailyLimit() bool { return e.Code == 402 }

// IsTokenExpired는 405 응답인지 확인 (토큰 만료)
func (e *APIError) IsTokenExpired() bool { return e.Code == 405 }

// reasonForStatus는 HTTP/API 코드를 기계가 분기하기 쉬운 짧은 값으로 변환한다.
func reasonForStatus(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusPaymentRequired:
		return "daily_limit"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusMethodNotAllowed:
		return "token_expired"
	case http.StatusConflict:
		return "conflict"
	case http.StatusTooManyRequests:
		return "rate_limit"
	default:
		if code >= 500 {
			return "server_error"
		}
		return "api_error"
	}
}

// shortHTTPMessage는 HTML 응답을 노출하지 않고 상태별 짧은 설명을 반환한다.
func shortHTTPMessage(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "요청 파라미터가 올바르지 않습니다"
	case http.StatusUnauthorized:
		return "API 요청 권한이 없습니다"
	case http.StatusForbidden:
		return "API 요청이 거부되었습니다"
	case http.StatusNotFound:
		return "API 엔드포인트를 찾을 수 없습니다"
	case http.StatusConflict:
		return "요청이 현재 상태와 충돌합니다"
	default:
		if code >= 500 {
			return "후잉 API 서버 오류입니다"
		}
		return fmt.Sprintf("HTTP %d 오류", code)
	}
}

// compactDetails는 --verbose 출력도 과도하게 커지지 않도록 본문을 제한한다.
func compactDetails(body []byte) string {
	const max = 4096
	details := strings.TrimSpace(string(body))
	if len(details) > max {
		details = details[:max] + "…"
	}
	return details
}
