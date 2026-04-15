package sdk

import (
	"os"
	"testing"
)

func TestUploadParallelFromEnv(t *testing.T) {
	t.Parallel()
	restore := os.Getenv("KUAKE_UPLOAD_PARALLEL")
	t.Cleanup(func() {
		if restore == "" {
			_ = os.Unsetenv("KUAKE_UPLOAD_PARALLEL")
		} else {
			_ = os.Setenv("KUAKE_UPLOAD_PARALLEL", restore)
		}
	})

	cases := []struct {
		name   string
		env    string
		wantN  int
		wantOK bool
	}{
		{"unset", "", 0, false},
		{"spaces only", "   ", 0, false},
		{"valid 1", "1", 1, true},
		{"valid 16", "16", 16, true},
		{"invalid zero", "0", 0, false},
		{"invalid over max", "17", 0, false},
		{"invalid non int", "x", 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.env == "" {
				_ = os.Unsetenv("KUAKE_UPLOAD_PARALLEL")
			} else {
				_ = os.Setenv("KUAKE_UPLOAD_PARALLEL", tc.env)
			}
			gotN, gotOK := uploadParallelFromEnv()
			if gotN != tc.wantN || gotOK != tc.wantOK {
				t.Fatalf("uploadParallelFromEnv() = (%d, %v), want (%d, %v)", gotN, gotOK, tc.wantN, tc.wantOK)
			}
		})
	}
}
