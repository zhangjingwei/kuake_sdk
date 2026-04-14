package validation

import "testing"

func TestDefaults_Apply(t *testing.T) {
	defaults := NewDefaults(map[string]interface{}{
		"page":  1,
		"size":  50,
		"order": "desc",
	})

	tests := []struct {
		name       string
		params     map[string]interface{}
		wantParams map[string]interface{}
	}{
		{
			name:       "empty params",
			params:     map[string]interface{}{},
			wantParams: map[string]interface{}{"page": 1, "size": 50, "order": "desc"},
		},
		{
			name:       "partial params",
			params:     map[string]interface{}{"page": 2},
			wantParams: map[string]interface{}{"page": 2, "size": 50, "order": "desc"},
		},
		{
			name:       "all params",
			params:     map[string]interface{}{"page": 2, "size": 100},
			wantParams: map[string]interface{}{"page": 2, "size": 100, "order": "desc"},
		},
		{
			name:       "zero values treated as missing",
			params:     map[string]interface{}{"page": 0, "size": 0, "order": ""},
			wantParams: map[string]interface{}{"page": 1, "size": 50, "order": "desc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := make(map[string]interface{})
			for k, v := range tt.params {
				params[k] = v
			}
			defaults.Apply(params)
			for k, v := range tt.wantParams {
				if params[k] != v {
					t.Errorf("Defaults.Apply() %s = %v, want %v", k, params[k], v)
				}
			}
			// Check no extra keys
			if len(params) != len(tt.wantParams) {
				t.Errorf("Defaults.Apply() extra keys: %v", params)
			}
		})
	}
}

func TestPageDefaults(t *testing.T) {
	params := map[string]interface{}{
		"page":  2,
		"size":  100,
	}
	PageDefaults.Apply(params)

	if params["page"] != 2 {
		t.Errorf("PageDefaults preserved page = %v, want 2", params["page"])
	}
	if params["size"] != 100 {
		t.Errorf("PageDefaults preserved size = %v, want 100", params["size"])
	}

	// New params should get defaults
	params2 := map[string]interface{}{}
	PageDefaults.Apply(params2)
	if params2["page"] != 1 {
		t.Errorf("PageDefaults page default = %v, want 1", params2["page"])
	}
	if params2["size"] != 50 {
		t.Errorf("PageDefaults size default = %v, want 50", params2["size"])
	}
}

func TestDefaults_CRUDAndMerge(t *testing.T) {
	defaults := NewDefaults(map[string]interface{}{
		"a": 1,
		"b": 2,
	})

	defaults.Add("b", 20)
	defaults.Add("c", 3)

	if val, ok := defaults.Get("b"); !ok || val != 2 {
		t.Fatalf("Add should not overwrite existing key b, got %v", val)
	}
	if val, ok := defaults.Get("c"); !ok || val != 3 {
		t.Fatalf("Add should set new key c, got %v", val)
	}

	defaults.Set("b", 20)
	if val, ok := defaults.Get("b"); !ok || val != 20 {
		t.Fatalf("Set should overwrite key b, got %v", val)
	}

	other := NewDefaults(map[string]interface{}{
		"b": 200,
		"d": 4,
	})
	defaults.Merge(other)
	if val, ok := defaults.Get("b"); !ok || val != 20 {
		t.Fatalf("Merge should preserve existing key b, got %v", val)
	}
	if val, ok := defaults.Get("d"); !ok || val != 4 {
		t.Fatalf("Merge should add new key d, got %v", val)
	}
}

func TestUploadOptionsDefaults(t *testing.T) {
	val, ok := UploadOptionsDefaults.Get("policy")
	if !ok {
		t.Fatal("UploadOptionsDefaults should contain policy")
	}
	if val != "skip" {
		t.Fatalf("UploadOptionsDefaults policy = %v, want skip", val)
	}

	params := map[string]interface{}{"policy": ""}
	UploadOptionsDefaults.Apply(params)
	if params["policy"] != "skip" {
		t.Fatalf("UploadOptionsDefaults.Apply() policy = %v, want skip", params["policy"])
	}
}
