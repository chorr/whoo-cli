package api

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"whoo-cli/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func newTestClient(t *testing.T, transport roundTripFunc) *WhooingClient {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("WHOOING_APP_ID", "test-app")
	t.Setenv("WHOOING_APP_SECRET", "test-secret")
	c := NewWhooingClient(&config.Config{Token: "test-token", TokenSecret: "test-token-secret"})
	c.httpClient.Transport = transport
	return c
}

func testResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}
}

func TestDoRequestOnceEncoding(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodDelete, http.MethodPost, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			params := url.Values{"section_id": {"s1"}, "item": {"점심 & 커피=+"}, "memo": {"첫 줄\n둘째 줄"}}
			before := time.Now().UnixMilli()
			body := `{"code":200,"results":{"unknown":"preserved"}}`
			calls := 0
			c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != method || r.URL.Scheme != "https" || r.URL.Host != "whooing.com" ||
					r.URL.Path != "/api/entries.json" {
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL)
				}
				if r.Header.Get("Accept") != "application/json" {
					t.Fatalf("Accept = %q", r.Header.Get("Accept"))
				}
				key := r.Header.Get("X-API-KEY")
				signiture := (&config.Config{TokenSecret: "test-token-secret"}).ComputeSigniture()
				prefix := "app_id=test-app,token=test-token,signiture=" + signiture + ",timestamp="
				if !strings.HasPrefix(key, prefix) {
					t.Fatalf("unexpected API key format: %q", key)
				}
				ts, err := strconv.ParseInt(strings.TrimPrefix(key, prefix), 10, 64)
				if err != nil || ts < before || ts > time.Now().UnixMilli() {
					t.Fatalf("invalid millisecond timestamp: %q", key)
				}
				got := r.URL.Query()
				if method == http.MethodPost || method == http.MethodPut {
					if r.URL.RawQuery != "" {
						t.Fatalf("form fields leaked to query: %s", r.URL)
					}
					if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded;charset=UTF-8" {
						t.Fatalf("Content-Type = %q", r.Header.Get("Content-Type"))
					}
					if err := r.ParseForm(); err != nil {
						t.Fatal(err)
					}
					got = r.PostForm
				} else if r.Body != nil {
					t.Fatal("GET/DELETE must send parameters in query")
				}
				if !reflect.DeepEqual(got, params) {
					t.Fatalf("parameters = %v, want %v", got, params)
				}
				return testResponse(http.StatusOK, body), nil
			})
			got, status, err := c.doRequestOnce(method, "/entries.json", params)
			if err != nil || status != http.StatusOK || string(got) != body || calls != 1 {
				t.Fatalf("response = %s, %d, %v; calls = %d", got, status, err, calls)
			}
		})
	}
}

func TestDoRequestOnceHTTPErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		reason string
	}{
		{name: "daily limit", status: 402, body: `{}`, reason: "daily_limit"},
		{name: "expired token", status: 405, body: `{}`, reason: "token_expired"},
		{name: "rate limit", status: 429, body: `{}`, reason: "rate_limit"},
		{name: "API message", status: 400, body: `{"message":"bad input"}`, reason: "bad_request"},
		{name: "HTML firewall error", status: 403, body: `<html>blocked</html>`, reason: "forbidden"},
		{name: "server error", status: 500, body: `<html>down</html>`, reason: "server_error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
				return testResponse(tt.status, tt.body), nil
			})
			data, status, err := c.doRequestOnce(http.MethodGet, "/user.json", nil)
			var apiErr *APIError
			if data != nil || status != tt.status || !errors.As(err, &apiErr) {
				t.Fatalf("response = %s, %d, %v", data, status, err)
			}
			if apiErr.Code != tt.status || apiErr.Reason != tt.reason || apiErr.Endpoint != "/user.json" {
				t.Fatalf("APIError = %+v", apiErr)
			}
			if tt.status == 400 && apiErr.Message != "bad input" {
				t.Fatalf("API message lost: %v", apiErr)
			}
			if strings.Contains(apiErr.Error(), "<html>") {
				t.Fatalf("HTML leaked into normal error: %v", apiErr)
			}
			c.config.Verbose = true
			_, _, err = c.doRequestOnce(http.MethodGet, "/user.json", nil)
			if !errors.As(err, &apiErr) || !strings.Contains(apiErr.Error(), tt.body) {
				t.Fatalf("verbose error lost response details: %v", err)
			}
		})
	}
}

type failingBody struct {
	err    error
	closed bool
}

func (b *failingBody) Read([]byte) (int, error) { return 0, b.err }
func (b *failingBody) Close() error {
	b.closed = true
	return nil
}

func TestDoRequestOnceIOErrors(t *testing.T) {
	sentinel := errors.New("test I/O error")
	t.Run("transport", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) { return nil, sentinel })
		_, _, err := c.doRequestOnce(http.MethodGet, "/user.json", nil)
		if !errors.Is(err, sentinel) {
			t.Fatalf("transport error chain lost: %v", err)
		}
	})
	t.Run("read and close", func(t *testing.T) {
		body := &failingBody{err: sentinel}
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: body}, nil
		})
		_, status, err := c.doRequestOnce(http.MethodGet, "/user.json", nil)
		if !errors.Is(err, sentinel) || status != 200 || !body.closed {
			t.Fatalf("read error = %v, status = %d, closed = %v", err, status, body.closed)
		}
	})
}

func TestDoRequestRetries(t *testing.T) {
	tests := []struct {
		name      string
		statuses  []int
		wantCalls int
		wantWaits int
		wantCode  int
	}{
		{name: "success", statuses: []int{200}, wantCalls: 1},
		{name: "daily limit stops", statuses: []int{402}, wantCalls: 1, wantCode: 402},
		{name: "expired token stops", statuses: []int{405}, wantCalls: 1, wantCode: 405},
		{name: "bad request stops", statuses: []int{400}, wantCalls: 1, wantCode: 400},
		{name: "rate limit recovers", statuses: []int{429, 200}, wantCalls: 2, wantWaits: 1},
		{name: "rate limit exhausted", statuses: []int{429}, wantCalls: 3, wantWaits: 2, wantCode: 429},
		{name: "server recovers", statuses: []int{500, 200}, wantCalls: 2, wantWaits: 1},
		{name: "server retries once", statuses: []int{500}, wantCalls: 2, wantWaits: 1, wantCode: 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls, waits := 0, 0
			c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
				index := calls
				if index >= len(tt.statuses) {
					index = len(tt.statuses) - 1
				}
				calls++
				return testResponse(tt.statuses[index], `{"code":200,"results":{}}`), nil
			})
			c.retryWait = func(d time.Duration) {
				waits++
				if d != 3*time.Second {
					t.Fatalf("retry delay = %s, want 3s", d)
				}
			}
			_, err := c.GetUser()
			if calls != tt.wantCalls || waits != tt.wantWaits {
				t.Fatalf("calls = %d, waits = %d; want %d, %d", calls, waits, tt.wantCalls, tt.wantWaits)
			}
			if tt.wantCode == 0 {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.Code != tt.wantCode {
				t.Fatalf("retry error = %v, want APIError %d", err, tt.wantCode)
			}
		})
	}
}

func TestParseResponse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		body    string
		wantErr bool
		wantID  int
	}{
		{name: "success", body: `{"code":200,"results":{"entry_id":42}}`, wantID: 42},
		{name: "no content", body: `{"code":204}`},
		{name: "API error", body: `{"code":400,"error_parameters":{"item":"required"}}`, wantErr: true},
		{name: "malformed envelope", body: `{`, wantErr: true},
		{name: "wrong results type", body: `{"code":200,"results":[]}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var entry Entry
			err := parseResponse([]byte(tt.body), &entry)
			if (err != nil) != tt.wantErr || entry.EntryID != tt.wantID {
				t.Fatalf("parseResponse() = %+v, %v", entry, err)
			}
		})
	}
}

func TestGetAccountsMapCacheIsScopedBySection(t *testing.T) {
	calls := 0
	c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		calls++
		return testResponse(200, `{"code":200,"results":{"assets":{"a1":{"title":"현금"}}}}`), nil
	})
	for _, section := range []string{"s1", "s1", "s2"} {
		accounts, err := c.GetAccountsMap(section)
		if err != nil || accounts.GetTitle("assets", "a1") != "현금" {
			t.Fatalf("GetAccountsMap(%q) = %+v, %v", section, accounts, err)
		}
	}
	if calls != 2 {
		t.Fatalf("requests = %d, want one per section", calls)
	}
	c.InvalidateCache("accounts:s1")
	if _, err := c.GetAccountsMap("s1"); err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatalf("requests after invalidation = %d, want 3", calls)
	}
}

func TestGetEntriesRequestAndResponse(t *testing.T) {
	tests := []struct {
		name   string
		limit  int
		cursor string
		body   string
		err    bool
	}{
		{
			name: "API defaults",
			body: `{"code":200,"rest_of_api":99,"results":{"rows":[{"entry_id":42,"entry_date":"20261029.0017"}]}}`,
		},
		{
			name: "exact pagination cursor", limit: 20, cursor: "20261029.0017",
			body: `{"code":200,"rest_of_api":99,"results":{"rows":[{"entry_id":42,"entry_date":"20261029.0017"}]}}`,
		},
		{name: "API error", body: `{"code":400,"message":"invalid date"}`, err: true},
		{name: "invalid rows shape", body: `{"code":200,"results":{"rows":{}}}`, err: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
				calls++
				want := url.Values{"section_id": {"s1"}, "start_date": {"20261001"}, "end_date": {"20261029"}}
				if tt.limit > 0 {
					want.Set("limit", strconv.Itoa(tt.limit))
				}
				if tt.cursor != "" {
					want.Set("max", tt.cursor)
				}
				if r.Method != "GET" || r.URL.Path != "/api/entries.json_array" || !reflect.DeepEqual(r.URL.Query(), want) {
					t.Fatalf("entries request = %s %s, want query %v", r.Method, r.URL, want)
				}
				return testResponse(200, tt.body), nil
			})
			rows, err := c.GetEntries("s1", "20261001", "20261029", tt.limit, tt.cursor)
			if (err != nil) != tt.err || calls != 1 {
				t.Fatalf("GetEntries() = %+v, %v; calls = %d", rows, err, calls)
			}
			if !tt.err && (len(rows) != 1 || rows[0].EntryID != 42 || rows[0].EntryDate != "20261029.0017") {
				t.Fatalf("entries response lost fields: %+v", rows)
			}
		})
	}
}
