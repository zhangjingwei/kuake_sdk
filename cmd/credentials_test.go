package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoredCookieLifecycleAndPermissions(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "config")
	t.Setenv("KUAKE_CONFIG_DIR", configDir)
	cookie := "__pus=secret-one; __puus=secret-two; tfstk=secret-three"

	path, err := saveStoredCookie(cookie)
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(configDir, credentialsFileName) {
		t.Fatalf("unexpected path %q", path)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("credentials mode = %o, want 600", got)
	}
	dirInfo, err := os.Stat(configDir)
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("config directory mode = %o, want 700", got)
	}
	loaded, _, err := loadStoredCookie()
	if err != nil {
		t.Fatal(err)
	}
	if loaded != cookie {
		t.Fatal("stored Cookie did not round trip")
	}
	if _, err := clearStoredCookie(); err != nil {
		t.Fatal(err)
	}
	loaded, _, err = loadStoredCookie()
	if err != nil || loaded != "" {
		t.Fatalf("Cookie still present after clear: cookie=%q err=%v", loaded, err)
	}
}

func TestAuthStatusDoesNotExposeCookieValues(t *testing.T) {
	t.Setenv("KUAKE_CONFIG_DIR", t.TempDir())
	cookie := "__pus=do-not-print; __puus=also-secret"
	if _, err := saveStoredCookie(cookie); err != nil {
		t.Fatal(err)
	}
	result := handleAuth([]string{"status"}, "")
	encoded := result.Message
	for _, value := range result.Data {
		encoded += " " + fmt.Sprint(value)
	}
	if strings.Contains(encoded, "do-not-print") || strings.Contains(encoded, "also-secret") {
		t.Fatal("auth status exposed a Cookie value")
	}
}
