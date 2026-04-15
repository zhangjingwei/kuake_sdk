package main

import (
	"testing"
)

func TestResolveUploadParallelForProcess(t *testing.T) {
	t.Run("flag_wins_over_env", func(t *testing.T) {
		t.Setenv("KUAKE_UPLOAD_PARALLEL", "2")
		if got := resolveUploadParallelForProcess("8"); got != "8" {
			t.Fatalf("got %q, want 8", got)
		}
	})
	t.Run("env_when_no_flag", func(t *testing.T) {
		t.Setenv("KUAKE_UPLOAD_PARALLEL", "3")
		if got := resolveUploadParallelForProcess(""); got != "3" {
			t.Fatalf("got %q, want 3", got)
		}
	})
	t.Run("invalid_env_ignored", func(t *testing.T) {
		t.Setenv("KUAKE_UPLOAD_PARALLEL", "99")
		if got := resolveUploadParallelForProcess(""); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})
}
