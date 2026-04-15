# cli-environment-config Specification

## Purpose

约定 `kuake` CLI 与 SDK 侧**凭证来源顺序**、**环境变量与文档一致**，以及上传并行度环境变量等行为，便于开发者与 CI 集成。

## Requirements
### Requirement: 凭证来源优先级

`kuake` SHALL 按固定顺序解析会话凭证：非空的 `KUAKE_COOKIE` 优先于命令行 `-cookies` / `--cookies`，再优先于配置文件中的 `Quark.access_tokens`。

#### Scenario: 环境变量覆盖命令行与配置文件

- **WHEN** 进程环境中 `KUAKE_COOKIE` 经 trim 后非空，且用户同时传入了非空的 `-cookies`，且配置文件中存在 `access_tokens`
- **THEN** 客户端 MUST 使用 `KUAKE_COOKIE` 的值作为认证凭证

#### Scenario: 无环境变量时使用命令行

- **WHEN** `KUAKE_COOKIE` 未设置或 trim 后为空，且用户传入了非空的 `-cookies`
- **THEN** 客户端 MUST 使用命令行传入的 cookie 值

#### Scenario: 回退到配置文件

- **WHEN** `KUAKE_COOKIE` 为空且未传入 `-cookies`
- **THEN** 客户端 MUST 从配置文件加载 `access_tokens`，并满足「至少一个 token」的既有约束

### Requirement: Cookie 字符串规范化一致性

从 `KUAKE_COOKIE` 与从 `-cookies` 获得的字符串 SHALL 经过相同的规范化规则（例如 `__pus=` 前缀与末尾分号处理），再传入客户端构造逻辑。

#### Scenario: 两路输入行为一致

- **WHEN** 同一规范化后字符串分别通过环境变量与通过 `-cookies` 传入
- **THEN** 客户端产生的认证请求 MUST 等价（不因来源不同而采用不同解析规则）

### Requirement: 用户文档中的环境变量参考

`docs/cli.md` SHALL 包含「环境变量参考」（或等价标题），列出本仓库 Go 代码实际读取的变量、含义、取值约定及适用范围（CLI、SDK、`go test`），且与实现一致。

#### Scenario: 文档与代码对照

- **WHEN** 维护者对照 `docs/cli.md` 环境变量表与仓库内 `Getenv` / `LookupEnv` 使用处
- **THEN** 表中每一条「由 kuake/SDK 读取」的变量 MUST 在代码中存在对应读取；表中 MUST 标明仅测试使用的变量

### Requirement: 并行上传环境变量与实现一致

关于 `KUAKE_UPLOAD_PARALLEL`，对外文档（`docs/cli.md`、`docs/CHANGELOG.md`）SHALL 与实现一致：要么实现用户级 `Getenv` 并文档化与 `--max_upload_parallel` 的优先级，要么移除「用户可自行 export 即可生效」等不实表述。

#### Scenario: 无虚假承诺

- **WHEN** 用户仅设置环境变量 `KUAKE_UPLOAD_PARALLEL` 而未传 `--max_upload_parallel`
- **THEN** 实际并行行为 MUST 与文档描述一致（若文档称生效则代码须读取；若文档称不生效则代码不得误导）

### Requirement: 破坏性变更的记录

`docs/CHANGELOG.md` SHALL 以 **BREAKING** 标注凭证优先级变更，并给出迁移建议（例如曾依赖 `-cookies` 覆盖环境变量的用户应先清除 `KUAKE_COOKIE`）。

#### Scenario: 升级可见性

- **WHEN** 用户阅读新版本 CHANGELOG
- **THEN** 能够识别认证优先级为破坏性变更并获知推荐迁移步骤

### Requirement: `KUAKE_PATH` 文档与二进制行为一致

若 `kuake` 二进制不读取 `KUAKE_PATH`，则 `README.md` 与 `docs/CHANGELOG.md` SHALL 不得将该变量表述为 CLI 自身支持的配置；若实现读取，则 SHALL 在 `docs/cli.md` 中说明语义与优先级。

#### Scenario: 无未实现承诺

- **WHEN** 仓库内不存在对 `KUAKE_PATH` 的 `Getenv` 调用
- **THEN** 面向用户的文档 MUST NOT 声称仅通过设置 `KUAKE_PATH` 即可由 `kuake` 解析可执行路径（除非改为实现该行为并更新文档）

