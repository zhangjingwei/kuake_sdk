# Robustness Refactoring - Proposal

## Why

当前 Kuake SDK 存在多处参数处理安全漏洞和类型不安全的代码，可能导致程序崩溃（panic）或安全风险（如路径穿越攻击）。这些问题在 `sdk/file.go`、`sdk/share.go` 和 `cmd/main.go` 等多个文件中散在出现，缺乏统一的参数校验层，导致：

1. **路径穿越风险**：`os.Open(filePath)` 未校验路径，攻击者可通过 `../../../` 读取任意文件
2. **类型断言 panic**：直接使用 `data["field"].(string)` 进行类型断言，异常数据可导致程序崩溃
3. **可预测随机数**：使用 `math/rand` 而非 `crypto/rand`，种子可预测
4. **分页参数未限制**：`page` 和 `size` 参数未校验范围，可能导致数据库慢查询或内存溢出
5. **空指针/越界访问**：未校验数据结构完整性，直接访问嵌套字段

## What Changes

### 新增能力

- **参数校验层**：统一的参数校验装饰器/中间件，支持非空、类型、格式、范围校验
- **默认值注入层**：为可选参数提供统一的默认值处理
- **标准化错误响应**：统一错误码格式 `MODULE_ERRCODE`

### 修改能力

- **路径安全**：新增路径规范化与安全检查，禁止路径穿越
- **类型安全**：引入安全的类型断言辅助函数，禁止直接 `.()` 断言

### 非目标

- 不涉及 HTTP 传输层（请求头、Cookie、超时、重试）
- 不改变业务逻辑层（分享权限、容量限制等）
- 不修改日志/监控等横切关注点

## Capabilities

### New Capabilities

- `parameter-validation`: 统一的参数校验层，包括：
  - 非空检查、类型校验、格式校验、范围校验
  - 路径规范化与安全检查
  - 安全的类型断言辅助函数
  - 标准化错误响应（统一错误码格式）

- `default-value-injection`: 可选参数默认值注入层

### Modified Capabilities

(None - no existing spec-level behavior changes)

## Impact

### 影响的文件

| 文件 | 修改内容 |
|------|----------|
| `sdk/validation/` (new) | 新增统一参数校验层 |
| `sdk/file.go` | 修复路径穿越漏洞，添加参数校验调用 |
| `sdk/share.go` | 修复随机数安全问题，添加分页参数校验 |
| `cmd/main.go` | 修复类型断言 panic，添加 CLI 参数校验 |
| `sdk/types.go` | 新增安全的类型断言辅助函数 |

### API 变更

- 保持向后兼容，现有 API 签名不变
- 在方法注释中明确参数要求
- 返回值保持 `StandardResponse` 或明确的错误类型

### 测试要求

- 每个参数校验逻辑需配套单元测试
- 参数校验覆盖率 ≥ 80%
