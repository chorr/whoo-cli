package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestConfigComputeSigniture(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("WHOOING_APP_SECRET", "app-secret")
	cfg := &Config{TokenSecret: "token-secret"}
	// SHA-1("app-secret|token-secret")의 고정 벡터로 구분자와 인코딩까지 검증한다.
	if got := cfg.ComputeSigniture(); got != "ad4794b182b86d80d63c565878208e973ea9eb0e" {
		t.Fatalf("ComputeSigniture() = %q", got)
	}
	cfg.TokenSecret = ""
	if got := cfg.ComputeSigniture(); got != "" {
		t.Fatalf("missing token secret: got %q", got)
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		content string
		write   bool
		wantErr bool
	}{
		{name: "missing file"},
		{name: "empty object", write: true, content: `{}`},
		{name: "malformed JSON", write: true, content: `{`, wantErr: true},
		{name: "wrong field type", write: true, content: `{"token":123}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			if tt.write {
				path, err := getConfigPath()
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(tt.content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			cfg, err := Load()
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "설정 파일 파싱 실패") {
					t.Fatalf("Load() = %+v, %v; want parse error", cfg, err)
				}
				return
			}
			if err != nil || !reflect.DeepEqual(cfg, &Config{}) {
				t.Fatalf("Load() = %+v, %v; want empty config", cfg, err)
			}
		})
	}
}

func TestConfigSaveRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &Config{
		Token: "test-token", TokenSecret: "test-token-secret", SectionID: "s1",
		PendingRequestToken: "pending", PendingSigniture: "legacy", Verbose: true,
	}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := *cfg
	want.Verbose = false
	if !reflect.DeepEqual(loaded, &want) {
		t.Fatalf("Load() = %+v, want %+v", loaded, want)
	}
	path := filepath.Join(os.Getenv("HOME"), ".config", "whoo", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "verbose") || strings.Contains(string(data), "app_secret") {
		t.Fatalf("runtime-only fields persisted: %s", data)
	}
}

func TestConfigSaveRestrictsPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file permissions")
	}
	tests := []struct {
		name     string
		existing bool
	}{
		{name: "new file"},
		{name: "existing readable file", existing: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			path, err := getConfigPath()
			if err != nil {
				t.Fatal(err)
			}
			if tt.existing {
				if err := os.WriteFile(path, []byte(`{}`), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(path, 0644); err != nil {
					t.Fatal(err)
				}
			}
			if err := (&Config{Token: "new-test-token"}).Save(); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if got := info.Mode().Perm(); got != 0600 {
				t.Fatalf("config permissions = %04o, want 0600", got)
			}
		})
	}
}

func TestConfigIsAuthenticated(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{name: "empty"},
		{name: "token only", cfg: Config{Token: "t"}},
		{name: "secret only", cfg: Config{TokenSecret: "s"}},
		{name: "pending only", cfg: Config{PendingRequestToken: "p"}},
		{name: "complete", cfg: Config{Token: "t", TokenSecret: "s"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.cfg.IsAuthenticated(); got != tt.want {
				t.Fatalf("IsAuthenticated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadWithTestTokens(t *testing.T) {
	tests := []struct {
		name   string
		saved  Config
		token  string
		secret string
		want   Config
	}{
		{name: "both overrides", token: "test", secret: "secret", want: Config{Token: "test", TokenSecret: "secret"}},
		{name: "partial override ignored", token: "test"},
		{
			name: "saved authentication wins", token: "test", secret: "secret",
			saved: Config{Token: "saved", TokenSecret: "saved-secret", SectionID: "s1"},
			want:  Config{Token: "saved", TokenSecret: "saved-secret", SectionID: "s1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("WHOOING_TEST_TOKEN", tt.token)
			t.Setenv("WHOOING_TEST_TOKEN_SECRET", tt.secret)
			if err := tt.saved.Save(); err != nil {
				t.Fatal(err)
			}
			got, err := LoadWithTestTokens()
			if err != nil || !reflect.DeepEqual(got, &tt.want) {
				t.Fatalf("LoadWithTestTokens() = %+v, %v; want %+v", got, err, tt.want)
			}
			persisted, err := Load()
			if err != nil || !reflect.DeepEqual(persisted, &tt.saved) {
				t.Fatalf("test token override changed disk config: %+v, %v", persisted, err)
			}
		})
	}
}
