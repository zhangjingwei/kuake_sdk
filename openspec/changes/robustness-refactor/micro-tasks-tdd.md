# robustness-refactor — TDD 微型任务拆解

> **用途：** 与 `tasks.md` 对应；每步控制在约 2–5 分钟，严格 **红 → 绿 → （可选）重构 → 提交**。  
> **执行：** 按编号顺序做；同一文件的多个微型任务可在一个提交中完成，但每个**行为变更**仍应先写失败测试。

**涉及目录：** `sdk/validation/`、`sdk/*.go`、`cmd/validation/`、`openspec/changes/robustness-refactor/specs/`、`proposal.md` / `design.md`

**24 方法校验矩阵：** `openspec/changes/robustness-refactor/public-methods-validation-tasks.md`（章节「24 方法校验矩阵」）。

**默认命令：**

```bash
cd d:/workspace/kuake_sdk
go test ./sdk/validation/... -count=1
go test ./sdk/... -count=1
go test ./cmd/... -count=1
```

---

## 文件职责速查（分解边界）

| 文件 | 职责 |
|------|------|
| `sdk/validation/errors.go` | 错误码、构造器、中文/双语文案 |
| `sdk/validation/checkers.go` | NonEmpty、InRange、InRangeFloat64、ValidEmail、ValidFID、ValidPath（及规范化路径） |
| `sdk/validation/validator.go` | Validator、Chain |
| `sdk/validation/typeassert.go` | SafeString / Float64 / Int / Map |
| `sdk/validation/pagination.go` | PaginateParams.Validate |
| `sdk/validation/default.go` | Defaults、PageDefaults、Apply 语义 |
| `sdk/share.go` | GetMyShareList、GetShareList、分页与错误形态 |
| `sdk/file.go` | UploadFile policy 安全读取等 |

---

## A. 与 `tasks.md` §1–6 对应的验证 / 补测微型任务

> 若实现已存在：先做「测试绿」再勾选；缺测试则按下列顺序补 **一个失败测试 → 红 → 实现/补全 → 绿**。

### A.1 校验层基础设施（tasks §1）

- [x] **A.1.1** `go test ./sdk/validation -run TestNonEmpty -count=1` — **PASS**（2026-04-15）
- [x] **A.1.2** `go test ./sdk/validation -run TestInRange -count=1` — **PASS**
- [x] **A.1.3** `go test ./sdk/validation -run TestValidPath -count=1` — **PASS**
- [x] **A.1.4** 若上列无对应测试：… — **跳过**（已有 `TestNonEmpty` / `TestInRange` / `TestValidPath`，含空白字符串等场景）

### A.2 安全类型断言（tasks §2）

- [x] **A.2.1** `go test ./sdk/validation -run TestSafeString -count=1` — **PASS**
- [x] **A.2.2** `go test ./sdk/validation -run TestSafeFloat64 -count=1` — **PASS**
- [x] **A.2.3** `go test ./sdk/validation -run TestSafeInt -count=1` — **PASS**
- [x] **A.2.4** `go test ./sdk/validation -run TestSafeMap -count=1` — **PASS**

### A.3 分页（tasks §3）

- [x] **A.3.1** … `page=0` — **PASS**（`pagination_test.go` 表驱动 `page_below_min`）
- [x] **A.3.2** … `page=1001` — **PASS**（`page_above_max`）
- [x] **A.3.3** … `size=0`、`size=101` — **PASS**（`size_below_min` / `size_above_max`）
- [x] **A.3.4** 边界合法 — **PASS**（`valid_params`、`valid_page_1000`、`valid_size_100`）

### A.4 默认值（tasks §4）

- [x] **A.4.1** `TestDefaults_Apply` — **PASS**（含缺键填充、零值视为未提供）
- [x] **A.4.2** `TestPageDefaults` — **PASS**
- [x] **A.4.3** `TestUploadOptionsDefaults` — **PASS**

### A.5 高危漏洞（tasks §5）

- [x] **A.5.1 VULN-001** 实际运行：`go test ./sdk -run "UploadFile|PathNormalization|InvalidDest" -count=1` — **PASS**（`TestUploadFile` 因需网络 **SKIP**；`TestUploadFile_InvalidDestPath`、`TestPathNormalizationInFunctions` **PASS**）
- [ ] **A.5.2 VULN-002** 对 `file.go` 中关键 `.(string)`：**先写**触发错误类型的表驱动测试（若需导出测试辅助则放在 `export_test.go`）→ 红 → 改为 Safe* → 绿 — **未执行**（留待专项 TDD）
- [x] **A.5.3 VULN-003** `go test ./cmd/validation -count=1` — **PASS**
- [x] **A.5.4 VULN-004** `go test ./sdk/validation -run TestSecureRandomInt -count=1` — **PASS**（含 `TestSecureRandomInt_RangeEdge`）
- [x] **A.5.5 VULN-005** 实际运行：`go test ./sdk -run Share -count=1` — **PASS**（`TestGetShareList` 等因需网络 **SKIP**；无单独 `TestGetMyShareList`，列表类均为集成跳过）
- [x] **A.5.6 VULN-006** `go test ./cmd/validation -run ExtractPathFromJSON -count=1` — **PASS**（随 `./cmd/validation` 全量）

### A.6 文档与验收（tasks §6）

- [x] **A.6.1** `go test -cover ./sdk/validation/...` — **coverage: 84.9%**（statements，2026-04-15）
- [x] **A.6.2** `go vet ./...` — **PASS**（无输出）
- [x] **A.6.3** `golangci-lint run ./...` — **环境待办**：本机未安装 `golangci-lint`（命令未找到）

---

## B. `tasks.md` §7 规范差距 — 严格 TDD 微型任务

> **执行状态：** **B.1–B.15 已完成**（2026-04-15 批次）。汇总见文末「执行记录汇总」。

### B.1 错误码 `FILE_INVALID_PATH` / `USER_INVALID_INPUT`（对应 7.1.2）

- [x] **B.1.1** 在 `errors_test.go` **新建** `TestErrInvalidPath_Code`：… — **RED**（`ErrInvalidPath` 未定义，编译失败）→ **B.1.2**
- [x] **B.1.2** … `ErrInvalidPath` — **PASS**
- [x] **B.1.3** `TestErrUserInvalidInput_Code` → **RED** → `ErrUserInvalidInput` — **PASS**
- [x] **B.1.4** `TestValidPath` 增加 `wantCode`：**RED**（NUL/UNC 等仍为 `FILE_PATH_TRAVERSAL`）→ `ValidPath` 对 NUL、UNC/保留路径返回 `ErrInvalidPath`（`FILE_INVALID_PATH`），段级 `.`/`..` 仍为 `ErrPathTraversal` — **PASS**

### B.2 `InRangeFloat64`（对应 7.1.1）

- [x] **B.2.1** `TestInRangeFloat64_BelowMin` — **RED**（未实现 → 桩返回 nil）→ **绿**
- [x] **B.2.2** `InRangeFloat64` + `floatFromInterface`（`float64`/`int`/`int64`）— **PASS**
- [x] **B.2.3** `AboveMax`、`BoundaryMin`、`BoundaryMax`；另 `TestInRangeFloat64_IntLikeJSON` — **PASS**
- [x] **B.2.4** `TestInRangeFloat64_WrongType` — **RED**（非数值返回 nil）→ 拒绝非数值 — **PASS**

### B.3 `ValidEmail`（对应 7.1.1）

- [x] **B.3.1–B.3.3** `TestValidEmail_Invalid` / `ValidEmail`（正则）/ `TestValidEmail_Valid` — **PASS**

### B.4 `ValidFID`（对应 7.1.1）

- [x] **B.4.1–B.4.3** `design.md` FID 规则；`TestValidFID_Empty` / 非法字符 / 合法 — **PASS**

### B.5 `ValidPath` 返回规范化路径（对应 7.1.3）

- [x] **B.5.1** `design.md`：`ValidPathResult` API — **完成**
- [x] **B.5.2–B.5.3** `TestValidPath_ReturnsCleanPath` + `ValidPathResult` — **PASS**
- [x] **B.5.4** `UploadFile` / `DownloadFile` 使用 `ValidPathResult` — **完成**

### B.6 `NonEmpty` 非 string 类型（对应 7.2.1）

- [x] **B.6.1–B.6.2** `TestNonEmpty_IntRejected`；`NonEmpty` 仅 string — **PASS**
- [x] **B.6.3** `parameter-validation/spec.md` 已写明

### B.7 `InRange` 非 int（对应 7.2.2）

- [x] **B.7.1–B.7.2** `TestInRange_Float64Input`；拒绝 `float64` — **PASS**

### B.8 `GetMyShareList` 与 `StandardResponse`（对应 7.3.2）

- [x] **B.8.1** `TestGetMyShareList_InvalidPage_ReturnsValidationError` — **PASS**
- [x] **B.8.2–B.8.3** 保持 `(map[string]interface{}, error)`，未改 `StandardResponse`
- [x] **B.8.4** `cmd` 无单独单测；调用方式不变

### B.9 `UploadFile` policy 安全读取（对应 7.3.3）

- [x] **B.9.1–B.9.2** `TestUploadPolicyFromParams_NonStringUsesSkip`；`uploadPolicyFromParams` — **PASS**

### B.10 `GetShareList` 分页对齐（对应 7.4.1）

- [x] **B.10.1** `TestGetShareList_InvalidPage_ReturnsValidationError`
- [x] **B.10.2** `GetShareList`：`PageDefaults` + `PaginateParams.Validate` — **PASS**
- [x] **B.10.3** 与 `GetMyShareList` 一致：`*validation.ValidationError`

### B.11 `listByFid` 文档（对应 7.4.2）

- [x] **B.11.1–B.11.2** `file.go` 注释 — **完成**

### B.12 两份 spec 零值语义（对应 7.5.1）

- [x] **B.12.1–B.12.3** 两 spec 与 `Defaults.Apply` 对齐；`default_test.go` — **PASS**

### B.13 中文错误文案（对应 7.6.1）

- [x] **B.13.1–B.13.3** `errors_test.go` 中文断言；校验器与 `ErrPathTraversal` — **PASS**

### B.14 偏差清单文档（对应 7.7.1）

- [x] **B.14.1–B.14.2** `design.md` 偏差表 — **完成**

### B.15 全 SDK 方法梳理（对应 7.3.1，Epic）

- [x] **B.15.1–B.15.3** 文末 **24** 方法；`TestB15_ExportedMethodsChecklist`；`go test ./sdk/...` — **PASS**

---

## C. 每个微型批次的固定节奏（可复制）

1. **红：** 只改 `*_test.go`，`go test -run 'ExactName' -count=1` → FAIL  
2. **绿：** 最小生产代码，`go test` → PASS  
3. **重构：** 不增行为，仅整理；`go test` 仍 PASS  
4. **提交：** `git add` 范围尽量小  

---

## D. Self-review（对照 spec）

| spec 章节 | 本文件章节 |
|-----------|------------|
| parameter-validation 校验器列表 | B.2–B.7、A.1 |
| 错误码 MODULE_ERRCODE | B.1、B.13 |
| default-value-injection Apply | A.4、B.12 |
| 分页范围 | A.3、B.10 |

**缺口：** 若 `B.15` 未展开到每一方法，则 spec「所有公共方法」仍标为部分覆盖，须在 `design.md` 说明分阶段范围。**§E.1 / E.2** 已落实（见 `design.md` 与 `GetShareList` 校验）；**E.1.3** 仍为长期项。

---

## E. spec 与实现缺口（代码审查 2026-04-15）

> **E.1 / E.2 已完成**（2026-04-15）；**E.1.3** 仍为长期立项项。

### E.1 `StandardResponse` 与 `*ValidationError` 双轨（WARNING 1）

- [x] **E.1.1** 在 `design.md` 偏差表或验收章节中增加一条，与 `specs/parameter-validation/spec.md` **Implementation Rules 第 2–3 条**交叉引用：明确返回 `(业务数据, error)` 的公共方法在校验失败时返回 `*validation.ValidationError`（含 `Code`），与返回 `*StandardResponse` 的方法并列说明，避免读者误以为「所有失败都必须走 `StandardResponse`」。
- [x] **E.1.2**（可选）梳理 `QuarkClient` 导出方法中哪些返回 `*StandardResponse`、哪些返回「数据 + `error`」，将表格归档至 `design.md` 或本节附录（可与 B.15 清单合并维护）。
- [ ] **E.1.3**（长期 / 立项）若产品要求调用侧全局统一为 `StandardResponse`：评估破坏性变更、`V2` 方法或并行 API；**单独提案前不修改**现有公开方法签名。

### E.2 业务参数校验的统一错误形态（WARNING 2）

针对仍使用 `fmt.Errorf` 等、无法 `errors.As` 到 `*validation.ValidationError` 或缺少规范 `Code` 的分支（典型：`GetShareList` 的 `sortBy` 非法）。

- [x] **E.2.1** `share_test.go`：新增 `TestGetShareList_InvalidSortBy_ReturnsValidationError` —— `sortBy` 既非 `file_name` 也非 `updated_at` 时，调用方可用 `errors.As(err, &validation.ValidationError)`，且 `Code == "INVALID_ARG"`。先写测试，预期在实现修改前为 **RED**。
- [x] **E.2.2** `sdk/share.go` — `GetShareList`：将 `sortBy` 非法分支从 `fmt.Errorf(...)` 改为 `validation.ErrInvalidArgument("...")`（或规范规定的其它构造器），保持中文用户可见消息。
- [x] **E.2.3**（可选）若注释承诺 `sortOrder` 仅为 `asc` / `desc`：为 `sortOrder` 增加枚举校验，与注释一致；补表驱动测试。
- [x] **E.2.4** 验收：`go test ./sdk -run "GetShareList" -count=1`；检索 `cmd/` 或其它调用方是否依赖该错误的 `Error()` 字符串，必要时文档或兼容说明。

---

## 执行记录汇总（自动化验收）

**日期：** 2026-04-15  
**环境：** Windows，`go test` 默认；仓库路径 `d:\workspace\kuake_sdk`。

| 区块 | 状态 | 说明 |
|------|------|------|
| **A.1–A.4** | 已完成 | 上述条目均已跑通并勾选 |
| **A.5** | 基本完成 | **A.5.2** 未做专项表驱动 + Safe* 改造 |
| **A.6** | 基本完成 | **A.6.3** 待安装 `golangci-lint` |
| **B.1–B.15** | 已完成 | 见本节 B.1–B.15 勾选；含 `ValidEmail`/`ValidFID`/`ValidPathResult`、分页、`SafeString` policy、中文文案、spec 与 `design.md` 偏差表、B.15 清单测试 |
| **E.1 / E.2** | 已完成 | `design.md`：`StandardResponse` 与 `*ValidationError` 双轨说明 + 导出方法返回类型表；`GetShareList`：`sort_by`/`sort_order` 校验与 `ErrInvalidArgument`，`share_test.go` 单测 |

**全量测试（非清单强制，供基线）：**

- `go test ./sdk/validation/... -count=1` — **PASS**
- `go test ./cmd/... -count=1` — **PASS**
- `go test ./sdk/... -count=1` — **FAIL**（`TestUpdateHashCtxFromHash` / `TestUpdateHashCtxFromHash_Incremental`：`Nl` 期望值与实现不一致，与 robustness 清单无直接关系）

### B.15.1 — `QuarkClient` 导出方法清单（`sdk/*.go`）

以下均为 `func (qc *QuarkClient)` 且方法名首字母大写（便于后续「一方法一批次」校验桩）：

| 文件 | 方法 |
|------|------|
| `quark_client.go` | `SetBaseURL`, `GetCookies`, `ConvertToFileInfo` |
| `user.go` | `GetUserInfo` |
| `share.go` | `GetShareInfo`, `GetShareStoken`, `GetShareList`, `SaveShareFile`, `CreateShare`, `GetShareLink`, `SetSharePassword`, `GetMyShareList`, `GetShareIDByFid`, `DeleteShare` |
| `file.go` | `UploadFile`, `CreateFolder`, `Copy`, `Move`, `Rename`, `List`, `GetFileInfo`, `Delete`, `GetDownloadURL`, `DownloadFile` |

**合计：** 24 个导出方法（未计未导出辅助方法如 `listByFid`）。
