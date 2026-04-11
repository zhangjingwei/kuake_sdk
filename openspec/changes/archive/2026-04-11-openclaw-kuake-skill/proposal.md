schema: spec-driven

# Why

本项目 README 和文档已提到“OpenClaw 技能集成”，但当前仓库中没有标准的 OpenClaw skill 包目录。为了让 OpenClaw 智能体能够直接使用 kuake CLI 功能，需基于 OpenClaw skill 标准格式创建一个 `kuake_skill`。

这个 change 旨在明确：
- 采用 OpenClaw skill 的标准目录与 `SKILL.md` 结构
- 将 `kuake` CLI 能力封装为可被 OpenClaw agent 调用的技能
- 补齐项目文档，让开发者和用户知道该 skill 的安装、加载和测试方式

# What Changes

- 新增 OpenClaw skill 包：`openclaw/kuake_skill/`，包含标准 `SKILL.md`
- 在 `README.md` 中补充该 skill 包的路径与使用说明
- 在 `docs/CHANGELOG.md` 中记录新增 OpenClaw skill 支持

# Capabilities

### New Capabilities
- `kuake_skill` : 提供 OpenClaw agent 访问 `kuake` CLI 的入口，用于执行与夸克网盘相关的命令（如列目录、上传、下载、分享等）

### Modified Capabilities
- `documentation` : 补充仓库 README 与 changelog 中的 OpenClaw skill 支持说明

# Impact

Affected files:
- 新增：`openclaw/kuake_skill/SKILL.md`
- 修改：`README.md`
- 修改：`docs/CHANGELOG.md`

Affected behavior:
- 仅增加项目文档与 agent skill 元数据，不直接改变现有运行时逻辑
- 为未来 OpenClaw agent 集成 `kuake` CLI 提供明确入口和说明
