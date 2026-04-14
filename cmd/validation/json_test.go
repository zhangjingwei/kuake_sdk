package validation

import "testing"

func TestSafeMap(t *testing.T) {
	cases := []struct {
		name      string
		data      map[string]interface{}
		key       string
		wantOK    bool
		wantValue map[string]interface{}
	}{
		{"valid map", map[string]interface{}{"data": map[string]interface{}{"path": "test"}}, "data", true, map[string]interface{}{"path": "test"}},
		{"non-map", map[string]interface{}{"data": "string"}, "data", false, nil},
		{"missing", map[string]interface{}{}, "data", false, nil},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := SafeMap(tt.data, tt.key)
			if ok != tt.wantOK {
				t.Fatalf("got ok %v, want %v", ok, tt.wantOK)
			}
			if ok {
				if got["path"] != tt.wantValue["path"] {
					t.Fatalf("got %v, want %v", got, tt.wantValue)
				}
			}
		})
	}
}

func TestExtractPathFromJSON(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		wantPath  string
		wantFid   string
		wantError bool
	}{
		{"full format", `{"success":true,"data":{"path":"/tmp/file.txt","fid":"abc"}}`, "/tmp/file.txt", "abc", false},
		{"simple format", `{"path":"/tmp/file.txt","fid":"abc"}` , "/tmp/file.txt", "abc", false},
		{"invalid json", `not json`, "", "", true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			path, fid, err := ExtractPathFromJSON(tt.input)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if path != tt.wantPath || fid != tt.wantFid {
				t.Fatalf("got path=%q fid=%q, want path=%q fid=%q", path, fid, tt.wantPath, tt.wantFid)
			}
		})
	}
}
