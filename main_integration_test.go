//go:build integration

package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCLI(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "whoo")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}

	commands := []string{
		"", "auth", "login", "user", "user_logs", "user_point_logs", "sections", "s", "accounts", "a",
		"entries", "e", "frequent", "freq", "f", "monthly", "month", "m", "inout", "io",
		"bs", "balance", "report", "r", "budget", "budget-goal", "goal", "bill", "b", "checkcard", "cc",
		"entries add", "entries batch", "entries flow account", "accounts edit", "sections delete",
		"frequent use", "monthly pay", "budget set", "budget-goal set", "goal set", "bill add", "checkcard edit",
	}
	for _, command := range commands {
		for _, flag := range []string{"help", "--help", "-h"} {
			args := append(strings.Fields(command), flag)
			t.Run(strings.Join(args, " "), func(t *testing.T) {
				stdout, stderr, err := runCLI(t, binary, args)
				if err != nil || stderr != "" || !strings.Contains(stdout, "whoo") {
					t.Fatalf("help failed: err = %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
				}
			})
		}
	}
	t.Run("unauthenticated command", func(t *testing.T) {
		stdout, stderr, err := runCLI(t, binary, []string{"entries"})
		if err == nil || stdout != "" || !strings.Contains(stderr, "인증이 필요합니다") {
			t.Fatalf("auth guard: err = %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}
	})
}

func runCLI(t *testing.T, binary string, args []string) (string, string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c := exec.CommandContext(ctx, binary, args...)
	home := t.TempDir()
	c.Dir = home
	c.Env = []string{}
	for _, variable := range os.Environ() {
		key, _, _ := strings.Cut(variable, "=")
		isWhooVariable := strings.HasPrefix(key, "WHOOING_") || strings.HasPrefix(key, "WHOO_")
		if key != "HOME" && !isWhooVariable {
			c.Env = append(c.Env, variable)
		}
	}
	c.Env = append(c.Env, "HOME="+home)
	var stdout, stderr bytes.Buffer
	c.Stdout, c.Stderr = &stdout, &stderr
	err := c.Run()
	if ctx.Err() != nil {
		t.Fatalf("CLI did not exit: %v", ctx.Err())
	}
	configPath := filepath.Join(home, ".config", "whoo", "config.json")
	if _, statErr := os.Stat(configPath); !os.IsNotExist(statErr) {
		t.Fatalf("read-only CLI wrote configuration: %v", statErr)
	}
	return stdout.String(), stderr.String(), err
}
