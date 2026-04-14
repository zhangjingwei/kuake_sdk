package validation

import (
	"testing"
)

func TestValidator(t *testing.T) {
	v := NewValidator("test", func(value interface{}) error {
		if value == nil {
			return ErrInvalidArgument("value is nil")
		}
		return nil
	})

	err := v.Validate("test")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	err = v.Validate(nil)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestChain(t *testing.T) {
	v1 := NewValidator("non-nil", func(value interface{}) error {
		if value == nil {
			return ErrInvalidArgument("value is nil")
		}
		return nil
	})

	v2 := NewValidator("non-empty", func(value interface{}) error {
		if s, ok := value.(string); ok && s == "" {
			return ErrInvalidArgument("value is empty")
		}
		return nil
	})

	chain := Chain(v1, v2)

	err := chain("test")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	err = chain(nil)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	err = chain("")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestErrInvalidArgument(t *testing.T) {
	err := ErrInvalidArgument("test message")
	if err.Code != "INVALID_ARG" {
		t.Errorf("expected code INVALID_ARG, got %s", err.Code)
	}
	if err.Type != ErrorTypeValidation {
		t.Errorf("expected type ErrorTypeValidation, got %d", err.Type)
	}
}

func TestErrPathTraversal(t *testing.T) {
	err := ErrPathTraversal("../etc/passwd")
	if err.Code != "FILE_PATH_TRAVERSAL" {
		t.Errorf("expected code FILE_PATH_TRAVERSAL, got %s", err.Code)
	}
	if err.Type != ErrorTypeValidation {
		t.Errorf("expected type ErrorTypeValidation, got %d", err.Type)
	}
}
