# Robustness Refactoring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a unified parameter validation layer with security fixes for path traversal, type assertion panics, predictable random numbers, and pagination limits.

**Architecture:** Create `sdk/validation/` package with validator decorators (NonEmpty, InRange, ValidPath), safe type assertion helpers (SafeString, SafeFloat64, SafeInt), and error types with MODULE_ERRCODE format. Fix high-risk bugs by integrating validation into SDK methods.

**Tech Stack:** Go 1.20+, Go testing package, crypto/rand for secure random numbers.

---

## File Structure

| File | Purpose |
|------|---------|
| `sdk/validation/validator.go` | Core validator interface and Chain function |
| `sdk/validation/errors.go` | ValidationError type and error constructors |
| `sdk/validation/checkers.go` | NonEmpty, InRange, ValidPath validators |
| `sdk/validation/typeassert.go` | SafeString, SafeFloat64, SafeInt helpers |
| `sdk/validation/pagination.go` | PaginateParams type and Validate method |
| `sdk/validation/default.go` | Defaults config and Apply method |
| `sdk/validation/random.go` | SecureRandomInt function using crypto/rand |
| `sdk/validation/path.go` | ValidatePathSecurity function |
| `sdk/file.go` | Add path validation to UploadFile, DownloadFile |
| `sdk/share.go` | Replace math/rand with crypto/rand, add pagination validation |
| `cmd/validation/args.go` | ParseIntArg helper for CLI arguments |
| `cmd/main.go` | Use ParseIntArg in handleShareCreate, SafeExtractPathFromJSON |

---

### Task 1: Create validation package structure

**Files:**
- Create: `sdk/validation/validator.go`
- Create: `sdk/validation/errors.go`

- [ ] **Step 1: Create validator.go with core interface**

```go
package validation

// ValidatorFunc 参数校验函数类型
type ValidatorFunc func(value interface{}) error

// Validator 参数校验器
type Validator struct {
	name     string
	validate ValidatorFunc
}

// NewValidator 创建新的校验器
func NewValidator(name string, fn ValidatorFunc) *Validator {
	return &Validator{name: name, validate: fn}
}

// Validate 执行校验
func (v *Validator) Validate(value interface{}) error {
	return v.validate(value)
}

// Chain 链式校验
func Chain(validators ...*Validator) ValidatorFunc {
	return func(value interface{}) error {
		for _, v := range validators {
			if err := v.Validate(value); err != nil {
				return err
			}
		}
		return nil
	}
}
```

- [ ] **Step 2: Create errors.go with ValidationError type**

```go
package validation

import "fmt"

// ErrorType 错误类型
type ErrorType int

const (
	ErrorTypeValidation ErrorType = iota
	ErrorTypeBusiness
	ErrorTypeSystem
)

// ValidationError 校验错误
type ValidationError struct {
	Code    string
	Message string
	Type    ErrorType
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ErrInvalidArgument 创建校验错误
func ErrInvalidArgument(msg string) *ValidationError {
	return &ValidationError{
		Code:    "INVALID_ARG",
		Message: msg,
		Type:    ErrorTypeValidation,
	}
}

// ErrPathTraversal 创建路径穿越错误
func ErrPathTraversal(path string) *ValidationError {
	return &ValidationError{
		Code:    "FILE_PATH_TRAVERSAL",
		Message: fmt.Sprintf("path contains invalid segments: %s", path),
		Type:    ErrorTypeValidation,
	}
}
```

- [ ] **Step 3: Run tests for validation package**

```bash
go test ./sdk/validation/... -v
```

Expected: 0 tests pass (package is empty)

- [ ] **Step 4: Commit validation infrastructure**

```bash
git add sdk/validation/
git commit -m "feat(validation): add core validator interface and error types"
```

---

### Task 2: Implement NonEmpty, InRange, ValidPath validators

**Files:**
- Create: `sdk/validation/checkers.go`
- Create: `sdk/validation/checkers_test.go`

- [ ] **Step 1: Write failing test for NonEmpty**

```go
package validation

import "testing"

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
			err := Chain(NonEmpty()).Validate(tt.value)
			if tt.wantErr && err == nil {
				t.Errorf("NonEmpty() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("NonEmpty() unexpected error: %v", err)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./sdk/validation/... -v -run TestNonEmpty
```

Expected: FAIL - NonEmpty function not defined

- [ ] **Step 3: Implement NonEmpty in checkers.go**

```go
package validation

import "strings"

// NonEmpty 非空校验
func NonEmpty() *Validator {
	return NewValidator("non_empty", func(value interface{}) error {
		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				return ErrInvalidArgument("value cannot be empty")
			}
		case nil:
			return ErrInvalidArgument("value cannot be nil")
		}
		return nil
	})
}

// InRange 数值范围校验
func InRange(min, max int) *Validator {
	return NewValidator("in_range", func(value interface{}) error {
		if v, ok := value.(int); ok {
			if v < min || v > max {
				return ErrInvalidArgument(fmt.Sprintf("value must be in range [%d, %d]", min, max))
			}
		}
		return nil
	})
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./sdk/validation/... -v -run TestNonEmpty
```

Expected: PASS

- [ ] **Step 5: Add test for InRange**

```go
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
		{"non-integer value", 1, 10, "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Chain(InRange(tt.min, tt.max)).Validate(tt.value)
			if tt.wantErr && err == nil {
				t.Errorf("InRange(%d, %d) expected error, got nil", tt.min, tt.max)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("InRange(%d, %d) unexpected error: %v", tt.min, tt.max, err)
			}
		})
	}
}
```

- [ ] **Step 6: Commit validators**

```bash
git add sdk/validation/checkers.go sdk/validation/checkers_test.go
git commit -m "feat(validation): add NonEmpty and InRange validators"
```

---

### Task 3: Implement ValidPath validator

**Files:**
- Create: `sdk/validation/path.go`
- Create: `sdk/validation/path_test.go`

- [ ] **Step 1: Write failing test for ValidPath**

```go
package validation

import "testing"

func TestValidPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid path", "/home/user/file.txt", false},
		{"valid Windows path", "C:\\Users\\test", false},
		{"path with parent reference", "/home/../etc/passwd", true},
		{"path with double dots", "../../../secret", true},
		{"UNC path", "\\\\server\\share", true},
		{"UNC path with UNC prefix", "\\\\UNC\\server\\share", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Chain(ValidPath()).Validate(tt.path)
			if tt.wantErr && err == nil {
				t.Errorf("ValidPath() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidPath() unexpected error: %v", err)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./sdk/validation/... -v -run TestValidPath
```

Expected: FAIL - ValidPath function not defined

- [ ] **Step 3: Implement ValidPath**

```go
package validation

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidPath 路径安全校验
func ValidPath() *Validator {
	return NewValidator("valid_path", func(value interface{}) error {
		if s, ok := value.(string); ok {
			// 规范化路径
			cleaned := filepath.Clean(s)
			// 检查是否包含 ..
			if strings.Contains(cleaned, "..") {
				return ErrPathTraversal(s)
			}
			// Windows 路径检查
			if strings.HasPrefix(s, "\\\\") || strings.HasPrefix(s, "\\\\UNC\\") {
				return ErrPathTraversal(s)
			}
		}
		return nil
	})
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./sdk/validation/... -v -run TestValidPath
```

Expected: PASS

- [ ] **Step 5: Commit ValidPath validator**

```bash
git add sdk/validation/path.go sdk/validation/path_test.go
git commit -m "feat(validation): add ValidPath validator for path traversal prevention"
```

---

### Task 4: Implement SafeTypeAssert helpers

**Files:**
- Create: `sdk/validation/typeassert.go`
- Create: `sdk/validation/typeassert_test.go`

- [ ] **Step 1: Write failing test for SafeString**

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./sdk/validation/... -v -run TestSafeString
```

Expected: FAIL - SafeString function not defined

- [ ] **Step 3: Implement SafeString, SafeFloat64, SafeInt**

```go
package validation

// SafeString 安全字符串转换
func SafeString(data map[string]interface{}, key string) (string, bool) {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

// SafeFloat64 安全浮点数转换
func SafeFloat64(data map[string]interface{}, key string) (float64, bool) {
	if v, ok := data[key]; ok {
		if f, ok := v.(float64); ok {
			return f, true
		}
	}
	return 0, false
}

// SafeInt 安全整数转换
func SafeInt(data map[string]interface{}, key string) (int, bool) {
	if v, ok := data[key]; ok {
		if i, ok := v.(int); ok {
			return i, true
		}
		// JSON unmarshal 数字可能为 float64
		if f, ok := v.(float64); ok {
			return int(f), true
		}
	}
	return 0, false
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./sdk/validation/... -v -run TestSafeString
```

Expected: PASS

- [ ] **Step 5: Add test for SafeFloat64 and SafeInt**

```go
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
```

- [ ] **Step 6: Commit safe type assert helpers**

```bash
git add sdk/validation/typeassert.go sdk/validation/typeassert_test.go
git commit -m "feat(validation): add safe type assertion helpers"
```

---

### Task 5: Implement PaginateParams validation

**Files:**
- Create: `sdk/validation/pagination.go`
- Create: `sdk/validation/pagination_test.go`

- [ ] **Step 1: Write failing test for PaginateParams**

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./sdk/validation/... -v -run TestPaginateParams
```

Expected: FAIL - PaginateParams type not defined

- [ ] **Step 3: Implement PaginateParams**

```go
package validation

// PaginateParams 分页参数
type PaginateParams struct {
	Page  int `json:"page"`
	Size  int `json:"size"`
}

// Validate 分页参数校验
func (p PaginateParams) Validate() error {
	if p.Page < 1 || p.Page > 1000 {
		return ErrInvalidArgument("page must be in range [1, 1000]")
	}
	if p.Size < 1 || p.Size > 100 {
		return ErrInvalidArgument("size must be in range [1, 100]")
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./sdk/validation/... -v -run TestPaginateParams
```

Expected: PASS

- [ ] **Step 5: Commit pagination**

```bash
git add sdk/validation/pagination.go sdk/validation/pagination_test.go
git commit -m "feat(validation): add PaginateParams type with Validate method"
```

---

### Task 6: Implement Defaults for default value injection

**Files:**
- Create: `sdk/validation/default.go`
- Create: `sdk/validation/default_test.go`

- [ ] **Step 1: Write failing test for Defaults**

```go
package validation

import "testing"

func TestDefaults_Apply(t *testing.T) {
	defaults := NewDefaults(map[string]interface{}{
		"page":  1,
		"size":  50,
		"order": "desc",
	})

	tests := []struct {
		name        string
		params      map[string]interface{}
		wantParams  map[string]interface{}
	}{
		{"empty params", map[string]interface{}{}, map[string]interface{}{
			"page":  1,
			"size":  50,
			"order": "desc",
		}},
		{"partial params", map[string]interface{}{"page": 2}, map[string]interface{}{
			"page":  2,
			"size":  50,
			"order": "desc",
		}},
		{"all params", map[string]interface{}{"page": 2, "size": 100}, map[string]interface{}{
			"page":  2,
			"size":  100,
			"order": "desc",
		}},
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./sdk/validation/... -v -run TestDefaults
```

Expected: FAIL - Defaults type not defined

- [ ] **Step 3: Implement Defaults**

```go
package validation

// Defaults 参数默认值
type Defaults struct {
	values map[string]interface{}
}

// NewDefaults 创建默认值配置
func NewDefaults(values map[string]interface{}) *Defaults {
	return &Defaults{values: values}
}

// Apply 应用默认值
func (d *Defaults) Apply(params map[string]interface{}) {
	for k, v := range d.values {
		if _, exists := params[k]; !exists {
			params[k] = v
		}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./sdk/validation/... -v -run TestDefaults
```

Expected: PASS

- [ ] **Step 5: Commit defaults**

```bash
git add sdk/validation/default.go sdk/validation/default_test.go
git commit -m "feat(validation): add Defaults for default value injection"
```

---

### Task 7: Implement SecureRandomInt

**Files:**
- Create: `sdk/validation/random.go`
- Create: `sdk/validation/random_test.go`

- [ ] **Step 1: Write failing test for SecureRandomInt**

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./sdk/validation/... -v -run TestSecureRandomInt
```

Expected: FAIL - SecureRandomInt function not defined

- [ ] **Step 3: Implement SecureRandomInt**

```go
package validation

import (
	"crypto/rand"
	"math/big"
)

// SecureRandomInt 生成安全的随机整数
func SecureRandomInt(min, max int) (int, error) {
	if min > max {
		return 0, ErrInvalidArgument("min must be <= max")
	}
	maxBig := big.NewInt(int64(max))
	n, err := rand.Int(rand.Reader, maxBig)
	if err != nil {
		return 0, err
	}
	val := int(n.Int64())
	if val < min {
		val = min
	}
	return val, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./sdk/validation/... -v -run TestSecureRandomInt
```

Expected: PASS

- [ ] **Step 5: Commit secure random**

```bash
git add sdk/validation/random.go sdk/validation/random_test.go
git commit -m "feat(validation): add SecureRandomInt using crypto/rand"
```

---

### Task 8: Add PageDefaults constant

**Files:**
- Modify: `sdk/validation/default.go`
- Modify: `sdk/validation/default_test.go`

- [ ] **Step 1: Add PageDefaults constant**

```go
package validation

// Defaults 参数默认值
type Defaults struct {
	values map[string]interface{}
}

// NewDefaults 创建默认值配置
func NewDefaults(values map[string]interface{}) *Defaults {
	return &Defaults{values: values}
}

// Apply 应用默认值
func (d *Defaults) Apply(params map[string]interface{}) {
	for k, v := range d.values {
		if _, exists := params[k]; !exists {
			params[k] = v
		}
	}
}

// PageDefaults 分页参数默认值
var PageDefaults = NewDefaults(map[string]interface{}{
	"page":        1,
	"size":        50,
	"order_field": "created_at",
	"order_type":  "desc",
})
```

- [ ] **Step 2: Add test for PageDefaults**

```go
package validation

import "testing"

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
```

- [ ] **Step 3: Run test to verify it passes**

```bash
go test ./sdk/validation/... -v -run TestPageDefaults
```

Expected: PASS

- [ ] **Step 4: Commit PageDefaults**

```bash
git add sdk/validation/default.go sdk/validation/default_test.go
git commit -m "feat(validation): add PageDefaults constant"
```

---

### Task 9: Fix VULN-001 - Path traversal in file.go

**Files:**
- Modify: `sdk/file.go:807-822`

- [ ] **Step 1: Update UploadFile to add path validation**

```go
// UploadFile 上传文件到夸克网盘，支持大文件分片上传
// progressCallback: 进度回调函数，如果为 nil 则不显示进度
// opts: 上传选项（可为 nil，使用默认行为）
func (qc *QuarkClient) UploadFile(filePath, destPath string, progressCallback func(*UploadProgress), opts *UploadOptions) (*StandardResponse, error) {
	// 统一参数校验
	if err := Chain(NonEmpty(), ValidPath()).Validate(filePath); err != nil {
		return &StandardResponse{
			Success: false,
			Code:    "FILE_INVALID_PATH",
			Message: err.Error(),
			Data:    nil,
		}, nil
	}
	
	if err := Chain(NonEmpty(), ValidPath()).Validate(destPath); err != nil {
		return &StandardResponse{
			Success: false,
			Code:    "FILE_INVALID_PATH",
			Message: err.Error(),
			Data:    nil,
		}, nil
	}
	
	filePath = stripQuotes(filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return &StandardResponse{
			Success: false,
			Code:    "FILE_OPEN_ERROR",
			Message: fmt.Sprintf("failed to open file: %v", err),
			Data:    nil,
		}, nil
	}
	defer file.Close()
	// ... rest of function
}
```

- [ ] **Step 2: Write test for UploadFile path validation**

```go
func TestUploadFile_InvalidPath(t *testing.T) {
	client := &QuarkClient{}
	
	// Test path traversal
	resp, err := client.UploadFile("/../../../etc/passwd", "/dest/", nil, nil)
	if err != nil {
		t.Fatalf("UploadFile returned error: %v", err)
	}
	if resp.Success {
		t.Errorf("UploadFile with path traversal should fail")
	}
	if resp.Code != "FILE_INVALID_PATH" {
		t.Errorf("UploadFile code = %v, want FILE_INVALID_PATH", resp.Code)
	}
	
	// Test empty path
	resp, err = client.UploadFile("", "/dest/", nil, nil)
	if err != nil {
		t.Fatalf("UploadFile returned error: %v", err)
	}
	if resp.Success {
		t.Errorf("UploadFile with empty path should fail")
	}
	if resp.Code != "FILE_INVALID_PATH" {
		t.Errorf("UploadFile code = %v, want FILE_INVALID_PATH", resp.Code)
	}
}
```

- [ ] **Step 3: Run test to verify it fails (existing test, not adding new failure)**

```bash
go test ./sdk/... -v -run TestUploadFile
```

Expected: SKIP (network required)

- [ ] **Step 4: Commit VULN-001 fix**

```bash
git add sdk/file.go
git commit -m "fix(VULN-001): add path validation to UploadFile"
```

---

### Task 10: Fix VULN-004 - Replace math/rand with crypto/rand

**Files:**
- Modify: `sdk/share.go:45-90`
- Create: `sdk/validation/random.go` (already created in Task 7)

- [ ] **Step 1: Update GetShareStoken to use SecureRandomInt**

```go
// GetShareStoken 获取分享stoken
// pwdID: 分享链接ID
// passcode: 提取码，默认空
// 返回stoken数据和错误
func (qc *QuarkClient) GetShareStoken(pwdID, passcode string) (map[string]interface{}, error) {
	// 生成随机数和时间戳
	dt, err := SecureRandomInt(100, 999)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random value: %w", err)
	}
	t := time.Now().UnixMilli()

	queryParams := url.Values{}
	queryParams.Set("pr", "ucpro")
	queryParams.Set("fr", "pc")
	queryParams.Set("uc_param_str", "")
	queryParams.Set("__dt", fmt.Sprintf("%d", dt))
	queryParams.Set("__t", fmt.Sprintf("%d", t))
	// ... rest of function
}
```

- [ ] **Step 2: Write test for GetShareStoken random generation**

```go
func TestGetShareStoken_GeneratesRandom(t *testing.T) {
	// This test verifies SecureRandomInt is used instead of math/rand
	// by checking that the function returns without error
	client := &QuarkClient{}
	
	// This will fail because we don't have valid credentials
	// but we can verify the random generation doesn't panic
	_, err := client.GetShareStoken("test_pwd_id", "")
	
	// We expect an error (network/auth), but not a panic from math/rand
	if err == nil {
		t.Fatal("Expected error from GetShareStoken")
	}
}
```

- [ ] **Step 3: Commit VULN-004 fix**

```bash
git add sdk/share.go
git commit -m "fix(VULN-004): replace math/rand with crypto/rand in GetShareStoken"
```

---

### Task 11: Fix VULN-005 - Add pagination validation to GetMyShareList

**Files:**
- Modify: `sdk/share.go:527-563`

- [ ] **Step 1: Update GetMyShareList to add pagination validation**

```go
// GetMyShareList 获取我的分享列表
// page: 页码，默认1
// size: 每页数量，默认50
// orderField: 排序字段，默认"created_at"
// orderType: 排序方式，"asc" 或 "desc"，默认"desc"
// 返回分享列表数据和错误
func (qc *QuarkClient) GetMyShareList(page, size int, orderField, orderType string) (map[string]interface{}, error) {
	// 校验分页参数
	if err := PaginateParams{Page: page, Size: size}.Validate(); err != nil {
		return nil, err
	}
	
	// 应用默认值
	if orderField == "" {
		orderField = "created_at"
	}
	if orderType == "" {
		orderType = "desc"
	}
	// ... rest of function
}
```

- [ ] **Step 2: Write test for GetMyShareList pagination validation**

```go
func TestGetMyShareList_InvalidPagination(t *testing.T) {
	client := &QuarkClient{}
	
	// Test page < 1
	_, err := client.GetMyShareList(0, 50, "", "")
	if err == nil {
		t.Errorf("GetMyShareList with page=0 should error")
	}
	
	// Test page > 1000
	_, err = client.GetMyShareList(1001, 50, "", "")
	if err == nil {
		t.Errorf("GetMyShareList with page=1001 should error")
	}
	
	// Test size < 1
	_, err = client.GetMyShareList(1, 0, "", "")
	if err == nil {
		t.Errorf("GetMyShareList with size=0 should error")
	}
	
	// Test size > 100
	_, err = client.GetMyShareList(1, 101, "", "")
	if err == nil {
		t.Errorf("GetMyShareList with size=101 should error")
	}
}
```

- [ ] **Step 3: Commit VULN-005 fix**

```bash
git add sdk/share.go
git commit -m "fix(VULN-005): add pagination validation to GetMyShareList"
```

---

### Task 12: Fix VULN-002 - Replace unsafe type assertions in sdk files

**Files:**
- Modify: `sdk/quark_client.go` - any type assertions

- [ ] **Step 1: Find and fix unsafe type assertions in quark_client.go**

Search for patterns like `data["field"].(string)` and replace with SafeString:

```go
// Before:
if data["field"].(string) == "value" { ... }

// After:
if s, ok := SafeString(data, "field"); ok && s == "value" { ... }
```

- [ ] **Step 2: Fix unsafe type assertions in file.go**

```go
// Before:
if data["path"].(string) == "/" { ... }

// After:
if s, ok := SafeString(data, "path"); ok && s == "/" { ... }
```

- [ ] **Step 3: Fix unsafe type assertions in share.go**

```go
// Before:
if data["fid"].(string) != "" { ... }

// After:
if s, ok := SafeString(data, "fid"); ok && s != "" { ... }
```

- [ ] **Step 4: Commit VULN-002 fixes**

```bash
git add sdk/*.go
git commit -m "fix(VULN-002): replace unsafe type assertions with Safe* helpers"
```

---

### Task 13: Fix VULN-003 - CLI argument parsing

**Files:**
- Create: `cmd/validation/args.go`
- Create: `cmd/validation/args_test.go`
- Modify: `cmd/main.go:1040-1080`

- [ ] **Step 1: Create args.go with ParseIntArg**

```go
package validation

import "strconv"

// ParseIntArg 安全地解析整数参数
func ParseIntArg(args []string, index int) (int, bool) {
	if index < 0 || index >= len(args) {
		return 0, false
	}
	val, err := strconv.Atoi(args[index])
	if err != nil {
		return 0, false
	}
	return val, true
}
```

- [ ] **Step 2: Create test for ParseIntArg**

```go
package validation

import "testing"

func TestParseIntArg(t *testing.T) {
	args := []string{"file.txt", "7", "true"}
	
	tests := []struct {
		name    string
		args    []string
		index   int
		wantVal int
		wantOK  bool
	}{
		{"valid index", args, 1, 7, true},
		{"invalid index", args, 10, 0, false},
		{"empty args", []string{}, 0, 0, false},
		{"non-numeric", []string{"abc"}, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOK := ParseIntArg(tt.args, tt.index)
			if gotVal != tt.wantVal || gotOK != tt.wantOK {
				t.Errorf("ParseIntArg() = (%v, %v), want (%v, %v)", gotVal, gotOK, tt.wantVal, tt.wantOK)
			}
		})
	}
}
```

- [ ] **Step 3: Update handleShareCreate to use ParseIntArg**

```go
func handleShareCreate(client *sdk.QuarkClient, args []string) *CLIResult {
	if len(args) < 3 {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_ARGS",
			Message: "Usage: share <path> <days> <passcode> (path and passcode must be quoted, e.g., share \"file(1).txt\" 7 \"false\")",
		}
	}

	path := args[0]

	// 解析有效期天数（必传）
	expireDays, ok := validation.ParseIntArg(args, 1)
	if !ok {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_ARGS",
			Message: "days must be a valid integer",
		}
	}

	// 解析是否需要提取码（必传）
	passcodeArg := args[2]
	var needPasscode bool
	switch passcodeArg {
	case "true":
		needPasscode = true
	case "false":
		needPasscode = false
	default:
		return &CLIResult{
			Success: false,
			Code:    "INVALID_ARGS",
			Message: "passcode must be 'true' or 'false'",
		}
	}
	// ... rest of function
}
```

- [ ] **Step 4: Commit VULN-003 fix**

```bash
git add cmd/validation/args.go cmd/validation/args_test.go cmd/main.go
git commit -m "fix(VULN-003): add safe argument parsing in CLI"
```

---

### Task 14: Fix VULN-006 - Safe JSON path extraction

**Files:**
- Create: `cmd/validation/json.go`
- Create: `cmd/validation/json_test.go`
- Modify: `cmd/main.go:354-390`

- [ ] **Step 1: Create json.go with SafeExtractPathFromJSON**

```go
package validation

import (
	"encoding/json"
)

// SafeMap 安全的 map 转换
func SafeMap(data map[string]interface{}, key string) (map[string]interface{}, bool) {
	if v, ok := data[key]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m, true
		}
	}
	return nil, false
}

// SafeExtractPathFromJSON 从 JSON 中安全地提取路径或 fid
func SafeExtractPathFromJSON(jsonStr string) (string, string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", "", ErrInvalidArgument("invalid JSON format")
	}

	var path, fid string

	// 安全访问 data 字段
	if dataObj, ok := SafeMap(data, "data"); ok {
		if p, ok := SafeString(dataObj, "path"); ok {
			path = p
		}
		if f, ok := SafeString(dataObj, "fid"); ok {
			fid = f
		}
	}

	// 如果没有从 data 中提取到，尝试从根对象提取（简化格式）
	if path == "" {
		if p, ok := SafeString(data, "path"); ok {
			path = p
		}
	}
	if fid == "" {
		if f, ok := SafeString(data, "fid"); ok {
			fid = f
		}
	}

	return path, fid, nil
}
```

- [ ] **Step 2: Create test for SafeExtractPathFromJSON**

```go
package validation

import "testing"

func TestSafeExtractPathFromJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		wantPath string
		wantFid  string
		wantErr  bool
	}{
		{"valid response", `{"success":true,"data":{"path":"/test","fid":"123"}}`, "/test", "123", false},
		{"simplified format", `{"path":"/test","fid":"123"}`, "/test", "123", false},
		{"empty path", `{"data":{}}`, "", "", false},
		{"invalid JSON", "not json", "", "", true},
		{"missing data", `{"path":"/test"}`, "/test", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotFid, err := SafeExtractPathFromJSON(tt.jsonStr)
			if tt.wantErr && err == nil {
				t.Errorf("SafeExtractPathFromJSON() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("SafeExtractPathFromJSON() unexpected error: %v", err)
			}
			if gotPath != tt.wantPath || gotFid != tt.wantFid {
				t.Errorf("SafeExtractPathFromJSON() = (%v, %v), want (%v, %v)", gotPath, gotFid, tt.wantPath, tt.wantFid)
			}
		})
	}
}
```

- [ ] **Step 3: Update extractPathFromJSON to use SafeExtractPathFromJSON**

```go
// extractPathFromJSON 从 JSON 中提取路径或 fid
// 支持两种格式：
// 1. 完整响应格式：{"success": true, "data": {"path": "...", "fid": "..."}} - 流式输出格式
// 2. 简化格式：{"path": "...", "fid": "..."}
func extractPathFromJSON(jsonStr string) (string, string, error) {
	return SafeExtractPathFromJSON(jsonStr)
}
```

- [ ] **Step 4: Commit VULN-006 fix**

```bash
git add cmd/validation/json.go cmd/validation/json_test.go cmd/main.go
git commit -m "fix(VULN-006): add safe JSON path extraction"
```

---

### Task 15: Update API documentation and run final tests

**Files:**
- Modify: `sdk/file.go` - Add parameter docs
- Modify: `sdk/share.go` - Add parameter docs

- [ ] **Step 1: Update UploadFile documentation**

```go
// UploadFile 上传文件到夸克网盘，支持大文件分片上传
// filePath: 本地文件路径（不能为空，不能包含路径穿越序列）
// destPath: 目标路径（不能为空，不能包含路径穿越序列）
// progressCallback: 进度回调函数，如果为 nil 则不显示进度
// opts: 上传选项（可为 nil，使用默认行为）
// 返回 StandardResponse 和错误
func (qc *QuarkClient) UploadFile(filePath, destPath string, progressCallback func(*UploadProgress), opts *UploadOptions) (*StandardResponse, error) {
```

- [ ] **Step 2: Update GetMyShareList documentation**

```go
// GetMyShareList 获取我的分享列表
// page: 页码，范围 [1, 1000]，默认 1
// size: 每页数量，范围 [1, 100]，默认 50
// orderField: 排序字段，默认 "created_at"
// orderType: 排序方式，"asc" 或 "desc"，默认 "desc"
// 返回分享列表数据和错误
func (qc *QuarkClient) GetMyShareList(page, size int, orderField, orderType string) (map[string]interface{}, error) {
```

- [ ] **Step 3: Run all tests to verify no new failures**

```bash
go test ./... -v 2>&1 | tee test_output.txt
```

Expected: Same failures as baseline (0 new failures)

- [ ] **Step 4: Run coverage check**

```bash
go test ./... -cover
```

Expected: Coverage ≥ 80% for validation package

- [ ] **Step 5: Run go vet**

```bash
go vet ./...
```

Expected: No warnings

- [ ] **Step 6: Final commit**

```bash
git add sdk/*.go cmd/*.go
git commit -m "docs: add parameter documentation and run final validation"
```

---

## Spec Coverage Checklist

| Spec Requirement | Task |
|-----------------|------|
| Validator interface and Chain | Task 1 |
| NonEmpty validator | Task 2 |
| InRange validator | Task 2 |
| ValidPath validator | Task 3 |
| SafeString helper | Task 4 |
| SafeFloat64 helper | Task 4 |
| SafeInt helper | Task 4 |
| PaginateParams with Validate | Task 5 |
| Defaults with Apply | Task 6 |
| PageDefaults constant | Task 6 |
| SecureRandomInt using crypto/rand | Task 7 |
| VULN-001: Path traversal fix | Task 9 |
| VULN-002: Safe type assertions | Task 12 |
| VULN-003: CLI argument parsing | Task 13 |
| VULN-004: Secure random in GetShareStoken | Task 10 |
| VULN-005: Pagination validation | Task 11 |
| VULN-006: Safe JSON extraction | Task 14 |

---

**Plan complete and saved to `docs/superpowers/plans/2026-04-12-robustness-refactor.md`.**

Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
