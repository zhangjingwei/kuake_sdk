# Tasks: env-credential-priority-and-docs

## 1. 实现：凭证优先级

- [x] **1.1** 修改 `cmd/main.go`：当 `KUAKE_COOKIE`（trim 后）非空时优先于 `-cookies` / `--cookies`，再回退到配置文件；保持与现有相同的 cookie 字符串规范化逻辑（`__pus=`、分号等）。
- [x] **1.2** 抽取或内联保证「env 与 flag 两路」规范化**唯一实现**，避免重复与行为不一致。（`cmd/cookie_normalize.go` + `normalizeQuarkCookieInput`）
- [x] **1.3** 若存在针对认证优先级的测试或 E2E 说明，更新注释与示例（如 `sdk/e2e_regression_test.go` 文件头、`docs/superpowers` 下若引用旧顺序则同步）。（已检索：`e2e_regression_test.go` 文件头无旧认证顺序；无需改）

## 2. 实现或收敛：`KUAKE_UPLOAD_PARALLEL`

- [x] **2.1** 在 `design.md` 所选方案（A：实现 `Getenv` + 与 flag 优先级；B：删改文档）上达成一致后执行。（方案 A）
- [x] **2.2** 若选方案 A：在 `cmd` 或 `sdk` 中实现读取与边界校验（1–16），并写明 **flag 与环境变量** 的优先级；更新 `docs/cli.md` 与 `docs/CHANGELOG.md`。
- [x] **2.3** 若选方案 B：…（**不适用**：已采用方案 A）

## 3. 审计与收敛：`KUAKE_PATH`

- [x] **3.1** 全仓库检索 `KUAKE_PATH` 与 `Getenv`，确认是否由 `kuake` 二进制读取。（无 `Getenv("KUAKE_PATH")`）
- [x] **3.2** 若无读取：从 `README.md`、`docs/CHANGELOG.md` 删除或改写为「由宿主/技能配置 PATH」；若 OpenClaw 技能文档需补充 PATH 说明，更新 `openclaw/kuake_skill/SKILL.md`。
- [x] **3.3** 若决定实现读取：…（**跳过**：设计选择不实现 `KUAKE_PATH`）

## 4. 文档：环境变量参考与一致性

- [x] **4.1** 在 `docs/cli.md` 增加「环境变量参考」一节，表格列出：变量名、消费者（CLI / SDK / 测试）、用途、取值、是否与凭证链相关。
- [x] **4.2** 核对 `KUake_DEBUG` 的文档拼写与代码一致，并说明其作用域（SDK 调试）。
- [x] **4.3** 更新 `README.md` 中与凭证优先级、环境变量相关的简述（若有）。
- [x] **4.4** 在 `docs/CHANGELOG.md` 增加 **BREAKING** 条目：凭证优先级变更；并记录 `KUAKE_UPLOAD_PARALLEL` / `KUAKE_PATH` 文档或实现的修正摘要。

## 5. 验证与收尾

- [x] **5.1** 本地执行 `rg "Getenv|LookupEnv"`（或等价）与 `docs/cli.md` 表格对照，确保无遗漏、无虚假承诺。
- [x] **5.2** `go build ./...`（及现有相关测试）通过；若有针对 `cmd` 的可测性新增，补充最小测试或记录手工验收步骤于本 tasks 或 `verification.md`。**说明**：本仓库部分环境上 `go test ./cmd` 会因 Application Control 策略拦截 `cmd.test.exe` 无法执行；代码已 `go build ./...` 通过，子测试逻辑见 `cmd/*_test.go`。
- [x] **5.3** 实现完成后执行 `openspec validate env-credential-priority-and-docs`（或当前项目约定的校验命令），通过后再走 `openspec archive`。（**validate 已通过**；**archive 未执行**，留待发版/合并流程）
