package validation

import "testing"

func TestParseIntArg(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		argName string
		want    int
		wantErr string
	}{
		{"valid", "42", "days", 42, ""},
		{"empty", "", "days", 0, "days must be int (例如：'1'); got ''"},
		{"invalid", "abc", "days", 0, "days must be int (例如：'1'); got 'abc'"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIntArg(tt.value, tt.argName)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("error %q, want %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseOptionalIntArg(t *testing.T) {
	cases := []struct {
		name     string
		value    string
		argName  string
		defaultV int
		want     int
		wantErr  string
	}{
		{"empty returns default", "", "page", 5, 5, ""},
		{"valid", "10", "page", 5, 10, ""},
		{"invalid", "abc", "page", 5, 0, "page must be int (例如：'1'); got 'abc'"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOptionalIntArg(tt.value, tt.argName, tt.defaultV)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("error %q, want %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseBoolArg(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		argName string
		want    bool
		wantErr string
	}{
		{"true", "true", "passcode", true, ""},
		{"false", "false", "passcode", false, ""},
		{"invalid", "yes", "passcode", false, "passcode must be bool (例如：'true', 'false'); got 'yes'"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBoolArg(tt.value, tt.argName)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("error %q, want %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
