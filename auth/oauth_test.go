package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"whoo-cli/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestOAuthTokenRequests(t *testing.T) {
	tests := []struct {
		name      string
		exchange  bool
		signiture string
		status    int
		body      string
		wantErr   bool
	}{
		{name: "request token", status: 200, body: `{"token":"temporary"}`},
		{name: "exchange", exchange: true, status: 200, body: `{"token":"access","token_secret":"secret"}`},
		{
			name: "legacy exchange", exchange: true, signiture: "legacy",
			status: 200, body: `{"token":"access","token_secret":"secret"}`,
		},
		{name: "request HTTP failure", status: 503, body: `unavailable`, wantErr: true},
		{name: "exchange HTTP failure", exchange: true, status: 503, body: `unavailable`, wantErr: true},
		{name: "request API error", status: 200, body: `{"code":400,"message":"bad app"}`, wantErr: true},
		{
			name: "exchange API error", exchange: true,
			status: 200, body: `{"code":405,"message":"bad pin"}`, wantErr: true,
		},
		{name: "request malformed JSON", status: 200, body: `{`, wantErr: true},
		{name: "exchange malformed JSON", exchange: true, status: 200, body: `{`, wantErr: true},
		{name: "request missing token", status: 200, body: `{}`, wantErr: true},
		{
			name: "exchange missing secret", exchange: true,
			status: 200, body: `{"token":"access"}`, wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("WHOOING_APP_ID", "test-app")
			t.Setenv("WHOOING_APP_SECRET", "test & secret=+")
			cfg := &config.Config{}
			o := NewOAuth(cfg)
			calls := 0
			o.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				path := "/app_auth/request_token"
				want := url.Values{"app_id": {"test-app"}, "app_secret": {"test & secret=+"}}
				if tt.exchange {
					path = "/app_auth/access_token"
					want.Set("token", "temp & token")
					want.Set("pin", "123 456")
					if tt.signiture != "" {
						want.Set("signiture", tt.signiture)
					}
				}
				if r.Method != http.MethodGet || r.URL.Host != "whooing.com" || r.URL.Path != path {
					t.Fatalf("unexpected OAuth request: %s %s", r.Method, r.URL)
				}
				if !reflect.DeepEqual(r.URL.Query(), want) {
					t.Fatalf("query = %v, want %v", r.URL.Query(), want)
				}
				return &http.Response{
					StatusCode: tt.status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(tt.body)),
				}, nil
			})
			var err error
			if tt.exchange {
				var result *AccessTokenResponse
				result, err = o.ExchangeToken("temp & token", tt.signiture, "123 456")
				if !tt.wantErr && (result == nil || result.Token != "access" || result.TokenSecret != "secret") {
					t.Fatalf("exchange result = %+v", result)
				}
			} else {
				var result *RequestTokenResponse
				result, err = o.RequestToken()
				if !tt.wantErr && (result == nil || result.Token != "temporary") {
					t.Fatalf("request result = %+v", result)
				}
			}
			if (err != nil) != tt.wantErr || calls != 1 {
				t.Fatalf("error = %v, calls = %d", err, calls)
			}
			if cfg.IsAuthenticated() {
				t.Fatal("token request must not persist authentication before completion")
			}
		})
	}
}

func TestOAuthGetAuthorizationURL(t *testing.T) {
	t.Parallel()
	o := NewOAuth(&config.Config{})
	u, err := url.Parse(o.GetAuthorizationURL("temp & token=+"))
	if err != nil {
		t.Fatal(err)
	}
	if u.Scheme != "https" || u.Host != "whooing.com" || u.Path != "/app_auth/authorize" ||
		u.Query().Get("token") != "temp & token=+" {
		t.Fatalf("authorization URL = %s", u)
	}
}

func TestOAuthCompleteAuth(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{SectionID: "s1"}
	o := NewOAuth(cfg)
	if err := o.SavePendingRequest("temporary", "legacy"); err != nil {
		t.Fatal(err)
	}
	pending, err := config.Load()
	if err != nil || pending.PendingRequestToken != "temporary" || pending.PendingSigniture != "legacy" {
		t.Fatalf("pending config = %+v, %v", pending, err)
	}
	if err := o.CompleteAuth("access", "access-secret"); err != nil {
		t.Fatal(err)
	}
	got, err := config.Load()
	want := &config.Config{Token: "access", TokenSecret: "access-secret", SectionID: "s1"}
	if err != nil || !reflect.DeepEqual(got, want) || !reflect.DeepEqual(cfg, want) {
		t.Fatalf("completed config: memory = %+v, disk = %+v, err = %v", cfg, got, err)
	}
}

func TestOAuthRejectsEmptyTokens(t *testing.T) {
	tests := []struct {
		name  string
		apply func(*OAuth) error
	}{
		{name: "empty pending", apply: func(o *OAuth) error { return o.SavePendingRequest("", "") }},
		{name: "empty access", apply: func(o *OAuth) error { return o.CompleteAuth("", "secret") }},
		{name: "empty secret", apply: func(o *OAuth) error { return o.CompleteAuth("token", "") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			cfg := &config.Config{Token: "original", TokenSecret: "original-secret"}
			if err := cfg.Save(); err != nil {
				t.Fatal(err)
			}
			before := *cfg
			if err := tt.apply(NewOAuth(cfg)); err == nil {
				t.Fatal("empty token accepted")
			}
			saved, err := config.Load()
			if err != nil || !reflect.DeepEqual(saved, &before) || !reflect.DeepEqual(cfg, &before) {
				t.Fatalf("rejected request changed config: memory = %+v, disk = %+v, err = %v", cfg, saved, err)
			}
		})
	}
}

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
