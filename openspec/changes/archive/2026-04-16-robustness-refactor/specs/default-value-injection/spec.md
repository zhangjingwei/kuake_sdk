# Default Value Injection Specification

## Purpose

定义统一的默认值注入层规范，为可选参数提供默认值处理，确保 SDK 方法在参数缺失时仍能正常工作。

## Requirements

### Requirement: 默认值配置结构

#### Defaults 配置结构

- **SHALL** 支持 `map[string]interface{}` 类型的默认值配置
- **SHALL** 支持动态添加默认值
- **SHALL** 支持合并多个默认值配置

#### Scenario: 创建默认值配置

- **WHEN** 调用 `NewDefaults(values)`
- **THEN** 返回 `Defaults` 配置对象

### Requirement: 默认值应用

#### Apply 方法

- **WHEN** 参数中不存在某键 **或** 该键值为零值（`nil`、`""`、`0` 等）
- **THEN** 使用默认值填充
- **WHEN** 参数中已存在某键且值为非零值
- **THEN** 不覆盖已有值

（实现见 `sdk/validation/default.go` 的 `isZeroValue`。）

#### Scenario: 分页参数默认值

- **WHEN** `GetMyShareList` 未提供 `page` 参数
- **THEN** 使用默认值 `1`
- **WHEN** 未提供 `size` 参数
- **THEN** 使用默认值 `50`
- **WHEN** 未提供 `order_field` 参数
- **THEN** 使用默认值 `"created_at"`
- **WHEN** 未提供 `order_type` 参数
- **THEN** 使用默认值 `"desc"`

### Requirement: 默认值配置示例

#### 分页参数默认值配置

```go
var PageDefaults = NewDefaults(map[string]interface{}{
    "page":  1,
    "size":  50,
    "order_field": "created_at",
    "order_type":  "desc",
})
```

#### 上传选项默认值配置

```go
var UploadOptionsDefaults = NewDefaults(map[string]interface{}{
    "policy": "skip",
})
```

### Requirement: 与校验层集成

#### 校验前注入默认值

- **WHEN** 方法调用时参数可选
- **THEN** 先注入默认值，再进行校验

#### Scenario: 完整的参数处理流程

1. 接收原始参数
2. 应用默认值注入
3. 执行参数校验
4. 进入业务逻辑

## API Specification

### 默认值配置

```go
package validation

// Defaults 参数默认值配置
type Defaults struct {
    values map[string]interface{}
}

// NewDefaults 创建默认值配置
func NewDefaults(values map[string]interface{}) *Defaults

// Add 添加默认值
func (d *Defaults) Add(key string, value interface{})

// Set 设置默认值（覆盖已存在的键）
func (d *Defaults) Set(key string, value interface{})

// Get 获取默认值
func (d *Defaults) Get(key string) (interface{}, bool)

// Merge 合并默认值配置
func (d *Defaults) Merge(other *Defaults)

// Apply 应用默认值到参数
func (d *Defaults) Apply(params map[string]interface{})
```

### 默认值配置实例

```go
// 分页参数默认值
var PageDefaults = NewDefaults(map[string]interface{}{
    "page":  1,
    "size":  50,
    "order_field": "created_at",
    "order_type":  "desc",
})

// 上传选项默认值
var UploadOptionsDefaults = NewDefaults(map[string]interface{}{
    "policy": "skip",
})
```

## Implementation Rules

1. 所有可选参数应配置默认值
2. 默认值应在参数校验前注入
3. 默认值配置应集中管理
4. 默认值应符合业务逻辑要求
