## 1. 校验层基础设施

- [ ] 1.1 创建 `sdk/validation/` 目录结构
- [ ] 1.2 实现 `sdk/validation/validator.go` - 核心校验器接口和 Chain 函数
- [ ] 1.3 实现 `sdk/validation/errors.go` - 统一错误类型和错误码定义
- [ ] 1.4 实现 `sdk/validation/checkers.go` - 基础校验器（NonEmpty, InRange, ValidPath）

## 2. 安全类型断言

- [ ] 2.1 实现 `sdk/validation/typeassert.go` - 安全类型断言辅助函数
  - [ ] 2.1.1 SafeString - 安全字符串转换
  - [ ] 2.1.2 SafeFloat64 - 安全浮点数转换
  - [ ] 2.1.3 SafeInt - 安全整数转换
  - [ ] 2.1.4 SafeMap - 安全 map 转换

## 3. 分页参数校验

- [ ] 3.1 实现 `sdk/validation/pagination.go` - 分页参数校验
- [ ] 3.2 编写单元测试 `sdk/validation/pagination_test.go`
  - [ ] 3.2.1 测试 page 范围 [1, 1000]
  - [ ] 3.2.2 测试 size 范围 [1, 100]
  - [ ] 3.2.3 测试默认值注入

## 4. 默认值注入层

- [ ] 4.1 实现 `sdk/validation/default.go` - 默认值注入配置
- [ ] 4.2 定义 `PageDefaults` - 分页参数默认值配置
- [ ] 4.3 编写单元测试 `sdk/validation/default_test.go`

## 5. 高危漏洞修复

- [ ] 5.1 **VULN-001**: 修复 `sdk/file.go` 路径穿越漏洞
  - [ ] 5.1.1 在 `UploadFile` 方法中添加路径安全校验
  - [ ] 5.1.2 在 `DownloadFile` 方法中添加路径安全校验
  - [ ] 5.1.3 编写路径穿越测试用例
- [ ] 5.2 **VULN-002**: 修复 `sdk/*.go` 类型断言 panic
  - [ ] 5.2.1 在 `sdk/quark_client.go` 中使用 SafeTypeAssert
  - [ ] 5.2.2 在 `sdk/file.go` 中使用 SafeTypeAssert
  - [ ] 5.2.3 在 `sdk/share.go` 中使用 SafeTypeAssert
- [ ] 5.3 **VULN-003**: 修复 `cmd/main.go` CLI 参数类型断言 panic
  - [ ] 5.3.1 实现 `cmd/validation/args.go` 参数解析辅助函数
  - [ ] 5.3.2 在 `handleShareCreate` 中使用安全的参数解析
  - [ ] 5.3.3 在其他 CLI 命令中应用相同的参数解析模式
- [ ] 5.4 **VULN-004**: 修复 `sdk/share.go` 可预测随机数
  - [ ] 5.4.1 在 `sdk/validation/random.go` 中实现安全随机数生成
  - [ ] 5.4.2 替换 `GetShareStoken` 中的 math/rand 为 crypto/rand
  - [ ] 5.4.3 在其他使用随机数的地方应用相同的修复
- [ ] 5.5 **VULN-005**: 修复 `sdk/share.go` 分页参数未限制
  - [ ] 5.5.1 在 `GetMyShareList` 中添加分页参数校验
  - [ ] 5.5.2 在其他分页方法中添加相同的校验
- [ ] 5.6 **VULN-006**: 修复 `cmd/main.go` 空指针/越界访问
  - [ ] 5.6.1 实现 `cmd/validation/json.go` 安全的 JSON 解析
  - [ ] 5.6.2 在 `extractPathFromJSON` 中使用安全的解析函数

## 6. 文档与验收

- [ ] 6.1 更新 API 文档 - 在方法注释中添加参数说明
- [ ] 6.2 运行 `go test -cover` 验证参数校验覆盖率 ≥ 80%
- [ ] 6.3 运行 `go vet` 检查潜在问题
- [ ] 6.4 运行 `golangci-lint` 检查代码质量
- [ ] 6.5 更新 CHANGELOG.md 记录本次重构

## 7. 规范差距与待办（Code Review 补充）

以下条目对照 `specs/parameter-validation/spec.md` 与 `specs/default-value-injection/spec.md`，记录当前实现与规范仍不一致或可改进之处。

### 7.1 规范 API 与错误码补全

- [ ] 7.1.1 实现 `sdk/validation/checkers.go` 中规范列出但尚未实现的校验器：`InRangeFloat64`、`ValidEmail`、`ValidFID`（或更新规范删除未采用项）
- [ ] 7.1.2 在 `sdk/validation/errors.go` 中补全规范错误码：`FILE_INVALID_PATH`、`USER_INVALID_INPUT`，并在对应校验失败路径使用
- [ ] 7.1.3 复核 `ValidPath`：规范要求通过安全检查时**返回规范化路径**；当前仅返回 `error`，需实现返回规范化路径的 API（或修订规范与实现保持一致）

### 7.2 校验器语义强化

- [ ] 7.2.1 `NonEmpty`：明确对非字符串类型是「仅校验 string/nil」还是「拒绝未知类型」；若规范要求后者，补充分支与测试
- [x] 7.2.2 `InRange`：对非 `int` 类型是静默通过还是返回 `ErrInvalidArgument`；按规范定稿并实现（方案 A：仅 `int` 参与区间；`float64` 与其它类型均 `ErrInvalidArgument`，见 `specs/parameter-validation/spec.md`）

### 7.3 全量 SDK 公共方法与响应形态

- [ ] 7.3.1 按规范「所有 SDK 公共方法的外部输入必须经过校验层」梳理 `sdk/*.go` 公开方法，分批接入校验（路径、分页、业务 ID 等）
- [ ] 7.3.2 统一校验失败对外形态：规范要求 `StandardResponse` 且 `Success: false`；当前 `GetMyShareList` 等返回 `error`（如 `*ValidationError`），需统一为 `StandardResponse` 或修订规范明确例外 API 列表
- [ ] 7.3.3 将仍使用裸类型断言的响应解析路径改为 `validation.Safe*` 或安全分支（含 `sdk/file.go` 中 `UploadFile` 对 `params["policy"].(string)` 等）

### 7.4 分页与默认值在其他 API 上对齐

- [ ] 7.4.1 在 `GetShareList`（及规范/tasks 涉及的其它对外分页方法）中应用 `PageDefaults.Apply` + `PaginateParams.Validate()`，与 `GetMyShareList` 行为一致
- [ ] 7.4.2 复核内部列表循环（如 `listByFid` 固定 page/size）是否在文档中说明与对外分页规范的关系

### 7.5 默认值 `Apply` 与两份 spec 的表述一致

- [ ] 7.5.1 `default-value-injection` 写的是「仅当键不存在时填充」；`parameter-validation` 写的是「nil 或零值」算未提供。当前实现对零值也会注入默认值。择一为准并**同步修订另一份 spec**或**在 spec 中显式约定零值语义**

### 7.6 错误消息语言

- [ ] 7.6.1 规范要求用户可见错误为中文；将 `ErrInvalidArgument`、`ErrPathTraversal` 等对外文案改为中文，或区分「日志英文 / 用户 Message 中文」两层字段

### 7.7 文档与验收（补充）

- [ ] 7.7.1 在 proposal/design 或 spec 中记录「未实现 API / 与规范偏差」清单，避免与 openspec 验收口径冲突

---

## 8. TDD 微型任务拆解（执行用）

> 已按 **writing-plans** 技能将上文各条拆成 **2–5 分钟一步** 的红绿循环；详细步骤、文件边界、默认 `go test` 命令与 §7 逐条对应关系见：  
> **[`micro-tasks-tdd.md`](./micro-tasks-tdd.md)**

### 8.1 结构说明

| 区块 | 内容 |
|------|------|
| **§A** | 对应 `tasks.md` §1–6：以「先跑测试 → 缺则补失败测试 → 实现至绿」为主 |
| **§B** | 对应 `tasks.md` §7：规范差距，按错误码、校验器、`GetShareList`/响应形态、spec 同步、中文文案等拆分 |
| **§C** | 每个批次固定节奏（红 → 绿 → 重构 → 提交）可复制 |
| **§D** | 与 `specs/*.md` 对照自检表 |

### 8.2 建议执行顺序

1. 完成 **§A**（确认基线全绿，缺测补测）。  
2. 按 **§B.1 → B.14** 顺序做规范补全（可并行 B.12 文档与 B.2–B.4 校验器，但 **B.5** 与 **B.9** 依赖 `ValidPath`/policy 决策）。  
3. **§B.15** 为 Epic，在 **§B.8 / B.10** 定稿错误形态后再铺开。

### 8.3 微型任务编号前缀

- **A.x.x**：验证 / 补测（§1–6）  
- **B.x.x**：§7 规范差距  
- 完整列表与勾选请维护在 **`micro-tasks-tdd.md`**，避免本文件重复两份 checklist。
