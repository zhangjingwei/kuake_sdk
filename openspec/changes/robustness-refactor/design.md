# Robustness Refactoring - Design Document

## Context

当前 Kuake SDK 的参数处理逻辑分散在各处，缺乏统一的校验层：

1. **散在的校验逻辑**：路径处理、参数校验、类型断言等逻辑散落在各文件
2. **不一致的错误处理**：不同方法使用不同的错误码格式
3. **安全漏洞**：路径穿越、类型断言 panic、可预测随机数等问题
4. **缺乏默认值处理**：可选参数的默认值处理不统一

## Goals / Non-Goals

**Goals:**
- 实现统一的参数校验层，覆盖所有 SDK 公共方法的外部输入
- 统一错误码格式为 `MODULE_ERRCODE`
- 修复所有高危安全漏洞（路径穿越、类型断言 panic、随机数安全）
- 提供可选参数的默认值注入
- 确保参数校验覆盖率 ≥ 80%

**Non-Goals:**
- 不修改 HTTP 传输层（请求头、Cookie、超时、重试）
- 不改变业务逻辑层（分享权限、容量限制等）
- 不修改日志/监控等横切关注点
- 不改变现有 API 的调用方式（保持向后兼容）

## Decisions

### 1. 参数校验层架构

采用**装饰器模式 + 中间件模式**的组合方案：

```
┌─────────────────────────────────────────────────────┐
│              Callers (CLI/Other SDKs)               │
└────────────────┬────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────┐
│     Validation Middleware / Decorator Layer         │
│  - Non-empty Check                                  │
│  - Type Validation                                  │
│  - Format Validation                                │
│  - Range Validation                                 │
│  - Path Sanitization                                │
└────────────────┬────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────┐
│              Business Logic (sdk/*.go)              │
└─────────────────────────────────────────────────────┘
```

### 2. 统一参数校验装饰器规范

**接口签名示例：**

```go
// validation/validator.go
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

**常用校验器：**

```go
// validation/checkers.go
package validation

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

// ValidPath 路径安全校验
func ValidPath() *Validator {
	return NewValidator("valid_path", func(value interface{}) error {
		if s, ok := value.(string); ok {
			// 规范化路径
			cleaned := filepath.Clean(s)
			// 检查是否包含 .. 
			if strings.Contains(cleaned, "..") {
				return ErrInvalidArgument("path contains invalid segments")
			}
			// Windows 路径检查
			if strings.Contains(s, "\\\\UNC\\") {
				return ErrInvalidArgument("UNC paths are not allowed")
			}
		}
		return nil
	})
}

// SafeTypeAssert 安全类型断言辅助函数
func SafeTypeAssert[T any](value interface{}) (T, bool) {
	v, ok := value.(T)
	return v, ok
}
```

**使用示例：**

```go
// sdk/file.go
func (qc *QuarkClient) UploadFile(filePath, destPath string, ...) (*StandardResponse, error) {
	// 统一参数校验
	validation.Chain(
		validation.NonEmpty(),
		validation.ValidPath(),
	).Validate(filePath)
	
	validation.Chain(
		validation.NonEmpty(),
		validation.ValidPath(),
	).Validate(destPath)
	
	// 业务逻辑...
}
```

### 3. 默认值注入规范

```go
// validation/default.go
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

// 示例：分页参数默认值
var PageDefaults = validation.NewDefaults(map[string]interface{}{
	"page": 1,
	"size": 50,
	"order_field": "created_at",
	"order_type": "desc",
})
```

### 4. 统一错误响应规范

```go
// validation/errors.go
package validation

// ErrorType 错误类型
type ErrorType int

const (
	ErrorTypeValidation ErrorType = iota
	ErrorTypeBusiness
	ErrorTypeSystem
)

// ValidationError 校验错误
type ValidationError struct {
	Code    string // MODULE_ERRCODE 格式
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

### 5. 分页参数校验规范

```go
// validation/pagination.go
package validation

// PaginateParams 分页参数
type PaginateParams struct {
	Page  int `json:"page"`
	Size  int `json:"size"`
}

// Validate 分页参数校验
func (p PaginateParams) Validate() error {
	// Page: [1, 1000], Size: [1, 100]
	if p.Page < 1 || p.Page > 1000 {
		return ErrInvalidArgument("page must be in range [1, 1000]")
	}
	if p.Size < 1 || p.Size > 100 {
		return ErrInvalidArgument("size must be in range [1, 100]")
	}
	return nil
}
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| 性能开销：额外的校验层可能带来性能影响 | 校验逻辑尽量简洁；对于高频调用，提供可选的"快速模式"跳过部分非关键校验 |
| 向后兼容性：新增校验可能拒绝原有合法输入 | 提供配置开关，允许逐步启用校验；文档明确参数要求 |
| 代码量增加：新增校验层会增加代码量 | 校验逻辑复用；统一错误处理减少重复代码 |

## Migration Plan

### 阶段 1：创建统一校验层（1-2天）
1. 创建 `sdk/validation/` 目录
2. 实现核心校验器（NonEmpty, InRange, ValidPath 等）
3. 实现安全的类型断言辅助函数
4. 编写单元测试（覆盖率 ≥ 80%）

### 阶段 2：修复高危漏洞（2-3天）
1. **VULN-001**：修复 `sdk/file.go:814` 路径穿越漏洞
2. **VULN-002**：修复 `sdk/*.go` 类型断言 panic
3. **VULN-003**：修复 `cmd/main.go:1054` 类型断言 panic

### 阶段 3：应用校验层（1-2天）
1. **VULN-004**：修复 `sdk/share.go:51-52` 随机数安全
2. **VULN-005**：修复 `sdk/share.go:533-539` 分页参数未限制
3. **VULN-006**：修复 `cmd/main.go:359-387` 空指针/越界访问

### 阶段 4：文档与验收（1天）
1. 更新 API 文档
2. 运行 `go test -cover` 验证覆盖率
3. 代码审查

## 5 个高危修复点

### 修复点 1：路径穿越漏洞 (VULN-001)

**漏洞描述：**
`os.Open(filePath)` 未校验路径，攻击者可通过 `../../../` 读取任意文件。

**影响范围：**
- `sdk/file.go:814` - `UploadFile` 方法
- `sdk/file.go` 其他文件操作方法

**修复方案概要：**
```go
// 在 sdk/validation/path.go 中实现
func ValidatePathSecurity(path string) error {
	// 规范化路径
	cleaned := filepath.Clean(path)
	
	// 检查是否包含 .. 或 ./
	if strings.Contains(cleaned, "..") {
		return ErrPathTraversal(path)
	}
	
	// Windows 特殊检查
	if strings.HasPrefix(path, "\\\\") || strings.HasPrefix(path, "\\\\UNC\\") {
		return ErrPathTraversal(path)
	}
	
	// 检查绝对路径是否在允许的基目录下（可选）
	return nil
}
```

### 修复点 2：类型断言 panic (VULN-002)

**漏洞描述：**
`data["field"].(string)` 未做类型检查，异常数据可导致 panic。

**影响范围：**
- `sdk/*.go` - 所有解析 API 响应的代码
- 典型模式：`data["field"].(string)`、`data["num"].(float64)`

**修复方案概要：**
```go
// 在 sdk/validation/typeassert.go 中实现
func SafeString(data map[string]interface{}, key string) (string, bool) {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

func SafeFloat64(data map[string]interface{}, key string) (float64, bool) {
	if v, ok := data[key]; ok {
		if f, ok := v.(float64); ok {
			return f, true
		}
	}
	return 0, false
}

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

### 修复点 3：CLI 参数类型断言 panic (VULN-003)

**漏洞描述：**
`strconv.Atoi(args[1])` 未处理转换失败。

**影响范围：**
- `cmd/main.go:1054` - `handleShareCreate` 方法

**修复方案概要：**
```go
// 在 cmd/validation/args.go 中实现
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

// 在 handleShareCreate 中使用
if days, ok := ParseIntArg(args, 1); !ok {
    return &CLIResult{Code: "INVALID_ARGS", Message: "days must be a valid integer"}
}
```

### 修复点 4：可预测随机数 (VULN-004)

**漏洞描述：**
使用 `math/rand` 生成随机数，种子可预测。

**影响范围：**
- `sdk/share.go:51-52` - `GetShareStoken` 方法

**修复方案概要：**
```go
// 在 sdk/validation/random.go 中实现
import (
    "crypto/rand"
    "math/big"
)

// SecureRandomInt 生成安全的随机整数
func SecureRandomInt(min, max int) (int, error) {
	maxBig := big.NewInt(int64(max))
	n, err := rand.Int(rand.Reader, maxBig)
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

// 在 sdk/share.go 中使用
// 原代码：
// rand.Seed(time.Now().UnixNano())
// dt := rand.Intn(900) + 100
// 新代码：
dt, err := SecureRandomInt(100, 999)
if err != nil {
    return nil, fmt.Errorf("failed to generate random value: %w", err)
}
```

### 修复点 5：分页参数未限制 (VULN-005)

**漏洞描述：**
`page, size` 未校验范围，可能导致数据库慢查询或内存溢出。

**影响范围：**
- `sdk/share.go:533-539` - `GetMyShareList` 方法

**修复方案概要：**
```go
// 在 sdk/validation/pagination.go 中实现（已在 Decisions 部分提供）
// Validate 方法已定义：
// Page: [1, 1000], Size: [1, 100]

// 在 GetMyShareList 中使用
func (qc *QuarkClient) GetMyShareList(page, size int, ...) (map[string]interface{}, error) {
    // 校验分页参数
    if page < 1 || page > 1000 {
        return nil, ErrInvalidArgument("page must be in range [1, 1000]")
    }
    if size < 1 || size > 100 {
        return nil, ErrInvalidArgument("size must be in range [1, 100]")
    }
    
    // 业务逻辑...
}
```

### 修复点 6：空指针/越界访问 (VULN-006)

**漏洞描述：**
`extractPathFromJSON` 未校验数据结构完整性，直接访问嵌套字段可能导致 panic。

**影响范围：**
- `cmd/main.go:359-387` - `extractPathFromJSON` 函数

**修复方案概要：**
```go
// 在 cmd/validation/json.go 中实现
func SafeExtractPathFromJSON(jsonStr string) (string, string, error) {
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
        return "", "", ErrInvalidArgument("invalid JSON format")
    }
    
    var path, fid string
    
    // 安全访问 data 字段
    if dataObj, ok := SafeMap(data, "data"); ok {
        if path, ok = SafeString(dataObj, "path"); !ok {
            path = ""
        }
        if fid, ok = SafeString(dataObj, "fid"); !ok {
            fid = ""
        }
    }
    
    // 安全访问根字段
    if path == "" {
        if path, ok = SafeString(data, "path"); !ok {
            path = ""
        }
    }
    if fid == "" {
        if fid, ok = SafeString(data, "fid"); !ok {
            fid = ""
        }
    }
    
    return path, fid, nil
}

// 辅助函数
func SafeMap(data map[string]interface{}, key string) (map[string]interface{}, bool) {
    if v, ok := data[key]; ok {
        if m, ok := v.(map[string]interface{}); ok {
            return m, true
        }
    }
    return nil, false
}
```

## ValidPathResult API（与 `parameter-validation` 对齐）

- **函数：** `validation.ValidPathResult(path string) (cleaned string, err error)`
- **语义：** 先执行与 `ValidPath()` 相同的安全检查；通过则返回远程路径风格的规范化结果（反斜杠转为 `/`、合并重复 `/`、非根路径去尾 `/`），供 `UploadFile` / `DownloadFile` 等使用。

## FID 格式（`ValidFID`）

- 非空字符串，长度 1–128。
- 字符集：`[0-9a-zA-Z_-]`。

## 24 方法校验矩阵（交叉引用）

各 `QuarkClient` 导出方法的入参校验策略见 [`public-methods-validation-tasks.md`](public-methods-validation-tasks.md) 章节「24 方法校验矩阵」。与 [`specs/parameter-validation/spec.md`](specs/parameter-validation/spec.md) **Implementation Rules 第 1 条**一致：该矩阵为验收与代码审查的单一事实来源（与 `sdk/quark_methods_b15_test.go` 同步更新）。

## 与 parameter-validation 偏差表（实现对照）

| 规范条目 | 实现说明 |
|----------|----------|
| `ValidPath` 中 UNC/NUL | 使用 `FILE_INVALID_PATH`（`ErrInvalidPath`），非 `FILE_PATH_TRAVERSAL` |
| 路径段 `.` / `..` | `FILE_PATH_TRAVERSAL`（`ErrPathTraversal`） |
| 错误消息语言 | `ErrInvalidArgument` / 校验器默认文案为中文；`ErrPathTraversal` 消息为中文 |
| `NonEmpty` | 仅接受 `string`；`nil` 与其它类型报错 |
| `InRange` | 不接受 `float64`；非 `int` 且非 `float64` 时保持不报错（与仅校验 int 的旧行为兼容） |
| **校验失败返回形态（与 `specs/parameter-validation/spec.md` Implementation Rules 第 2–3 条对齐）** | 见下节「校验失败：`StandardResponse` 与 `*ValidationError`」；**并非**所有失败都必须返回 `StandardResponse`。 |

## 校验失败：`StandardResponse` 与 `*ValidationError`（交叉引用）

与 [`specs/parameter-validation/spec.md`](specs/parameter-validation/spec.md) **Implementation Rules 第 2–3 条**一致：

- **返回 `(*StandardResponse, error)` 的方法**（如 `UploadFile`、`CreateFolder` 等）：参数校验失败时，在 **`StandardResponse` 结果**中设置 `Success: false`、`Code` 为规范错误码；`error` 通常为 `nil`。
- **返回 `(业务数据, error)` 且签名中不含 `StandardResponse` 的方法**（如返回 `map[string]interface{}` 的 `GetShareList`、`GetMyShareList` 等）：为保持向后兼容、**不改变方法签名**，校验失败时在 **`error` 中返回 `*validation.ValidationError`**（含 `Code`、`Message`），调用方应使用 `errors.As(err, &validation.ValidationError{})` 判别。

### `QuarkClient` 导出方法返回类型一览（便于验收）

| 返回类型 | 导出方法（与 `micro-tasks-tdd.md` B.15.1 一致） |
|----------|------|
| `(*StandardResponse, error)` | `GetUserInfo`，以及 `UploadFile`、`CreateFolder`、`Copy`、`Move`、`Rename`、`List`、`GetFileInfo`、`Delete` |
| 其它业务类型 + `error`（参数校验失败时可为 `*validation.ValidationError`，依方法实现而定） | `GetShareInfo`、`GetShareStoken`、`GetShareList`、`SaveShareFile`、`CreateShare`、`GetShareLink`、`GetMyShareList`、`GetShareIDByFid`、`GetDownloadURL` |
| 仅 `error` | `SetSharePassword`、`DeleteShare`、`DownloadFile` |
| 无 `error` / 非请求 API | `SetBaseURL`、`GetCookies`、`ConvertToFileInfo` |

**说明：** `GetDownloadURL` 为 `(string, error)`，**不是** `StandardResponse`。凡返回 `StandardResponse` 的方法，校验失败形态见上节第一点；返回 `map` / 指针 / `string` + `error` 的方法，见上节第二点。

## Open Questions

1. 是否需要在 CLI 层也应用相同的校验逻辑，还是仅在 SDK 层？
2. 是否需要引入独立的配置开关来控制校验的严格程度？
3. 是否需要添加校验失败的统计指标（如通过 Prometheus）？
