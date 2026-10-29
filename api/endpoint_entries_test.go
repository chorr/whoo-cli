package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestCreateEntriesBatch(t *testing.T) {
	for _, count := range []int{0, 1, 300, 301, 601} {
		t.Run(fmt.Sprintf("%d rows", count), func(t *testing.T) {
			rows := make([]EntryInput, count)
			for i := range rows {
				rows[i] = EntryInput{LAccount: "expenses", RAccount: "assets", Money: int64(i + 1), Item: "점심 & 커피"}
			}
			gotRows := make([]EntryInput, 0, count)
			calls := 0
			lastBody := ""
			c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != http.MethodPost || r.URL.Path != "/api/entries.json" {
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL)
				}
				if err := r.ParseForm(); err != nil {
					t.Fatal(err)
				}
				if r.PostForm.Get("section_id") != "s1" || r.PostForm.Get("data_type") != "json" {
					t.Fatalf("unexpected form: %v", r.PostForm)
				}
				chunk := []EntryInput{}
				if err := json.Unmarshal([]byte(r.PostForm.Get("entries")), &chunk); err != nil {
					t.Fatal(err)
				}
				if len(chunk) == 0 || len(chunk) > 300 {
					t.Fatalf("chunk size = %d", len(chunk))
				}
				gotRows = append(gotRows, chunk...)
				lastBody = fmt.Sprintf(`{"code":200,"rest_of_api":%d,"results":{}}`, 100-calls)
				return testResponse(http.StatusOK, lastBody), nil
			})
			data, err := c.CreateEntriesBatch("s1", rows)
			if err != nil {
				t.Fatal(err)
			}
			if calls != (count+299)/300 || !reflect.DeepEqual(gotRows, rows) {
				t.Fatalf("calls = %d; sent rows do not match input: %v", calls, gotRows)
			}
			if string(data) != lastBody {
				t.Fatalf("response = %s, want last response %s", data, lastBody)
			}
			if count > 0 && c.LastRestOfAPI() != 100-calls {
				t.Fatalf("remaining quota = %d, want %d", c.LastRestOfAPI(), 100-calls)
			}
		})
	}
}

func TestCreateEntriesBatchStopsOnError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		failOn int
	}{
		{name: "HTTP error", status: 400, body: `{"message":"bad input"}`, failOn: 1},
		{name: "API error in HTTP success", status: 200, body: `{"code":400,"message":"bad input"}`, failOn: 1},
		{name: "malformed response", status: 200, body: `{`, failOn: 1},
		{name: "later chunk API error", status: 200, body: `{"code":400,"message":"bad input"}`, failOn: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
				calls++
				if calls < tt.failOn {
					return testResponse(200, `{"code":200,"results":{}}`), nil
				}
				return testResponse(tt.status, tt.body), nil
			})
			data, err := c.CreateEntriesBatch("s1", make([]EntryInput, 601))
			if err == nil || data != nil || calls != tt.failOn {
				t.Fatalf("batch = %s, %v; calls = %d, want error and %d calls", data, err, calls, tt.failOn)
			}
			chunkContext := fmt.Sprintf("chunk %d~%d", (tt.failOn-1)*300, tt.failOn*300-1)
			if !strings.Contains(err.Error(), chunkContext) {
				t.Fatalf("missing failing chunk context: %v", err)
			}
			if tt.body != "{" {
				var apiErr *APIError
				if !errors.As(err, &apiErr) || apiErr.Code != 400 {
					t.Fatalf("wrapped APIError lost: %v", err)
				}
			}
		})
	}
}

func TestBatchMutationsCheckAPIResponse(t *testing.T) {
	operations := []struct {
		name   string
		method string
		run    func(*WhooingClient) ([]byte, error)
	}{
		{
			name: "update", method: http.MethodPut,
			run: func(c *WhooingClient) ([]byte, error) {
				return c.UpdateEntriesBatch("s1", []int64{42, 43}, EntryChanges{MemoSet: true})
			},
		},
		{
			name: "delete", method: http.MethodDelete,
			run: func(c *WhooingClient) ([]byte, error) { return c.DeleteEntries("s1", []int64{42, 43}) },
		},
	}
	responses := []struct {
		name    string
		status  int
		body    string
		wantErr bool
		apiCode int
	}{
		{name: "success preserves raw JSON", status: 200, body: `{"code":200,"rest_of_api":19,"results":{"extra":true}}`},
		{name: "no content envelope", status: 200, body: `{"code":204,"rest_of_api":19}`},
		{name: "API error", status: 200, body: `{"code":400,"message":"invalid"}`, wantErr: true, apiCode: 400},
		{name: "HTTP error", status: 400, body: `{}`, wantErr: true, apiCode: 400},
		{name: "malformed JSON", status: 200, body: `{`, wantErr: true},
	}
	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			for _, response := range responses {
				t.Run(response.name, func(t *testing.T) {
					calls := 0
					c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
						calls++
						if r.Method != op.method || r.URL.Path != "/api/entries/42,43/s1.json" {
							t.Fatalf("unexpected mutation: %s %s", r.Method, r.URL)
						}
						if op.method == http.MethodPut {
							if err := r.ParseForm(); err != nil {
								t.Fatal(err)
							}
							if !reflect.DeepEqual(r.PostForm, url.Values{"memo": {""}}) {
								t.Fatalf("empty memo not sent explicitly: %v", r.PostForm)
							}
						}
						return testResponse(response.status, response.body), nil
					})
					data, err := op.run(c)
					if calls != 1 || (err != nil) != response.wantErr {
						t.Fatalf("mutation = %s, %v; calls = %d", data, err, calls)
					}
					if response.apiCode != 0 {
						var apiErr *APIError
						if !errors.As(err, &apiErr) || apiErr.Code != response.apiCode {
							t.Fatalf("typed APIError lost: %v", err)
						}
					}
					if !response.wantErr && (string(data) != response.body || c.LastRestOfAPI() != 19) {
						t.Fatalf("raw response or quota lost: %s, rest = %d", data, c.LastRestOfAPI())
					}
				})
			}
		})
	}
}
