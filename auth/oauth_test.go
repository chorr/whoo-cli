package auth

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRequestTokenResponse_tokenOnly(t *testing.T) {
	// 현재 Whooing API: {"token":"..."} 만 반환 (signiture 없음)
	body := []byte(`{"token":"temp-token-abc"}`)
	var result RequestTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.Token == "" {
		t.Fatal("token should be present")
	}
	if result.Signiture != "" {
		t.Fatalf("signiture should be empty for current API, got %q", result.Signiture)
	}

	// 구 API 응답 호환: token + signiture
	bodyLegacy := []byte(`{"token":"t","signiture":"s"}`)
	var legacy RequestTokenResponse
	if err := json.Unmarshal(bodyLegacy, &legacy); err != nil {
		t.Fatalf("unmarshal legacy: %v", err)
	}
	if legacy.Token != "t" || legacy.Signiture != "s" {
		t.Fatalf("legacy parse failed: %+v", legacy)
	}
}

func TestCheckOAuthAPIError(t *testing.T) {
	if err := checkOAuthAPIError([]byte(`{"token":"x"}`)); err != nil {
		t.Fatalf("no code field should be ok: %v", err)
	}
	if err := checkOAuthAPIError([]byte(`{"code":200,"message":"ok"}`)); err != nil {
		t.Fatalf("code 200 should be ok: %v", err)
	}
	err := checkOAuthAPIError([]byte(`{"code":405,"message":"bad pin"}`))
	if err == nil || !strings.Contains(err.Error(), "405") {
		t.Fatalf("expected API error, got %v", err)
	}
}

func TestTruncateBody(t *testing.T) {
	if got := truncateBody([]byte("hello")); got != "hello" {
		t.Fatalf("got %q", got)
	}
	long := strings.Repeat("a", 250)
	got := truncateBody([]byte(long))
	if !strings.HasSuffix(got, "...") || len(got) != 203 {
		t.Fatalf("unexpected truncate: len=%d %q", len(got), got)
	}
}
