# Design: env-credential-priority-and-docs

## 1. 凭证优先级（目标顺序）

**（归档说明）** 本节为变更当时草案；当前实现已在 `KUAKE_COOKIE` 与 `-cookies` 之间加入 **`KUAKE_PUS` / `KUAKE_PUUS`** 拼接路径，并以仓库内 **`openspec/specs/cli-environment-config/spec.md`** 为单一事实来源。

```
KUAKE_COOKIE（trim 后非空）
        │
        ▼ 否
 -cookies / --cookies（非空）
        │
        ▼ 否
 配置文件 Quark.access_tokens
```

### 1.1 实现要点（`cmd/main.go`）

- 将当前 `if cookies != "" { ... } else if envCookie ... else { config }` 调整为**先判断 `KUAKE_COOKIE`**，再判断 CLI `cookies` 变量，最后走 `NewQuarkClient(configPath)`。
- 两处来源在传入 `NewQuarkClient` 之前复用**同一套**规范化逻辑（避免复制粘贴分叉；若已有函数则抽取为局部函数或包内小函数）。

### 1.2 迁移与兼容性

- **破坏性**：曾依赖「命令行 `-cookies` 覆盖已 export 的 `KUAKE_COOKIE`」的用法将失效。迁移方式：临时调试时先 `unset KUAKE_COOKIE`（POSIX）或 `$env:KUAKE_COOKIE=$null`（PowerShell），或改用仅配置文件。
- 不在本变更中引入「第三优先级开关」或兼容旧顺序的环境变量（保持简单；若未来需要再在 spec 中扩展）。

## 2. 环境变量清单（文档侧结构建议）

在 `docs/cli.md` 建议用表格列至少：

| 变量 | 读取方 | 用途 | 备注 |
|------|--------|------|------|
| `KUAKE_COOKIE` | `cmd` | 会话凭证 | 优先级见上 |
| `KUake_DEBUG` | `sdk`（QuarkClient） | 调试 | 名称与 `KUAKE_*` 不一致，文档写清 |
| `KUAKE_UPLOAD_PARALLEL` | 待本 change 结论 | 上传并行度 | 与实现同步后填入 |
| `E2E_REGRESSION` / `INTEGRATION_TEST` | `go test ./sdk` | 启用 E2E 回归 | 凭证为 `KUAKE_COOKIE` 或 `KUAKE_PUS`+`KUAKE_PUUS`；**不**使用 `KUAKE_E2E_CONFIG`（见 `docs/cli.md`） |

## 3. `KUAKE_UPLOAD_PARALLEL` 决策分支

**方案 A（推荐若希望保留文档承诺）**：在 `cmd` 解析 upload 子命令时，若未传 `--max_upload_parallel`，则 `strconv.Atoi(os.Getenv("KUAKE_UPLOAD_PARALLEL"))`，校验范围 1–16，合法则与 flag 等效地传入 SDK（或通过 `Setenv` + SDK `Getenv` 二选一，**避免**仅 `Setenv` 而无读取的半套实现）。

**方案 B**：删除/改写 `docs/cli.md`、`docs/CHANGELOG.md` 中「用户可 export 该变量」的表述，仅保留 flag 与（若保留）「CLI 内部 Setenv」实现细节不出现在用户文档中。

**优先级（若采用方案 A）**：建议 `--max_upload_parallel` **高于**环境变量（单次命令显式覆盖部署默认值），并在 spec/doc 中写明；若产品坚持「环境变量绝对最大」，则与 REQ-CLI-ENV-001 的哲学一致，但需单独写一条需求避免与凭证规则混淆。

本 design 默认 **方案 A + flag > env** 仅针对上传并行度（与凭证链解耦），除非任务执行时团队另有决定。

## 4. `KUAKE_PATH`

- **核实**：全仓库 `Getenv("KUAKE_PATH")` 若不存在，则要么删除 README/CHANGELOG 中「二进制支持 `KUAKE_PATH`」的表述，要么在 `cmd` 中实现「当 argv[0] 为 `kuake` 且在 PATH 中找不到时，用 `KUAKE_PATH` 指定可执行文件」等明确语义（实现成本与 OpenClaw 集成方式需对齐）。
- 本 change 的 tasks 中采用「先审计再二选一」任务项。

## 5. 验证

- 单元或轻量集成：构造 `KUAKE_COOKIE` 与 `-cookies` 同时存在时，断言使用 env（mock HTTP 或仅测 `NewQuarkClient` 入参需评估可测性；最低限度为对 `main` 包提取纯函数做表测，若当前结构难以测则任务中注明「以手工矩阵验收」）。
- 文档：人工对照 `docs/cli.md` 表格与 `rg Getenv` 结果。
