# Buglist 端到端回归收口 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 Windows 与 WSL 上各跑通 `TestE2E_Regression_CoreFlow`，并按设计更新 `buglist.txt`（及可选 CHANGELOG），完成 buglist 回归验证记录。

**Architecture:** 无 SDK 行为变更；执行顺序为先本机 Windows 验证、再在 WSL 中挂载同一仓库路径验证；两次均通过后一次性编辑 `buglist.txt` 写入「回归验证记录」并视结果更新 ISSUE-006；CHANGELOG 仅在发版或团队要求时追加。

**Tech Stack:** Go `go test`、PowerShell（Windows）、Bash（WSL）、仓库内 `config.json` / `KUAKE_E2E_CONFIG`。

---

## File map

| 文件 | 动作 |
|------|------|
| `buglist.txt` | 新增「回归验证记录」小节；按需更新 ISSUE-006 状态；可选在 BUG-001～005 末行补充「于回归验证记录中确认」 |
| `docs/CHANGELOG.md` | 可选：仅当发版或明确要求时增加一条 |
| `README.md` / `docs/cli.md` | 可选：仅当 E2E 启用说明与现状不一致时微调 |
| `sdk/e2e_regression_test.go` | 只读参考，不修改（除非稳定复现 bug 需另开变更） |

---

### Task 1: Windows 上执行 E2E 并记录输出

**Files:**
- Read: `sdk/e2e_regression_test.go`（确认开关与 config 解析）
- Modify: 无（本步仅执行与记录）

- [ ] **Step 1: 进入仓库根目录并确认配置可读**

在 **PowerShell** 中（将路径换成你的实际仓库路径）：

```powershell
Set-Location D:\workspace\kuake_sdk
Test-Path .\config.json
# 期望: True；若使用自定义路径，改为 Test-Path $env:KUAKE_E2E_CONFIG 的目标文件
```

- [ ] **Step 2: 设置环境变量并运行测试**

使用与有效登录一致的配置（二选一）：

**A. 使用仓库根目录 `config.json`（测试会自动解析）：**

```powershell
$env:E2E_REGRESSION = "1"
go test ./sdk -run TestE2E_Regression_CoreFlow -count=1 -v
```

**B. 显式指定配置文件：**

```powershell
$env:E2E_REGRESSION = "1"
$env:KUAKE_E2E_CONFIG = "D:\workspace\kuake_sdk\config.json"
go test ./sdk -run TestE2E_Regression_CoreFlow -count=1 -v
```

- [ ] **Step 3: 记录元数据供 buglist 使用**

```powershell
go version
```

将以下内容抄到临时笔记（后续写入 `buglist.txt`）：

- 平台：`Windows`（可注明版本，如 `Windows 10/11`）
- `go version` 完整一行
- 测试结果：`PASS` 或 `FAIL`
- 若 FAIL：完整错误片段（尤其是否含 `403`、`Callback`、`DownloadFile`）

- [ ] **Step 4: 若失败则重试一次**

仅当 Step 2 为 FAIL 时，等待约 30 秒后重复 Step 2 **一次**。若第二次仍 FAIL，**不要**在此任务里改代码；在 Task 3 的备注中记录为「稳定失败」并保留日志。

- [ ] **Step 5: Commit**

本任务不产生代码变更；跳过 commit。若你仅记录了本地笔记，无需 git 操作。

---

### Task 2: WSL 上执行 E2E 并记录输出

**Files:**
- Modify: 无

- [ ] **Step 1: 在 WSL 中进入同一仓库**

假设 Windows 盘挂载为 `/mnt/d`：

```bash
cd /mnt/d/workspace/kuake_sdk
test -f config.json && echo OK
```

若 `config.json` 仅在 Windows 用户目录，请复制一份到仓库内 **或** 在 WSL 中设置 `KUAKE_E2E_CONFIG` 为 **WSL 路径下可读的文件**（内容应与 Windows 侧等价、同一账号）。

- [ ] **Step 2: 运行同一测试**

```bash
export E2E_REGRESSION=1
# 如需显式配置：
# export KUAKE_E2E_CONFIG=/mnt/d/workspace/kuake_sdk/config.json
go test ./sdk -run TestE2E_Regression_CoreFlow -count=1 -v
```

- [ ] **Step 3: 记录元数据**

```bash
go version
uname -a
```

笔记字段：平台 `WSL (Linux)`、`go version`、`uname -a` 一行、PASS/FAIL、失败时错误摘要。

- [ ] **Step 4: 失败时重试一次**

同 Task 1 Step 4。

- [ ] **Step 5: Commit**

无变更则跳过。

---

### Task 3: 更新 `buglist.txt`

**Files:**
- Modify: `buglist.txt`

- [ ] **Step 1: 在文件末尾「端到端测试说明」之前插入新小节**

在 `buglist.txt` 中找到行 `----------------------------------------------------------------------------` 且下一行为 `端到端测试说明` 的位置，**在其上方**插入下列模板（将占位符换成 Task 1/2 的真实记录）：

```text
----------------------------------------------------------------------------
回归验证记录（设计见 docs/superpowers/specs/2026-04-15-buglist-e2e-regression-design.md）
----------------------------------------------------------------------------
| 日期       | 平台   | Go 版本（可选）     | 结果 | 备注 |
|------------|--------|---------------------|------|------|
| YYYY-MM-DD | Windows | `go version 输出` | PASS |      |
| YYYY-MM-DD | WSL     | `go version 输出` | PASS |      |

说明：BUG-001～005 在上述两次回归中通过核心链路间接验证（列表、POSIX 路径、上传、GetFileInfo、下载）。
```

若任一侧为 FAIL，将对应行 `结果` 写 `FAIL`，`备注` 写简短原因（例如 `DownloadFile: 403 ...`）。

表格若不便对齐，可用固定宽度简化版，例如：

```text
- YYYY-MM-DD  Windows  …  PASS/FAIL  …  备注
- YYYY-MM-DD  WSL      …  PASS/FAIL  …  备注
```

- [ ] **Step 2: 更新 ISSUE-006**

- 若两侧均为 **PASS**（含下载）：将 ISSUE-006 的「状态」改为说明：**在有效登录下已于 YYYY-MM-DD 通过 Windows + WSL 核心链路验证（含 DownloadFile）**。
- 若任一侧 **稳定 FAIL** 且错误与下载/OSS 相关：更新「现象」为当前日志摘要，「状态」标明**有效登录下仍复现**，并保留「建议」中与跳过下载测试相关的说明供排查。

- [ ] **Step 3: 可选 — 在 BUG-001～005 状态行追加短语**

若希望强调已验证，可在各条「状态：已修复」行末追加：`（见回归验证记录）`。非必须，保持文件可读即可。

- [ ] **Step 4: 运行测试（非 E2E）确保未破坏构建**

```powershell
go test ./sdk -count=1 -short
```

若项目无 `-short` 惯例，可改为：

```powershell
go test ./sdk -count=1
```

期望：与改动前一致全部 PASS（`buglist.txt` 不影响编译；此步为习惯检查）。

- [ ] **Step 5: Commit**

```bash
git add buglist.txt
git commit -m "docs: record E2E regression verification on Windows and WSL"
```

---

### Task 4（可选）: `docs/CHANGELOG.md`

**Files:**
- Modify: `docs/CHANGELOG.md`

- [ ] **Step 1: 决策**

若 **近期不发版** 且 **无人要求** 在变更日志记录 QA：跳过整个 Task 4。

- [ ] **Step 2: 在下一版本小节下增加一条**

例如在 `## v1.4.2` 或 `## Unreleased` 下：

```markdown
- 端到端回归：Windows 与 WSL 下 `TestE2E_Regression_CoreFlow` 已通过（详见 `buglist.txt`）
```

- [ ] **Step 3: Commit**

```bash
git add docs/CHANGELOG.md
git commit -m "docs: changelog note for E2E regression verification"
```

---

### Task 5（可选）: 核对 `README.md` / `docs/cli.md` 与 E2E 说明一致

**Files:**
- Read: `README.md`, `docs/cli.md`
- Modify: 仅当发现与 `e2e_regression_test.go` 注释冲突时

- [ ] **Step 1: 对照检查**

打开 `sdk/e2e_regression_test.go` 文件头注释（约 L13–22），确认 README/cli 中关于 `E2E_REGRESSION`、`INTEGRATION_TEST`、`KUAKE_E2E_CONFIG` 的描述一致。

- [ ] **Step 2: 若有差异，做最小修改并 commit**

```bash
git add README.md docs/cli.md
git commit -m "docs: align E2E env instructions with e2e_regression_test.go"
```

若无差异，跳过。

---

## Spec coverage（自检）

| 设计章节 | 对应任务 |
|----------|----------|
| §1 目标与成功标准、失败重试 | Task 1–2 |
| §2 环境变量、命令、双平台 | Task 1–2 |
| §3.1 buglist 回归记录与 ISSUE-006 | Task 3 |
| §3.2 CHANGELOG 可选 | Task 4 |
| §3.3 其他文档可选 | Task 5 |
| §4 风险（不贸然改结论、配置一致） | Task 1–2 备注与 Task 3 ISSUE-006 |

## Placeholder scan

本计划无 TBD/TODO；日期与路径由执行者替换为实际值。

---

## Execution handoff

Plan complete and saved to `docs/superpowers/plans/2026-04-15-buglist-e2e-regression.md`. Two execution options:

**1. Subagent-Driven (recommended)** — Dispatch a fresh subagent per task, review between tasks, fast iteration.

**2. Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints.

Which approach?
