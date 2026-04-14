package validation

import "testing"

func TestSafeString(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		wantVal  string
		wantOK   bool
	}{
		{"valid string", map[string]interface{}{"name": "test"}, "name", "test", true},
		{"non-string value", map[string]interface{}{"num": 123}, "num", "", false},
		{"missing key", map[string]interface{}{"name": "test"}, "missing", "", false},
		{"nil map", nil, "key", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOK := SafeString(tt.data, tt.key)
			if gotVal != tt.wantVal || gotOK != tt.wantOK {
				t.Errorf("SafeString() = (%v, %v), want (%v, %v)", gotVal, gotOK, tt.wantVal, tt.wantOK)
			}
		})
	}
}

func TestSafeFloat64(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		wantVal  float64
		wantOK   bool
	}{
		{"valid float", map[string]interface{}{"num": 123.45}, "num", 123.45, true},
		{"integer from float64", map[string]interface{}{"num": 100}, "num", 100, true},
		{"non-numeric", map[string]interface{}{"num": "100"}, "num", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOK := SafeFloat64(tt.data, tt.key)
			if gotVal != tt.wantVal || gotOK != tt.wantOK {
				t.Errorf("SafeFloat64() = (%v, %v), want (%v, %v)", gotVal, gotOK, tt.wantVal, tt.wantOK)
			}
		})
	}
}

func TestSafeInt(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		wantVal  int
		wantOK   bool
	}{
		{"valid int", map[string]interface{}{"num": 100}, "num", 100, true},
		{"float64 from JSON", map[string]interface{}{"num": 123.0}, "num", 123, true},
		{"non-numeric", map[string]interface{}{"num": "100"}, "num", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOK := SafeInt(tt.data, tt.key)
			if gotVal != tt.wantVal || gotOK != tt.wantOK {
				t.Errorf("SafeInt() = (%v, %v), want (%v, %v)", gotVal, gotOK, tt.wantVal, tt.wantOK)
			}
		})
	}
}

func TestSafeMap(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]interface{}
		key     string
		wantVal map[string]interface{}
		wantOK  bool
	}{
		{"valid map", map[string]interface{}{"obj": map[string]interface{}{"path": "file.txt"}}, "obj", map[string]interface{}{"path": "file.txt"}, true},
		{"non-map value", map[string]interface{}{"obj": "notamap"}, "obj", nil, false},
		{"missing key", map[string]interface{}{"obj": map[string]interface{}{"path": "file.txt"}}, "missing", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOK := SafeMap(tt.data, tt.key)
			if gotOK != tt.wantOK {
				t.Errorf("SafeMap() = (%v, %v), wantOK %v", gotVal, gotOK, tt.wantOK)
				return
			}
			if gotOK {
				if len(gotVal) != len(tt.wantVal) {
					t.Errorf("SafeMap() length = %d, want %d", len(gotVal), len(tt.wantVal))
				}
				for k, v := range tt.wantVal {
					if gotVal[k] != v {
						t.Errorf("SafeMap()[%s] = %v, want %v", k, gotVal[k], v)
					}
				}
			}
		})
	}
}
