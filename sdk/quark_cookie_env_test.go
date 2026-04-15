package sdk

import (
	"os"
	"testing"
)

func TestNormalizeQuarkCookieInput(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"whitespace_only", "   \t", ""},
		{"add_prefix_and_semicolon", "abc", "__pus=abc;"},
		{"preserve_pus_add_semi", "__pus=val", "__pus=val;"},
		{"trim_then_normalize", "  token  ", "__pus=token;"},
		{"already_complete", "__pus=x;y=1;", "__pus=x;y=1;"},
		{"puus_only_no_extra_pus_prefix", "__puus=val", "__puus=val;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeQuarkCookieInput(tt.in); got != tt.want {
				t.Fatalf("NormalizeQuarkCookieInput(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

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

func TestCookieFromSplitEnv(t *testing.T) {
	restoreEnv(t, "KUAKE_PUS")
	restoreEnv(t, "KUAKE_PUUS")
	restoreEnv(t, "KUAKE_COOKIE")

	_ = os.Unsetenv("KUAKE_COOKIE")
	_ = os.Unsetenv("KUAKE_PUS")
	_ = os.Unsetenv("KUAKE_PUUS")
	if got := cookieFromSplitEnv(); got != "" {
		t.Fatalf("empty env: got %q", got)
	}

	_ = os.Setenv("KUAKE_PUS", "p1")
	_ = os.Unsetenv("KUAKE_PUUS")
	if got := NormalizeQuarkCookieInput(cookieFromSplitEnv()); got != "__pus=p1;" {
		t.Fatalf("pus only: got %q", got)
	}

	_ = os.Unsetenv("KUAKE_PUS")
	_ = os.Setenv("KUAKE_PUUS", "u1")
	if got := NormalizeQuarkCookieInput(cookieFromSplitEnv()); got != "__puus=u1;" {
		t.Fatalf("puus only: got %q", got)
	}

	_ = os.Setenv("KUAKE_PUS", "p2")
	_ = os.Setenv("KUAKE_PUUS", "u2")
	if got := NormalizeQuarkCookieInput(cookieFromSplitEnv()); got != "__pus=p2; __puus=u2;" {
		t.Fatalf("both: got %q", got)
	}
}

func TestResolveEnvCookieString_KUAKECookieWins(t *testing.T) {
	for _, k := range []string{"KUAKE_COOKIE", "KUAKE_PUS", "KUAKE_PUUS"} {
		restoreEnv(t, k)
	}
	_ = os.Setenv("KUAKE_COOKIE", "__pus=whole;")
	_ = os.Setenv("KUAKE_PUS", "ignored")
	_ = os.Setenv("KUAKE_PUUS", "ignored")
	if got := ResolveEnvCookieString(); got != "__pus=whole;" {
		t.Fatalf("want whole cookie, got %q", got)
	}
}
