//go:build integration

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"whoo-cli/config"
)

type entriesCLIRequest struct {
	Method string
	Path   string
	Query  url.Values
	Form   url.Values
}

type entriesCLIResponse struct {
	Status int
	Body   string
}

type entriesCLIScenario struct {
	Args      []string
	Responses []entriesCLIResponse
	Capture   string
}

type entriesCLITransport func(*http.Request) (*http.Response, error)

func (f entriesCLITransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// TestEntriesCLIProcess는 os.Exit을 사용하는 CLI 핸들러의 격리 실행 진입점이다.
// 전역 transport 교체는 이 자식 프로세스 안에서만 수행한다.
func TestEntriesCLIProcess(t *testing.T) {
	path := os.Getenv("WHOO_CLI_TEST_SCENARIO")
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var scenario entriesCLIScenario
	if err := json.Unmarshal(data, &scenario); err != nil {
		t.Fatal(err)
	}
	requests := []entriesCLIRequest{}
	http.DefaultTransport = entriesCLITransport(func(r *http.Request) (*http.Response, error) {
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		request := entriesCLIRequest{Method: r.Method, Path: r.URL.Path}
		if len(r.URL.Query()) > 0 {
			request.Query = r.URL.Query()
		}
		if len(r.PostForm) > 0 {
			request.Form = r.PostForm
		}
		requests = append(requests, request)
		captured, err := json.Marshal(requests)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(scenario.Capture, captured, 0600); err != nil {
			return nil, err
		}
		index := len(requests) - 1
		if index >= len(scenario.Responses) {
			return nil, fmt.Errorf("예상하지 않은 API 요청: %s %s", r.Method, r.URL.Path)
		}
		response := scenario.Responses[index]
		return &http.Response{
			StatusCode: response.Status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(response.Body)),
		}, nil
	})
	cfg := &config.Config{Token: "test-token", TokenSecret: "test-secret", SectionID: "saved"}
	args, options, err := ParseGlobalOptions(cfg, append([]string{"entries"}, scenario.Args...))
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	WarnSectionOverride(options, args)
	RunEntries(cfg, args[1:])
	os.Exit(0)
}

func TestEntriesCLI(t *testing.T) {
	t.Parallel()
	entry := `{"entry_id":42,"entry_date":"20261029.0001","money":1200,"item":"커피"}`
	created := entriesCLIResponse{Status: 200, Body: `{"code":200,"results":[` + entry + `]}`}
	updated := entriesCLIResponse{Status: 200, Body: `{"code":200,"results":` + entry + `}`}
	ok := entriesCLIResponse{Status: 200, Body: `{"code":200,"results":{}}`}
	apiFailure := entriesCLIResponse{Status: 200, Body: `{"code":400,"message":"invalid entry"}`}
	httpFailure := entriesCLIResponse{Status: 400, Body: `{"message":"invalid entry"}`}
	addArgs := []string{
		"add", "--l-account", "expenses", "--l-id", "e1", "--r-account", "assets", "--r-id", "a1",
		"--money", "1200", "--date", "20261029", "--item", "커피",
	}
	addForm := url.Values{
		"section_id": {"saved"}, "entry_date": {"20261029"}, "l_account": {"expenses"}, "l_account_id": {"e1"},
		"r_account": {"assets"}, "r_account_id": {"a1"}, "money": {"1200"}, "item": {"커피"},
	}
	addRequest := entriesCLIRequest{Method: "POST", Path: "/api/entries.json", Form: addForm}
	withItem := func(item string) entriesCLIRequest {
		request := addRequest
		request.Form = maps.Clone(addForm)
		request.Form.Set("item", item)
		return request
	}
	searchArgs := []string{"search", "--from", "20261001", "--to", "20261029"}
	searchQuery := url.Values{
		"section_id": {"saved"}, "start_date": {"20261001"}, "end_date": {"20261029"},
		"limit": {"20"}, "sort_order": {"desc"},
	}
	accountsResponse := entriesCLIResponse{Status: 200, Body: `{"code":200,"results":{
		"expenses":{"e1":{"title":"식비"}},"assets":{"a1":{"title":"현금"}}
	}}`}
	accountsRequest := entriesCLIRequest{
		Method: "GET", Path: "/api/accounts.json", Query: url.Values{"section_id": {"saved"}},
	}
	resolvedQuery := maps.Clone(searchQuery)
	resolvedQuery.Set("l_account", "expenses")
	resolvedQuery.Set("l_account_id", "e1")
	resolvedQuery.Set("r_account", "assets")
	resolvedQuery.Set("r_account_id", "a1")
	filteredQuery := maps.Clone(searchQuery)
	for key, value := range map[string]string{
		"account": "expenses", "account_id": "e1", "max": "20261029.0017",
		"item": "커피*", "memo": "점심 !제외", "money_from": "100", "money_to": "2000",
		"sort_column": "money", "sort_order": "asc", "limit": "5",
	} {
		filteredQuery.Set(key, value)
	}
	batchJSON := `[{"entry_date":20261029,"l_account":"expenses","l_account_id":"e1",` +
		`"r_account":"assets","r_account_id":"a1","item":"커피","money":1200,"memo":"점심"}]`
	batchRequest := entriesCLIRequest{Method: "POST", Path: "/api/entries.json", Form: url.Values{
		"section_id": {"saved"}, "data_type": {"json"}, "entries": {batchJSON},
	}}
	tests := []struct {
		name      string
		args      []string
		responses []entriesCLIResponse
		requests  []entriesCLIRequest
		wantErr   string
		wantJSON  string
		fileJSON  string
	}{
		{
			name: "add", args: addArgs, responses: []entriesCLIResponse{created},
			requests: []entriesCLIRequest{addRequest}, wantJSON: `{"entry_id":42,"section_id":"saved"}`,
		},
		{
			name: "repeat command", args: append(append([]string{}, addArgs...), "--repeat", "12"),
			responses: []entriesCLIResponse{created}, requests: []entriesCLIRequest{withItem("커피**12")},
			wantJSON: `{"entry_id":42}`,
		},
		{
			name: "split command", args: append(append([]string{}, addArgs...), "--split", "3", "--fee", "15.5"),
			responses: []entriesCLIResponse{created}, requests: []entriesCLIRequest{withItem("커피//3(15.50%)")},
			wantJSON: `{"entry_id":42}`,
		},
		{
			name: "add API failure", args: addArgs, responses: []entriesCLIResponse{apiFailure},
			requests: []entriesCLIRequest{addRequest}, wantErr: "invalid entry",
		},
		{
			name: "add HTTP failure", args: addArgs, responses: []entriesCLIResponse{httpFailure},
			requests: []entriesCLIRequest{addRequest}, wantErr: "invalid entry",
		},
		{name: "add required fields", args: []string{"add", "--item", "커피"}, wantErr: "필수"},
		{
			name: "add invalid item command", args: append(append([]string{}, addArgs...), "--split", "1"),
			wantErr: "아이템 명령어 오류",
		},
		{
			name: "batch file", args: []string{"batch", "--file", "rows.json"}, fileJSON: batchJSON,
			responses: []entriesCLIResponse{ok}, wantJSON: `{"code":200}`,
			requests: []entriesCLIRequest{batchRequest},
		},
		{
			name: "batch file API failure", args: []string{"batch", "--file", "rows.json"}, fileJSON: batchJSON,
			responses: []entriesCLIResponse{apiFailure}, wantErr: "invalid entry",
			requests: []entriesCLIRequest{batchRequest},
		},
		{
			name: "batch malformed file", args: []string{"batch", "--file", "rows.json"},
			fileJSON: `{`, wantErr: "JSON 파싱 실패",
		},
		{
			name: "batch missing file", args: []string{"batch", "--file", "missing.json"}, wantErr: "파일 읽기 실패",
		},
		{
			name: "update changed field only", args: []string{"update", "42", "--memo", "점심 & 커피"},
			responses: []entriesCLIResponse{updated}, wantJSON: `{"entry_id":42,"section_id":"saved"}`,
			requests: []entriesCLIRequest{{Method: "PUT", Path: "/api/entries/42.json", Form: url.Values{
				"section_id": {"saved"}, "memo": {"점심 & 커피"},
			}}},
		},
		{
			name: "update clears memo", args: []string{"update", "42", "--memo", ""},
			responses: []entriesCLIResponse{updated}, wantJSON: `{"entry_id":42}`,
			requests: []entriesCLIRequest{{Method: "PUT", Path: "/api/entries/42.json", Form: url.Values{
				"section_id": {"saved"}, "memo": {""},
			}}},
		},
		{name: "update invalid ID", args: []string{"update", "bad", "--money", "100"}, wantErr: "entry_id"},
		{
			name: "update API failure", args: []string{"update", "42", "--money", "100"},
			responses: []entriesCLIResponse{apiFailure}, wantErr: "invalid entry",
			requests: []entriesCLIRequest{{Method: "PUT", Path: "/api/entries/42.json", Form: url.Values{
				"section_id": {"saved"}, "money": {"100"},
			}}},
		},
		{
			name: "batch update", args: []string{"update", "42,43", "--money", "100"},
			responses: []entriesCLIResponse{ok}, wantJSON: `{"code":200}`,
			requests: []entriesCLIRequest{{
				Method: "PUT", Path: "/api/entries/42,43/saved.json", Form: url.Values{"money": {"100"}},
			}},
		},
		{
			name: "batch update clears memo", args: []string{"update", "42,43", "--memo", ""},
			responses: []entriesCLIResponse{ok}, wantJSON: `{"code":200}`,
			requests: []entriesCLIRequest{{
				Method: "PUT", Path: "/api/entries/42,43/saved.json", Form: url.Values{"memo": {""}},
			}},
		},
		{
			name: "batch update API failure", args: []string{"update", "42,43", "--money", "100"},
			responses: []entriesCLIResponse{apiFailure}, wantErr: "invalid entry",
			requests: []entriesCLIRequest{{
				Method: "PUT", Path: "/api/entries/42,43/saved.json", Form: url.Values{"money": {"100"}},
			}},
		},
		{
			name: "delete comma IDs", args: []string{"delete", "42,43"},
			responses: []entriesCLIResponse{ok}, wantJSON: `{"code":200,"section_id":"saved"}`,
			requests: []entriesCLIRequest{{Method: "DELETE", Path: "/api/entries/42,43/saved.json"}},
		},
		{
			name: "delete separate IDs", args: []string{"rm", "42", "43"},
			responses: []entriesCLIResponse{ok}, wantJSON: `{"code":200}`,
			requests: []entriesCLIRequest{{Method: "DELETE", Path: "/api/entries/42,43/saved.json"}},
		},
		{
			name: "delete API failure", args: []string{"delete", "42"},
			responses: []entriesCLIResponse{apiFailure}, wantErr: "invalid entry",
			requests: []entriesCLIRequest{{Method: "DELETE", Path: "/api/entries/42/saved.json"}},
		},
		{name: "delete no IDs", args: []string{"delete"}, wantErr: "entry_id"},
		{name: "delete invalid ID", args: []string{"del", "42,bad"}, wantErr: "entry_id"},
		{
			name:      "search resolves both account types",
			args:      append(append([]string{}, searchArgs...), "--l-id", "e1", "--r-id", "a1"),
			responses: []entriesCLIResponse{accountsResponse, ok}, wantJSON: `{"code":200}`,
			requests: []entriesCLIRequest{
				accountsRequest, {Method: "GET", Path: "/api/entries.json", Query: resolvedQuery},
			},
		},
		{
			name: "search filters and exact cursor",
			args: append(append([]string{}, searchArgs...),
				"--account", "expenses", "--account-id", "e1", "--max", "20261029.0017",
				"--item", "커피*", "--memo", "점심 !제외", "--money-from", "100", "--money-to", "2000",
				"--sort", "money", "--order", "asc", "--limit", "5",
			),
			responses: []entriesCLIResponse{ok}, wantJSON: `{"code":200}`,
			requests: []entriesCLIRequest{{Method: "GET", Path: "/api/entries.json", Query: filteredQuery}},
		},
		{
			name:      "search unknown account stops before entries",
			args:      append(append([]string{}, searchArgs...), "--account-id", "unknown"),
			responses: []entriesCLIResponse{accountsResponse}, wantErr: "찾을 수 없습니다",
			requests: []entriesCLIRequest{accountsRequest},
		},
		{
			name:      "search accounts failure stops before entries",
			args:      append(append([]string{}, searchArgs...), "--account-id", "e1"),
			responses: []entriesCLIResponse{httpFailure}, wantErr: "항목 타입 확인 실패",
			requests: []entriesCLIRequest{accountsRequest},
		},
		{
			name: "section override", args: []string{"delete", "42", "--section", "other", "--allow-section-override"},
			responses: []entriesCLIResponse{ok}, wantJSON: `{"section_id":"other"}`,
			requests: []entriesCLIRequest{{Method: "DELETE", Path: "/api/entries/42/other.json"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			home := t.TempDir()
			if tt.fileJSON != "" {
				if err := os.WriteFile(filepath.Join(home, "rows.json"), []byte(tt.fileJSON), 0600); err != nil {
					t.Fatal(err)
				}
			}
			scenario := entriesCLIScenario{
				Args: tt.args, Responses: tt.responses, Capture: filepath.Join(home, "requests.json"),
			}
			data, err := json.Marshal(scenario)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(home, "scenario.json")
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			child := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestEntriesCLIProcess$")
			child.Dir = home
			child.Env = []string{}
			for _, variable := range os.Environ() {
				key, _, _ := strings.Cut(variable, "=")
				isWhoo := strings.HasPrefix(key, "WHOO_") || strings.HasPrefix(key, "WHOOING_")
				if key != "HOME" && !isWhoo {
					child.Env = append(child.Env, variable)
				}
			}
			child.Env = append(child.Env,
				"HOME="+home, "WHOO_CLI_TEST_SCENARIO="+path,
				"WHOOING_APP_ID=test-app", "WHOOING_APP_SECRET=test-app-secret",
			)
			var stdout, stderr bytes.Buffer
			child.Stdout, child.Stderr = &stdout, &stderr
			err = child.Run()
			if ctx.Err() != nil {
				t.Fatalf("CLI timed out: %v", ctx.Err())
			}
			if tt.wantErr != "" {
				if err == nil || stdout.Len() != 0 || !strings.Contains(stderr.String(), tt.wantErr) {
					t.Fatalf("expected error %q: err = %v, stdout = %s, stderr = %s", tt.wantErr, err, &stdout, &stderr)
				}
			} else {
				if err != nil || stderr.Len() != 0 {
					t.Fatalf("CLI failed: %v, stderr = %s", err, &stderr)
				}
				var got, want map[string]interface{}
				if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
					t.Fatalf("stdout is not JSON: %s: %v", &stdout, err)
				}
				if err := json.Unmarshal([]byte(tt.wantJSON), &want); err != nil {
					t.Fatal(err)
				}
				for key, value := range want {
					if !reflect.DeepEqual(got[key], value) {
						t.Errorf("stdout[%q] = %v, want %v", key, got[key], value)
					}
				}
			}
			requests := []entriesCLIRequest{}
			captured, err := os.ReadFile(scenario.Capture)
			if err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			if len(captured) > 0 {
				if err := json.Unmarshal(captured, &requests); err != nil {
					t.Fatal(err)
				}
			}
			if len(requests) != len(tt.requests) || (len(requests) > 0 && !reflect.DeepEqual(requests, tt.requests)) {
				t.Fatalf("requests = %+v, want %+v", requests, tt.requests)
			}
		})
	}
}
