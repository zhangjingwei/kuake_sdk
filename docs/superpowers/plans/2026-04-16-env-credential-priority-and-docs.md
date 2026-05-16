# env-credential-priority-and-docs Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将凭证解析顺序改为 `KUAKE_COOKIE`（trim 后非空）优先于 `-cookies`，再优先于配置文件；按 `openspec/changes/env-credential-priority-and-docs/design.md` 采用 **方案 A** 实现 `KUAKE_UPLOAD_PARALLEL` 且 **`--max_upload_parallel` 高于环境变量**；修正 `KUAKE_PATH` 等文档与实现不一致；补齐 `docs/cli.md` 环境变量参考表与 **BREAKING** CHANGELOG。

**Architecture:** 在 `cmd` 包内新增可单测的纯函数（Cookie 规范化、上传并行度解析），`main` 仅编排优先级；不上改 `sdk.NewQuarkClient` 签名。`KUAKE_PATH` 按 design **不实现 Getenv**，仅从 README/CHANGELOG/技能文档去掉对 `kuake` 二进制「读取该变量」的误述，改为强调 **PATH** 与宿主配置。

**Tech Stack:** Go 1.18+、`go test`、`openspec validate`、仓库路径 `d:\workspace\kuake_cli`（执行者请替换为本地路径）。

---

## File map

| 文件 | 动作 |
|------|------|
| `cmd/cookie_normalize.go` | **新建**：`normalizeQuarkCookieInput(raw string) string`，集中 `__pus=` 与末尾分号逻辑 |
| `cmd/cookie_normalize_test.go` | **新建**：表驱动单测 |
| `cmd/upload_parallel_env.go` | **新建**：`resolveUploadParallelForProcess(flagValue string) string`（flag 非空优先，否则合法 env 1–16，否则 `""`） |
| `cmd/upload_parallel_env_test.go` | **新建**：单测（含 `t.Setenv`） |
| `cmd/main.go` | 修改客户端初始化分支顺序（约 L121–157）、`printUsage` 中英文说明（约 L133、L235–313）、`handleUpload` 在 `Setenv` 前合并 env（约 L479–546） |
| `docs/cli.md` | 新增「环境变量参考」、更新 Cookie 优先级与 `-cookies`/配置文件表述 |
| `docs/CHANGELOG.md` | **BREAKING**、并行度、`KUAKE_PATH` 修正摘要 |
| `README.md` | 去掉 `KUAKE_PATH` 与 `kuake` 强绑定表述，保留 OpenClaw + `KUAKE_COOKIE` |
| `openclaw/kuake_skill/SKILL.md` | 补充「将 `kuake` 安装目录加入 PATH」类说明（若仍要求 PATH 中的 `kuake`） |
| `openspec/changes/env-credential-priority-and-docs/tasks.md` | 全部完成后勾选（可选） |
| `sdk/e2e_regression_test.go` | 仅当文件头注释出现旧认证顺序时更新（当前无则跳过） |

---

### Task 1: Cookie 规范化纯函数 + 单测

**Files:**
- Create: `cmd/cookie_normalize.go`
- Create: `cmd/cookie_normalize_test.go`

- [ ] **Step 1: 新建 `cmd/cookie_normalize.go`（完整内容）**

```go
package main

import "strings"

// normalizeQuarkCookieInput 对「非空原始凭证串」应用与历史 CLI 一致的规则：trim 后若仍为空则返回 ""；
// 否则若无 "__pus=" 子串则加前缀 "__pus="；若末尾无 ";" 则追加 ";"。
// 供 KUAKE_COOKIE 与 -cookies 两路共用（openspec：两路等价规范化）。
func normalizeQuarkCookieInput(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if !strings.Contains(s, "__pus=") {
		s = "__pus=" + s
	}
	if !strings.HasSuffix(s, ";") {
		s = s + ";"
	}
	return s
}
```

- [ ] **Step 2: 新建 `cmd/cookie_normalize_test.go`（完整内容）**

```go
package main

import "testing"

func TestNormalizeQuarkCookieInput(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"whitespace_only", "   \t", ""},
		{"add_prefix_and_semicolon", "abc", "__pus=abc;"},
		{"preserve_pus_add_semi", "__pus=val", "__pus=val;"},
		{"trim_then_normalize", "  token  ", "__pus=token;"},
		{"already_complete", "__pus=x;y=1;", "__pus=x;y=1;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeQuarkCookieInput(tt.in); got != tt.want {
				t.Fatalf("normalizeQuarkCookieInput(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 3: 运行测试（应 PASS）**

```powershell
Set-Location d:\workspace\kuake_cli
go test ./cmd -run TestNormalizeQuarkCookieInput -count=1 -v
```

期望：全部 `PASS`。

- [ ] **Step 4: 提交**

```bash
git add cmd/cookie_normalize.go cmd/cookie_normalize_test.go
git commit -m "refactor(cmd): extract Quark cookie normalization for shared use"
```

---

### Task 2: 上传并行度 env 解析（flag > env）+ 单测

**Files:**
- Create: `cmd/upload_parallel_env.go`
- Create: `cmd/upload_parallel_env_test.go`

- [ ] **Step 1: 新建 `cmd/upload_parallel_env.go`（完整内容）**

```go
package main

import (
	"os"
	"strconv"
	"strings"
)

const uploadParallelEnvMax = 16

// resolveUploadParallelForProcess 返回应写入 KUAKE_UPLOAD_PARALLEL 的十进制字符串，或 "" 表示不设置。
// 优先级（design）：命令行 flag 已解析出的值 > 环境变量 KUAKE_UPLOAD_PARALLEL（1–16）；非法 env 视为未设置。
func resolveUploadParallelForProcess(flagValue string) string {
	if strings.TrimSpace(flagValue) != "" {
		return strings.TrimSpace(flagValue)
	}
	s := strings.TrimSpace(os.Getenv("KUAKE_UPLOAD_PARALLEL"))
	if s == "" {
		return ""
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > uploadParallelEnvMax {
		return ""
	}
	return strconv.Itoa(n)
}
```

- [ ] **Step 2: 新建 `cmd/upload_parallel_env_test.go`（完整内容）**

```go
package main

import (
	"testing"
)

func TestResolveUploadParallelForProcess(t *testing.T) {
	t.Run("flag_wins_over_env", func(t *testing.T) {
		t.Setenv("KUAKE_UPLOAD_PARALLEL", "2")
		if got := resolveUploadParallelForProcess("8"); got != "8" {
			t.Fatalf("got %q, want 8", got)
		}
	})
	t.Run("env_when_no_flag", func(t *testing.T) {
		t.Setenv("KUAKE_UPLOAD_PARALLEL", "3")
		if got := resolveUploadParallelForProcess(""); got != "3" {
			t.Fatalf("got %q, want 3", got)
		}
	})
	t.Run("invalid_env_ignored", func(t *testing.T) {
		t.Setenv("KUAKE_UPLOAD_PARALLEL", "99")
		if got := resolveUploadParallelForProcess(""); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})
}
```

- [ ] **Step 3: 运行测试**

```powershell
Set-Location d:\workspace\kuake_cli
go test ./cmd -run TestResolveUploadParallelForProcess -count=1 -v
```

期望：`PASS`。

- [ ] **Step 4: 提交**

```bash
git add cmd/upload_parallel_env.go cmd/upload_parallel_env_test.go
git commit -m "feat(cmd): resolve KUAKE_UPLOAD_PARALLEL when flag unset (1-16)"
```

---

### Task 3: `main.go` 凭证优先级 + 调用规范化函数

**Files:**
- Modify: `cmd/main.go`（客户端创建块约 L121–157；解析循环内可为 `cookies` 赋值的行保留，不在循环内 normalize）

- [ ] **Step 1: 将 `// 优先级：cookies 参数 > ...` 整段替换为以下逻辑（保持 `defer` 与 `client` 声明不变）**

在 `// 创建客户端` 与 `defer func()` 之后，用下列片段**替换**原 L133–157（注释与 `if/else` 整体）：

```go
	// 优先级：环境变量 KUAKE_COOKIE（trim 后非空）> -cookies/--cookies > 配置文件（BREAKING，见 CHANGELOG）
	envRaw := os.Getenv("KUAKE_COOKIE")
	if norm := normalizeQuarkCookieInput(envRaw); norm != "" {
		client = sdk.NewQuarkClient(configPath, norm)
	} else if cookies != "" {
		cookies = normalizeQuarkCookieInput(cookies)
		if cookies == "" {
			client = sdk.NewQuarkClient(configPath)
		} else {
			client = sdk.NewQuarkClient(configPath, cookies)
		}
	} else {
		client = sdk.NewQuarkClient(configPath)
	}
```

说明：`cookies` 来自 CLI，可能含仅空格；`normalizeQuarkCookieInput` 返回 `""` 时回退配置文件，避免把 `__pus=;` 一类异常传给 SDK（与「空则不用 -cookies」语义一致）。

- [ ] **Step 2: 编译**

```powershell
Set-Location d:\workspace\kuake_cli
go build -o kuake.exe ./cmd
```

期望：无编译错误。

- [ ] **Step 3: 手工矩阵（可选但推荐）**

准备两个不同 cookie 前缀的**假**串（勿用真账号时可仅用长度不同的占位），在同一 shell：

1. `$env:KUAKE_COOKIE="A"`；`.\kuake.exe -cookies "B" user` → 期望走 **A**（若 API 返回错误，至少确认请求侧或日志不指向 B；有集成配置时可看服务端返回）。
2. `Remove-Item Env:KUAKE_COOKIE`；`.\kuake.exe -cookies "你的有效片段" user` → 期望走 **-cookies**。

- [ ] **Step 4: 提交**

```bash
git add cmd/main.go
git commit -m "fix(cmd)!: prefer KUAKE_COOKIE over -cookies for auth source"
```

---

### Task 4: `handleUpload` 接入 `resolveUploadParallelForProcess`

**Files:**
- Modify: `cmd/main.go`（`handleUpload` 内 `if uploadParallel != "" { _ = os.Setenv(...) }` 一段，约 L544–546）

- [ ] **Step 1: 将**

```go
	if uploadParallel != "" {
		_ = os.Setenv("KUAKE_UPLOAD_PARALLEL", uploadParallel)
	}
```

**替换为**

```go
	if v := resolveUploadParallelForProcess(uploadParallel); v != "" {
		_ = os.Setenv("KUAKE_UPLOAD_PARALLEL", v)
	}
```

（变量名 `uploadParallel` 仍为 flag 解析结果字符串；无 flag 时为空串，由 `resolveUploadParallelForProcess` 读 env。）

- [ ] **Step 2: 运行 `go test ./cmd -count=1`**

期望：含 Task 1–2 的测试全部 PASS。

- [ ] **Step 3: 提交**

```bash
git add cmd/main.go
git commit -m "feat(cmd): apply KUAKE_UPLOAD_PARALLEL when upload flag omitted"
```

---

### Task 5: 更新 `printUsage` 与 `main.go` 内嵌帮助文案

**Files:**
- Modify: `cmd/main.go`（`printUsage` 中涉及 cookies / parallel 的英文段落）

- [ ] **Step 1: 全文搜索 `printUsage` 内子串 `cookies`、`KUAKE_COOKIE`、`parallel`，将认证顺序改为 `KUAKE_COOKIE` > `-cookies` > config；将并行度一行改为明确 `--max_upload_parallel` overrides `KUAKE_UPLOAD_PARALLEL`。**

示例替换句（按你文件实际上下文粘贴，以下为英文帮助片段范例）：

```text
Auth source precedence: KUAKE_COOKIE env (non-empty after trim) > -cookies/--cookies > config file access_tokens.
Upload parallel: --max_upload_parallel N overrides KUAKE_UPLOAD_PARALLEL (1-16); if neither set, default applies.
```

- [ ] **Step 2: `go build ./cmd`**

- [ ] **Step 3: 提交**

```bash
git add cmd/main.go
git commit -m "docs(cli): sync embedded usage with env auth and upload parallel"
```

---

### Task 6: `docs/cli.md` — 环境变量参考 + Cookie / 上传说明

**Files:**
- Modify: `docs/cli.md`（「选项」或「配置说明」附近插入新节；修正 Cookie 与 `-cookies` 段落）

- [ ] **Step 1: 在「### 基本用法」或「**选项**」之后插入新小节（完整 Markdown 块，可按排版微调）**

```markdown
### 环境变量参考

| 变量名 | 谁读取 | 用途 | 说明 |
|--------|--------|------|------|
| `KUAKE_COOKIE` | `kuake`（cmd） | 整段会话 Cookie | **优先于**拆分变量、`-cookies`、配置文件（trim 并规范化后非空则生效） |
| `KUAKE_PUS` / `KUAKE_PUUS` | `kuake`（cmd） | `__pus` / `__puus` 裸值 | 仅当 `KUAKE_COOKIE` 无效时拼接后规范化；见 `docs/cli.md` |
| `-cookies` / `--cookies` | `kuake`（cmd） | 同上 | 当 `ResolveEnvCookieString()` 为空时使用；规范化规则与 env 一致 |
| `KUAKE_UPLOAD_PARALLEL` | `kuake`（cmd，`upload`） | 上传并行度 1–16 | 未传 `--max_upload_parallel` 时读取；**命令行 flag 优先于本变量** |
| `KUake_DEBUG` | SDK（`QuarkClient`） | 调试输出 | 设为 `1` 开启；变量名大小写以代码为准 |
| `E2E_REGRESSION` / `INTEGRATION_TEST` | `go test ./sdk` | 启用 `TestE2E_Regression_CoreFlow` | 凭证仅 `KUAKE_COOKIE` 或 `KUAKE_PUS`+`KUAKE_PUUS`；**不**读取 `KUAKE_E2E_CONFIG` / `config.json` |

集成方（如 OpenClaw）需保证 **`kuake` 在 `PATH` 中**；`kuake` **不**读取 `KUAKE_PATH` 环境变量。
```

- [ ] **Step 2: 将文中仍写「`-cookies` 参数 > 环境变量」或「使用 `-cookies` 时不会读取配置文件」且与**新优先级**冲突的句子改为：当使用**生效的** Cookie 来源为 `-cookies` 且未走配置文件 token 时，不加载配置文件中的 access_tokens；并写明 **若设置了 `KUAKE_COOKIE` 则以其为准**。

- [ ] **Step 3: 提交**

```bash
git add docs/cli.md
git commit -m "docs(cli): env reference table and auth precedence BREAKING"
```

---

### Task 7: `README.md` 与 `openclaw/kuake_skill/SKILL.md` 去掉误述 `KUAKE_PATH`

**Files:**
- Modify: `README.md`（约 L26）
- Modify: `openclaw/kuake_skill/SKILL.md`（Prerequisites）

- [ ] **Step 1: 将 `README.md` 中**  
`KUAKE_COOKIE` / `KUAKE_PATH`**  
改为类似：**

```markdown
- JSON 输出、管道模式（与 `jq` 等组合）；可选 **OpenClaw** 技能（见 [openclaw/](openclaw/)）；集成时可使用环境变量 `KUAKE_COOKIE` 注入凭证，并确保 `kuake` 已加入 **`PATH`**
```

- [ ] **Step 2: 在 `openclaw/kuake_skill/SKILL.md` 的 `## Prerequisites` 下追加一条要点（中文或英文与全文一致即可）**

示例：

```markdown
- On hosts where `kuake` is not on PATH, install it to a directory that is on PATH (this repository's `kuake` binary does not read `KUAKE_PATH`).
```

- [ ] **Step 3: 提交**

```bash
git add README.md openclaw/kuake_skill/SKILL.md
git commit -m "docs: remove KUAKE_PATH claim; clarify PATH for kuake binary"
```

---

### Task 8: `docs/CHANGELOG.md` — BREAKING + 并行度 + PATH

**Files:**
- Modify: `docs/CHANGELOG.md`（在 `[Unreleased]` 或下一版本小节顶部追加；并修正历史条目中错误描述如「`-cookies` > `KUAKE_COOKIE`」「支持 KUAKE_PATH」若作为**事实陈述**仍误导，可加脚注或改为过去时说明已更正）

- [ ] **Step 1: 追加一条 **BREAKING**（示例文本，版本号按团队规范改）**

```markdown
### BREAKING

- **认证凭证来源优先级**调整为：`KUAKE_COOKIE`（trim 后非空）优先于 `-cookies` / `--cookies`，再优先于配置文件。曾依赖「命令行覆盖已 export 的 `KUAKE_COOKIE`」的脚本须先清除环境变量（POSIX: `unset KUAKE_COOKIE`；PowerShell: `Remove-Item Env:KUAKE_COOKIE`）或改用配置文件。
- **上传**：未传 `--max_upload_parallel` 时，`kuake` 会读取 `KUAKE_UPLOAD_PARALLEL`（1–16）；传入 flag 时 **flag 优先**。
- **文档**：已移除「`kuake` 二进制通过 `KUAKE_PATH` 解析路径」的表述；请通过系统 **PATH** 或包装脚本定位 `kuake`。
```

- [ ] **Step 2: 若 CHANGELOG 中旧版本仍写「认证优先级：`-cookies` > …」且会误导读者，在该行末追加简短说明「（已于 x.x 更正，见 BREAKING）」或等价表述。**

- [ ] **Step 3: 提交**

```bash
git add docs/CHANGELOG.md
git commit -m "docs(changelog): BREAKING auth order, upload env, PATH clarification"
```

---

### Task 9: 全仓库字符串审计（`KUAKE_PATH`、旧优先级）

**Files:**
- Read-only：`rg` 结果驱动修改

- [ ] **Step 1: 运行**

```powershell
Set-Location d:\workspace\kuake_cli
rg "KUAKE_PATH" -g"*.{md,go,txt,yml,yaml}"
rg "cookies.*KUAKE_COOKIE|KUAKE_COOKIE.*cookies|优先级" -g"*.md"
```

- [ ] **Step 2: 对仍声称「`-cookies` 优先于 `KUAKE_COOKIE`」或「`kuake` 读取 `KUAKE_PATH`」的 Markdown（含 `docs\CHANGELOG.md` 重复路径若存在、openspec archive 下历史文档）**逐条**改为与实现一致，或明确标注为历史错误已修复。**

（`openspec/changes/env-credential-priority-and-docs/proposal.md` 可保留「当前实现为旧顺序」叙述，无需改成将来时，除非团队要求。）

- [ ] **Step 3: 提交（可能为空或多项）**

```bash
git add -A
git commit -m "docs: align remaining mentions of KUAKE_PATH and auth precedence"
```

---

### Task 10: 全量构建、测试与 OpenSpec 校验

**Files:**
- 全仓库

- [ ] **Step 1: 构建**

```powershell
Set-Location d:\workspace\kuake_cli
go build ./...
```

期望：退出码 0。

- [ ] **Step 2: 测试（含 race 可选）**

```powershell
go test ./... -count=1
```

若 `INTEGRATION_TEST` / E2E 类测试在无密钥环境失败：使用仓库既有约定跳过（不引入新失败）；至少 `go test ./cmd ./sdk -count=1` 在默认环境下应通过（集成测试已 `t.Skip` 则忽略）。

- [ ] **Step 3: OpenSpec**

```powershell
openspec validate env-credential-priority-and-docs
```

期望：`Change 'env-credential-priority-and-docs' is valid`。

- [ ] **Step 4: 勾选 `openspec/changes/env-credential-priority-and-docs/tasks.md` 全部项并提交（可选单独 commit）**

```bash
git add openspec/changes/env-credential-priority-and-docs/tasks.md
git commit -m "chore(openspec): complete env-credential-priority-and-docs tasks"
```

- [ ] **Step 5: 归档（按团队流程，可能另开 PR）**

```powershell
openspec archive env-credential-priority-and-docs
```

若 `archive` 会合并 spec 到 `openspec/specs/`，执行前确认无未提交改动。

---

## Self-review（对照 `specs/cli-environment-config/spec.md`）

| Spec 要点 | 覆盖 Task |
|-----------|-----------|
| 凭证顺序：env > flag > 配置 | Task 3 |
| 两路 Cookie 规范化一致 | Task 1 + Task 3 |
| `docs/cli.md` 环境变量表与真实 `Getenv` 一致 | Task 6 + Task 9 |
| `KUAKE_UPLOAD_PARALLEL` 与文档一致（方案 A） | Task 2 + Task 4 + Task 6 + Task 8 |
| CHANGELOG **BREAKING** | Task 8 |
| `KUAKE_PATH` 不误称二进制支持 | Task 7 + Task 8 + Task 9 |

**Placeholder 扫描：** 本计划未使用「TBD / 稍后实现」类占位；并行度上界常量 `16` 与 `design.md` 及现有 `docs/cli.md` 描述一致。

**一致性：** `normalizeQuarkCookieInput` 与 `resolveUploadParallelForProcess` 命名在全文统一；`uploadParallelEnvMax` 与文档「1–16」一致。

---

**Plan complete and saved to `docs/superpowers/plans/2026-04-16-env-credential-priority-and-docs.md`. Two execution options:**

**1. Subagent-Driven（推荐）** — 每个 Task 派生子代理，任务间人工或自动复核，迭代快。

**2. Inline Execution** — 本会话按 Task 顺序执行，配合 executing-plans 的检查点。

**Which approach?**

（说明：`writing-plans` 技能建议专用 worktree；若你已在 `main` 上工作，执行前可自行 `git worktree add` 隔离，非本文件强制步骤。）
