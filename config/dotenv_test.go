package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyDotEnvFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "# comment\n" +
		"WHOOING_APP_ID=from-file\n" +
		"export WHOOING_APP_SECRET=\"secret-value\"\n" +
		"ALREADY_SET=file-value\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ALREADY_SET", "shell-value")
	// 파일에서 채운 값과 기존 환경변수를 테스트 종료 시 함께 복원한다.
	t.Setenv("WHOOING_APP_ID", "")
	t.Setenv("WHOOING_APP_SECRET", "")

	if err := applyDotEnvFile(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("WHOOING_APP_ID"); got != "from-file" {
		t.Fatalf("APP_ID=%q", got)
	}
	if got := os.Getenv("WHOOING_APP_SECRET"); got != "secret-value" {
		t.Fatalf("APP_SECRET=%q", got)
	}
	// 기존 환경변수는 덮어쓰지 않음
	if got := os.Getenv("ALREADY_SET"); got != "shell-value" {
		t.Fatalf("ALREADY_SET overwritten: %q", got)
	}
}
