package validation

import (
	"strings"
	"testing"
)

func TestErrInvalidPath_Code(t *testing.T) {
	err := ErrInvalidPath("bad path")
	if err.Code != "FILE_INVALID_PATH" {
		t.Errorf("Code = %q, want FILE_INVALID_PATH", err.Code)
	}
	if err.Type != ErrorTypeValidation {
		t.Errorf("Type = %v, want ErrorTypeValidation", err.Type)
	}
}

func TestErrUserInvalidInput_Code(t *testing.T) {
	err := ErrUserInvalidInput("bad input")
	if err.Code != "USER_INVALID_INPUT" {
		t.Errorf("Code = %q, want USER_INVALID_INPUT", err.Code)
	}
	if err.Type != ErrorTypeValidation {
		t.Errorf("Type = %v, want ErrorTypeValidation", err.Type)
	}
}

func TestErrInvalidArgument_MessageChinese(t *testing.T) {
	err := NonEmpty().Validate("")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "空") {
		t.Errorf("expected Chinese empty hint, got %q", err.Error())
	}
}

func TestErrPathTraversal_MessageChinese(t *testing.T) {
	err := ErrPathTraversal("/x/y")
	if !strings.Contains(err.Error(), "路径") {
		t.Errorf("expected Chinese path hint, got %q", err.Error())
	}
}
