package validation

import "testing"

func TestSecureRandomInt(t *testing.T) {
	// Test generates values within range
	for i := 0; i < 100; i++ {
		val, err := SecureRandomInt(100, 999)
		if err != nil {
			t.Errorf("SecureRandomInt() returned error: %v", err)
			continue
		}
		if val < 100 || val > 999 {
			t.Errorf("SecureRandomInt(100, 999) = %d, want [100, 999]", val)
		}
	}
}

func TestSecureRandomInt_RangeEdge(t *testing.T) {
	// Test edge cases
	for i := 0; i < 10; i++ {
		val, err := SecureRandomInt(1, 1)
		if err != nil {
			t.Errorf("SecureRandomInt(1, 1) returned error: %v", err)
			continue
		}
		if val != 1 {
			t.Errorf("SecureRandomInt(1, 1) = %d, want 1", val)
		}
	}
}
