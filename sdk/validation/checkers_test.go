package validation

import (
	"errors"
	"testing"
)

func TestNonEmpty(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"empty string", "", true},
		{"whitespace string", "   ", true},
		{"nil value", nil, true},
		{"non-empty string", "hello", false},
		{"non-empty with spaces", " hello ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NonEmpty()
			err := v.Validate(tt.value)
			if tt.wantErr && err == nil {
				t.Errorf("NonEmpty() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("NonEmpty() unexpected error: %v", err)
			}
		})
	}
}

func TestInRange(t *testing.T) {
	tests := []struct {
		name    string
		min     int
		max     int
		value   interface{}
		wantErr bool
	}{
		{"value in range", 1, 10, 5, false},
		{"value below range", 1, 10, 0, true},
		{"value above range", 1, 10, 11, true},
		{"non-int string", 1, 10, "hello", true},
		{"nil value", 1, 10, nil, true},
		{"int64 not accepted", 1, 10, int64(5), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := InRange(tt.min, tt.max)
			err := v.Validate(tt.value)
			if tt.wantErr && err == nil {
				t.Errorf("InRange(%d, %d) expected error, got nil", tt.min, tt.max)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("InRange(%d, %d) unexpected error: %v", tt.min, tt.max, err)
			}
		})
	}
}

func TestValidPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantErr  bool
		wantCode string // when wantErr; empty means do not assert code
	}{
		{"valid path", "/home/user/file.txt", false, ""},
		{"valid Windows path", "C:\\Users\\test", false, ""},
		{"valid path with at sign", "/home/user/release@team.txt", false, ""},
		{"filename with .. in name is not parent segment", "/photos/file..jpg", false, ""},
		{"path with parent reference", "/home/../etc/passwd", true, "FILE_PATH_TRAVERSAL"},
		{"path with double dots", "../../../secret", true, "FILE_PATH_TRAVERSAL"},
		{"single dot segment path", ".", true, "FILE_PATH_TRAVERSAL"},
		{"double dot segment path", "..", true, "FILE_PATH_TRAVERSAL"},
		{"NUL byte", "/foo\x00/bar", true, "FILE_INVALID_PATH"},
		{"UNC path", "\\\\server\\share", true, "FILE_INVALID_PATH"},
		{"UNC path with UNC prefix", "\\\\UNC\\server\\share", true, "FILE_INVALID_PATH"},
		{"forward slash UNC style", "//server/share/file", true, "FILE_INVALID_PATH"},
		{"Windows long path prefix", "\\\\?\\C:\\Windows", true, "FILE_INVALID_PATH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ValidPath()
			err := v.Validate(tt.path)
			if tt.wantErr && err == nil {
				t.Errorf("ValidPath() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidPath() unexpected error: %v", err)
			}
			if tt.wantErr && err != nil && tt.wantCode != "" {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Fatalf("ValidPath() want *ValidationError, got %T", err)
				}
				if ve.Code != tt.wantCode {
					t.Errorf("ValidPath() Code = %q, want %q", ve.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestInRangeFloat64_BelowMin(t *testing.T) {
	v := InRangeFloat64(1.0, 10.0)
	err := v.Validate(0.5)
	if err == nil {
		t.Fatal("expected error below min")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Code != "INVALID_ARG" {
		t.Fatalf("want INVALID_ARG, got %v", err)
	}
}

func TestInRangeFloat64_AboveMax(t *testing.T) {
	v := InRangeFloat64(1.0, 10.0)
	err := v.Validate(10.5)
	if err == nil {
		t.Fatal("expected error above max")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Code != "INVALID_ARG" {
		t.Fatalf("want INVALID_ARG, got %v", err)
	}
}

func TestInRangeFloat64_BoundaryMin(t *testing.T) {
	v := InRangeFloat64(1.0, 10.0)
	if err := v.Validate(1.0); err != nil {
		t.Fatalf("unexpected error at min boundary: %v", err)
	}
}

func TestInRangeFloat64_BoundaryMax(t *testing.T) {
	v := InRangeFloat64(1.0, 10.0)
	if err := v.Validate(10.0); err != nil {
		t.Fatalf("unexpected error at max boundary: %v", err)
	}
}

func TestInRangeFloat64_IntLikeJSON(t *testing.T) {
	v := InRangeFloat64(1.0, 10.0)
	if err := v.Validate(5); err != nil {
		t.Fatalf("int in range: %v", err)
	}
	if err := v.Validate(int64(7)); err != nil {
		t.Fatalf("int64 in range: %v", err)
	}
}

func TestInRangeFloat64_WrongType(t *testing.T) {
	v := InRangeFloat64(0.0, 1.0)
	err := v.Validate("not-a-number")
	if err == nil {
		t.Fatal("expected error for non-numeric value")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Code != "INVALID_ARG" {
		t.Fatalf("want INVALID_ARG, got %v", err)
	}
}

func TestValidEmail_Invalid(t *testing.T) {
	if err := ValidEmail().Validate("a@"); err == nil {
		t.Fatal("expected error for invalid email")
	}
}

func TestValidEmail_Valid(t *testing.T) {
	if err := ValidEmail().Validate("user@example.com"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidFID_Empty(t *testing.T) {
	if err := ValidFID().Validate(""); err == nil {
		t.Fatal("expected error for empty fid")
	}
}

func TestValidFID_InvalidChar(t *testing.T) {
	if err := ValidFID().Validate("bad@fid"); err == nil {
		t.Fatal("expected error for invalid character")
	}
}

func TestValidFID_Valid(t *testing.T) {
	if err := ValidFID().Validate("abc123_01"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidPath_ReturnsCleanPath(t *testing.T) {
	_, err := ValidPathResult("a/b/../c")
	if err == nil {
		t.Fatal("expected error for traversal segment")
	}
	got, err := ValidPathResult("foo//bar/")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got != "foo/bar" {
		t.Errorf("cleaned = %q, want foo/bar", got)
	}
}

func TestNonEmpty_IntRejected(t *testing.T) {
	if err := NonEmpty().Validate(42); err == nil {
		t.Fatal("expected type error")
	}
}

func TestInRange_Float64Input(t *testing.T) {
	if err := InRange(1, 10).Validate(float64(5)); err == nil {
		t.Fatal("expected error for float64")
	}
}

func TestChainWithValidators(t *testing.T) {
	intChain := Chain(InRange(1, 100))
	if err := intChain(50); err != nil {
		t.Errorf("Chain(InRange) unexpected error: %v", err)
	}
	if err := intChain(200); err == nil {
		t.Errorf("Chain(InRange) should error for out of range")
	}

	strChain := Chain(NonEmpty(), NewValidator("short", func(value interface{}) error {
		s, ok := value.(string)
		if !ok {
			return ErrInvalidArgument("need string")
		}
		if len(s) > 10 {
			return ErrInvalidArgument("too long")
		}
		return nil
	}))
	if err := strChain(""); err == nil {
		t.Errorf("Chain(NonEmpty,...) should error for empty string")
	}
	if err := strChain("ok"); err != nil {
		t.Errorf("Chain(NonEmpty,...) valid: %v", err)
	}
}
