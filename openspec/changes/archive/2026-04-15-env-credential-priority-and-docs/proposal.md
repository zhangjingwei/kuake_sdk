# Why

开发者与 CI 常以环境变量注入敏感凭证（如 `KUAKE_COOKIE`），期望**运行环境**中的配置优先于本地文件或单次命令行参数，避免仓库内 `config.json` 或脚本里的 `-cookies` 在无意间覆盖已注入的 Secret。当前实现为 `-cookies` 优先于环境变量，与这一预期不一致。

同时，文档与 CHANGELOG 中提到的部分环境变量（例如 `KUAKE_UPLOAD_PARALLEL`、`KUAKE_PATH`）与仓库内 Go 代码的实际读取行为**未对齐**，易误导集成方与排障。

# What Changes

- **BREAKING**：调整 `kuake` 认证凭证来源的**优先级**为：`KUAKE_COOKIE`（非空）优先于 `-cookies` / `--cookies`，再优先于配置文件中的 `access_tokens`（与当前顺序相反处需明确写入 spec 与 `docs/cli.md` / `docs/CHANGELOG.md`）。
- 梳理并**统一文档**：在 `docs/cli.md`（及必要时 `README.md`、`docs/CHANGELOG.md`）中列出与本仓库相关的环境变量，区分「CLI/SDK 运行时读取」「仅测试」「仅写入/内部传递」「文档曾提及待核实」等类别，并与实现一致。
- 对 **`KUAKE_UPLOAD_PARALLEL`**：实现与文档二选一或同时收敛——要么在并行上传路径中**读取**该环境变量（并与 `--max_upload_parallel` 的优先级在 design 中写清），要么从文档中删除/改写为「仅由 CLI  flag 在进程内设置」等真实行为。
- 对 **`KUAKE_PATH`**：在代码中实现约定、或从本仓库文档中删除/迁移说明至 OpenClaw 宿主侧文档，避免「文档承诺、二进制未实现」。

# Capabilities

## New Capabilities

- `cli-environment-config`：规范 `kuake`/SDK 与凭证相关的环境变量行为（含凭证来源优先级、与文档一致性要求），以及与本变更相关的其它已文档化环境变量的真实语义。

## Modified Capabilities

- （无）——现有 `openspec/specs/` 下能力（如 `parameter-validation`）的对外需求不因本变更必须修改；若后续实现触碰共享校验逻辑，可在实现阶段再评估是否追加 delta spec。

# Impact

- **代码**：`cmd/main.go`（凭证分支顺序）；可能涉及 `sdk/` 中与上传并行度、调试开关相关的读取逻辑；若有 `KUAKE_PATH` 结论则涉及 CLI 入口或帮助文案。
- **文档**：`docs/cli.md`、`docs/CHANGELOG.md`、`README.md`；必要时 `openclaw/kuake_skill/SKILL.md` 中与 PATH/环境变量一致的表述。
- **行为**：升级后依赖「`-cookies` 覆盖环境变量」的脚本将表现为环境变量优先，属破坏性变更，需在 CHANGELOG 中醒目标注。
