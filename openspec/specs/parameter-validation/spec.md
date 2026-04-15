# Parameter Validation Specification

## Purpose

定义统一的参数校验层规范，确保所有 SDK 公共方法的外部输入参数经过统一校验，防止安全漏洞和运行时错误。

## Requirements

### Requirement: 校验层结构

参数校验层应包含以下组件：

1. **核心校验器**：非空、类型、格式、范围校验
2. **路径安全校验**：防止路径穿越攻击
3. **安全类型断言**：避免 panic 的类型转换
4. **默认值注入**：为可选参数提供默认值
5. **统一错误响应**：标准化的错误码和错误消息

#### Scenario: 校验层调用

- **WHEN** 调用 SDK 公共方法
- **THEN** 参数在校验层进行校验，校验失败返回**标准化错误信息**（见下方 Implementation Rules 第 2 条：`StandardResponse` 或 `*validation.ValidationError`）
- **AND** 校验通过后进入业务逻辑层

### Requirement: 核心校验器

#### NonEmpty - 非空校验

校验参数非空（字符串非空白、非 nil）。

- **WHEN** 参数为 `nil` 或空字符串（仅空白）
- **THEN** 返回 `ErrInvalidArgument` 错误

#### ValidPath - 路径安全校验

校验文件路径安全性，防止路径穿越。

- **WHEN** 参数包含独立路径段 `.` 或 `..`
- **THEN** 返回 `ErrPathTraversal`（`FILE_PATH_TRAVERSAL`）
- **WHEN** 参数为 Windows UNC、保留路径形式或含 NUL 等
- **THEN** 返回 `ErrInvalidPath`（`FILE_INVALID_PATH`）
- **WHEN** 路径通过安全检查
- **THEN** 可由 `ValidPathResult` 返回规范化后的远程路径（斜杠统一、去尾斜杠等）

#### NonEmpty - 仅字符串

- **WHEN** 参数非 `string` 且非 `nil`
- **THEN** 返回 `ErrInvalidArgument`（非空校验仅支持字符串）

#### InRange - 仅接受 Go `int`

`InRange` 只对 **`interface{}` 中实际类型为 `int`** 的值做 `[min, max]` 校验；其它类型一律视为无效入参（早失败），避免静默跳过。

- **WHEN** 参数为 `int` 且在 `[min, max]` 内
- **THEN** 校验通过
- **WHEN** 参数为 `int` 且越界
- **THEN** 返回 `ErrInvalidArgument`，消息包含范围信息
- **WHEN** 参数为 `float64`（含 JSON 数字常见类型）
- **THEN** 返回 `ErrInvalidArgument`（整型范围校验不接受浮点数；浮点区间请用 `InRangeFloat64`）
- **WHEN** 参数为任何其他类型（含 `nil`、`string`、`int64`、`bool` 等）
- **THEN** 返回 `ErrInvalidArgument`（整型范围校验仅接受 `int` 类型）

### Requirement: 安全类型断言

提供安全的类型断言辅助函数，避免直接使用 `.(type)` 导致 panic。

#### SafeString - 安全字符串转换

- **WHEN** 值为字符串类型
- **THEN** 返回字符串和 `true`
- **WHEN** 值为其他类型或不存在
- **THEN** 返回空字符串和 `false`

#### SafeFloat64 - 安全浮点数转换

- **WHEN** 值为浮点数类型
- **THEN** 返回浮点数和 `true`
- **WHEN** 值为其他类型或不存在
- **THEN** 返回 0 和 `false`

#### SafeInt - 安全整数转换

- **WHEN** 值为整数或浮点数类型（JSON 解析的数字）
- **THEN** 返回整数和 `true`
- **WHEN** 值为其他类型或不存在
- **THEN** 返回 0 和 `false`

### Requirement: 分页参数校验

#### Page 范围校验

- **WHEN** `page < 1` 或 `page > 1000`
- **THEN** 返回 `ErrInvalidArgument` 错误

#### Size 范围校验

- **WHEN** `size < 1` 或 `size > 100`
- **THEN** 返回 `ErrInvalidArgument` 错误

### Requirement: 默认值注入

#### 可选参数默认值

- **WHEN** 可选参数未提供（nil 或零值）
- **THEN** 使用预定义的默认值

**默认值配置：**

| 参数 | 默认值 |
|------|--------|
| `page` | 1 |
| `size` | 50 |
| `order_field` | "created_at" |
| `order_type` | "desc" |

### Requirement: 统一错误响应

#### 错误码格式

错误码格式为 `MODULE_ERRCODE`：

- `FILE_INVALID_PATH` - 文件路径无效
- `FILE_PATH_TRAVERSAL` - 路径穿越攻击
- `USER_INVALID_INPUT` - 用户输入无效
- `INVALID_ARG` - 通用无效参数

#### 错误消息规范

- 中文消息用于用户显示
- 错误包含详细的日志信息（如无效值、预期值）

## API Specification

### 校验器接口

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
func NewValidator(name string, fn ValidatorFunc) *Validator

// Validate 执行校验
func (v *Validator) Validate(value interface{}) error

// Chain 链式校验
func Chain(validators ...*Validator) ValidatorFunc
```

### 校验器函数

```go
// 非空校验
func NonEmpty() *Validator

// 范围校验
func InRange(min, max int) *Validator

// 范围校验（float64）
func InRangeFloat64(min, max float64) *Validator

// 路径安全校验
func ValidPath() *Validator

// 路径安全校验通过后返回规范化远程路径
func ValidPathResult(path string) (cleaned string, err error)

// 邮箱格式校验
func ValidEmail() *Validator

// FID 格式校验
func ValidFID() *Validator
```

### 默认值注入

```go
// Defaults 参数默认值配置
type Defaults struct {
    values map[string]interface{}
}

// NewDefaults 创建默认值配置
func NewDefaults(values map[string]interface{}) *Defaults

// Apply 应用默认值
func (d *Defaults) Apply(params map[string]interface{})
```

### 分页参数

```go
// PaginateParams 分页参数
type PaginateParams struct {
    Page  int `json:"page"`
    Size  int `json:"size"`
}

// Validate 分页参数校验
func (p PaginateParams) Validate() error
```

### 错误类型

```go
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

func (e *ValidationError) Error() string

// ErrInvalidArgument 创建校验错误
func ErrInvalidArgument(msg string) *ValidationError

// ErrPathTraversal 创建路径穿越错误
func ErrPathTraversal(path string) *ValidationError
```

## Implementation Rules

1. 所有 SDK 公共方法的外部输入参数必须经过校验层。**方法级校验清单（入参字段、校验器、备注）**见归档变更文档 [`public-methods-validation-tasks.md`](../../../openspec/changes/archive/2026-04-16-robustness-refactor/public-methods-validation-tasks.md) 章节「24 方法校验矩阵」，并与 `sdk/quark_methods_b15_test.go` 中的导出方法表对照维护。
2. **校验失败时的返回形态（与向后兼容并存）：**
   - **返回 `(*StandardResponse, error)` 的方法**：校验失败时 **SHALL** 在**结果**中返回 `StandardResponse`，且 `Success` 为 `false`，`Code` 为规范错误码（如 `INVALID_ARG`、`FILE_PATH_TRAVERSAL` 等）；`error` 通常为 `nil`，与现有调用约定一致。
   - **返回业务数据与 `error`、且签名中不含 `StandardResponse` 的方法**（例如返回 `map[string]interface{}` 等）：为 **保持向后兼容、不改变现有方法签名**，校验失败时 **SHALL** 在 **`error` 中返回 `*validation.ValidationError`**（含 `Code` / `Message`），**不**将失败结果强行塞进 `StandardResponse`。上述两类均属「标准化错误信息」，差异仅在**错误信息的承载位置**（结构体字段 vs `error` 值）。
3. **向后兼容**：不改变现有公开方法的参数列表与返回类型；新增校验不得要求调用方改为只认 `StandardResponse`（除非该方法原本即返回 `StandardResponse`）。若未来需将「仅返回 `error`」的 API 统一为 `StandardResponse`，应通过 **新增方法**、**新版本 SDK** 或 **主版本升级** 单独立项，而非在本变更中改写现有签名。
4. 在方法注释中明确参数要求
5. **API JSON 响应体**字段读取允许双值断言（`v, ok := …`）等安全分支，**不强制**全盘改为 `validation.Safe*`；**外部入参**仍 **SHALL** 经校验层或已约定的 `Safe*` 路径，禁止依赖单值 `.(T)` 导致 panic。
