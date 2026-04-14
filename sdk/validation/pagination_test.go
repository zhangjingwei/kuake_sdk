package validation

import "testing"

func TestPaginateParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		page    int
		size    int
		wantErr bool
	}{
		{"valid params", 1, 50, false},
		{"valid page 1000", 1000, 50, false},
		{"valid size 100", 1, 100, false},
		{"page below min", 0, 50, true},
		{"page above max", 1001, 50, true},
		{"size below min", 1, 0, true},
		{"size above max", 1, 101, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PaginateParams{Page: tt.page, Size: tt.size}
			err := p.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("PaginateParams.Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("PaginateParams.Validate() unexpected error: %v", err)
			}
		})
	}
}
