# 变更日志

## v1.4.2

- **SDK 路径与列表**
  - 修复 `listByFid` 翻页：仅以本页条数是否满页决定是否继续，避免依赖不可靠的 `total` 导致列表缺项
  - Windows 下远程路径统一按 POSIX 处理：`GetFileInfo` / `UploadFile` 使用 `path.Base` 解析远端路径；`GetFileInfo` 列表回退分支使用 `path.Join` 拼接子路径
  - `listByFid` 兼容 JSON 将 `fid` 解析为 `float64` 的情况
- **测试与记录**
  - 新增可选端到端回归 `TestE2E_Regression_CoreFlow`（`E2E_REGRESSION=1` 或 `INTEGRATION_TEST=1`，配置见 `KUAKE_E2E_CONFIG`）
  - 问题与回归说明见仓库根目录 `buglist.txt`

## v1.4.1

- 新增主规格架构文档：`specs/architecture/spec.md`
- 补充项目目录和模块划分说明
- 更新 `.gitignore`，排除本地辅助配置但保留 `.github/workflows/`
- 归档 OpenSpec 变更：`openspec/changes/archive/2026-04-11-main-architecture`
- 新增 OpenClaw skill 包支持：添加 `openclaw/kuake_skill/SKILL.md`，提供标准 OpenClaw skill 格式以便 agent 集成 kuake CLI 能力

## v1.4.0

- **OpenClaw 技能集成**
  - 新增 kuake OpenClaw 技能支持
  - 添加环境变量 `KUAKE_COOKIE` 支持，符合 OpenClaw 标准配置方式
  - 认证优先级：`-cookies` 参数 > 环境变量 `KUAKE_COOKIE` > 配置文件
  - 支持通过 `KUAKE_PATH` 环境变量指定完整路径，不依赖 PATH 检测
  - 优化 OpenClaw 技能文档，添加 fallback 逻辑说明
  - 简化部署文档，提供更清晰的配置选项

## v1.3.9

- 新增 `--policy` 上传去重策略（PR #16）
  - 新增 `UploadPolicy`/`UploadOptions` 类型定义
  - 支持三种策略：`skip`（跳过已存在文件）、`rename`（重命名）、`overwrite`（覆盖）
  - `UploadFile` 函数签名扩展为 4 参数，支持策略配置
- 并行上传优化（PR #18，8 项核心改动）
  - 嵌入式哈希：MD5+SHA1 嵌入分片读取，提高上传效率
  - `parallel_upload` 握手协议：优化并行上传协商流程
  - `X-Oss-Hash-Ctx` MarshalBinary 修复：修复序列化问题
  - Nl/Nh 32 位拆分：支持 >536MB 大文件上传
  - 多线程并发上传：提升上传速度至 7-14 MB/s
  - 分片级指数退避重试：每个分片最多重试 3 次，提高成功率
  - 断点续传 PartThread 恢复：支持中断后恢复并行上传状态
  - `x-oss-user-agent` 版本统一：统一版本标识
- `user` 命令容量查询 + `--version`（PR #17）
  - `getMemberInfo()` 合并容量/会员信息：统一用户信息获取接口
  - 版本号常量定义：规范化版本管理
  - `--version` 参数拦截：新增版本号查询命令参数
- 新增管道模式（Pipeline Pattern）支持，支持命令链式组
  - `list` 命令新增 `--stream` 选项，输出流式 JSON（每行一个文件对象）
  - `delete`、`info`、`download` 命令支持从 stdin 读取 JSON 输入
  - 自动检测 stdin，有数据时自动进入管道模式
  - 支持与其他 Unix 工具（如 `jq`、`grep`、`head` 等）组合使用
  - 保持向后兼容，无 stdin 时使用命令行参数
  - 实现流式处理，支持逐行处理大量文件，内存占用低
  - 改进错误处理，优雅处理 broken pipe 错误
  - 统一数据类型处理，只处理 `QuarkFileInfo` 类型，提高代码一致性和可维护性
  - 优化代码结构，移除多类型处理的复杂逻辑，简化代码实现

## v1.3.8

- 新增 `-cookies` 参数支持，可直接通过命令行指定 cookie 值，无需配置文件
  - 自动为 cookie 值添加 `__pus=` 前缀（如果缺失）
  - 自动添加末尾分号（如果缺失）
  - 使用 `-cookies` 参数时，不会读取配置文件，提高效率并避免不一致
- 修复并行上传逻辑，多分片文件禁用并行上传（因为需要使用 X-Oss-Hash-Ctx）

## v1.3.7

- 新增并行上传功能，支持通过 `--max_upload_parallel` 参数或 `KUAKE_UPLOAD_PARALLEL` 环境变量配置并行度（1-16，默认 4）
- 改进路径参数处理，明确要求所有路径参数必须用引号包裹
- 新增转存分享文件功能，新增 `share-save` CLI 命令

## v1.3.6

- 新增 X-Oss-Hash-Ctx 支持，实现 OSS 分片上传的增量 SHA1 哈希上下文
- 改进断点续传功能，支持 HashCtx 的保存和恢复

## v1.3.5

- 新增断点续传功能，上传中断后可自动恢复
- 改进上传进度显示，显示上传速度、剩余时间等信息
- 优化命令行参数解析，支持 `-c/--config` 参数指定配置文件路径
- 增强上传错误处理和超时处理
- 改进分享创建错误处理，增加重试机制

## v1.3.4

- 修复配置文件读取路径问题，支持从可执行文件所在目录读取配置文件

## v1.3.3

- 修复 Windows 路径处理问题，支持跨平台路径兼容性

## v1.3.2

- 新增取消分享功能，新增 `share-delete` CLI 命令

## v1.3.1

- 修复 CLI 错误消息转义问题
- 优化 API 错误响应处理
- 新增完整的单元测试套件
