package main

import (
	"os"
	"path/filepath"
	"testing"
)

func restoreEnv(t *testing.T, key string) {
	prev, ok := os.LookupEnv(key)
	t.Cleanup(func() {
		if ok {
			_ = os.Setenv(key, prev)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

func TestLoadDotEnvFiles_FromCwd(t *testing.T) {
	restoreEnv(t, "KUAKE_LOAD_DOTENV")
	restoreEnv(t, "KUAKE_DOTENV_INTEGRATION_TEST")
	_ = os.Unsetenv("KUAKE_LOAD_DOTENV")
	_ = os.Unsetenv("KUAKE_DOTENV_INTEGRATION_TEST")

	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte("KUAKE_DOTENV_INTEGRATION_TEST=from_cwd\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	loadDotEnvFiles("config.json")
	if got := os.Getenv("KUAKE_DOTENV_INTEGRATION_TEST"); got != "from_cwd" {
		t.Fatalf("KUAKE_DOTENV_INTEGRATION_TEST: want from_cwd, got %q", got)
	}
}

func TestLoadDotEnvFiles_Disabled(t *testing.T) {
	restoreEnv(t, "KUAKE_LOAD_DOTENV")
	restoreEnv(t, "KUAKE_DOTENV_INTEGRATION_TEST")
	_ = os.Setenv("KUAKE_LOAD_DOTENV", "0")
	_ = os.Unsetenv("KUAKE_DOTENV_INTEGRATION_TEST")

	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte("KUAKE_DOTENV_INTEGRATION_TEST=should_not_apply\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	loadDotEnvFiles("config.json")
	if got := os.Getenv("KUAKE_DOTENV_INTEGRATION_TEST"); got != "" {
		t.Fatalf("KUAKE_DOTENV_INTEGRATION_TEST: want empty when disabled, got %q", got)
	}
}

func TestLoadDotEnvFiles_ConfigDirSecond(t *testing.T) {
	restoreEnv(t, "KUAKE_LOAD_DOTENV")
	restoreEnv(t, "KUAKE_DOTENV_INTEGRATION_TEST")
	_ = os.Unsetenv("KUAKE_LOAD_DOTENV")
	_ = os.Unsetenv("KUAKE_DOTENV_INTEGRATION_TEST")

	tmp := t.TempDir()
	sub := filepath.Join(tmp, "nested")
	if err := os.MkdirAll(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(sub, "config.json")
	if err := os.WriteFile(cfg, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, ".env"), []byte("KUAKE_DOTENV_INTEGRATION_TEST=from_config_dir\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	loadDotEnvFiles(cfg)
	if got := os.Getenv("KUAKE_DOTENV_INTEGRATION_TEST"); got != "from_config_dir" {
		t.Fatalf("KUAKE_DOTENV_INTEGRATION_TEST: want from_config_dir, got %q", got)
	}
}
