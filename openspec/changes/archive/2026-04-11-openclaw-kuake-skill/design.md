# Design

## 目标

基于 OpenClaw skill 标准格式，创建一个独立 skill 目录，使 OpenClaw agent 可以识别并加载 `kuake` CLI 能力。

## 结构

使用仓库内统一的 `openclaw/` 根目录：
- `openclaw/kuake_skill/`
  - `SKILL.md`
  - 可选文档或引用文件

OpenClaw 官方建议 skill 目录根包含一个 `SKILL.md`，该文件由 YAML frontmatter 和 Markdown 说明组成。

## `SKILL.md` 内容

`SKILL.md` 需要包含：
- `name`: `kuake_skill`
- `description`: 简要说明 skill 能做什么
- 可选 `metadata.openclaw.requires.bins`: `['kuake']`，强调该 skill 依赖本地 `kuake` 可执行文件

示例前言：
- 让智能体在用户请求与夸克网盘交互时，优先使用 `exec` 工具调用本地 `kuake` 命令
- 如果 `kuake` 未安装或不可用，则说明如何从仓库构建/安装

Markdown 说明应包含：
- 触发条件：用户询问与夸克网盘、文件上传、下载、列目录、分享等操作相关的任务
- 目标行为：将用户意图转换成 `kuake` CLI 命令，并使用 `exec` 工具执行
- 安全边界：禁止直接执行未经验证的任意 shell 命令；仅调用 `kuake` 并传递已验证参数
- 示例命令模式：如 `kuake list "/"`、`kuake upload "file.txt" "/file.txt"`

## 文档更新

- `README.md` 增加 `openclaw/kuake_skill/` 的说明，说明该目录是 OpenClaw skill 包，用户可将项目工作区作为 OpenClaw workspace 加载
- `docs/CHANGELOG.md` 添加新增 OpenClaw skill 支持的简要描述

## 测试与验证

验证方案：
1. 在本地构建或安装 `kuake` 可执行文件
2. 将仓库工作区作为 OpenClaw workspace 或将 `openclaw/kuake_skill/` 添加到 `skills.load.extraDirs`
3. 重启 OpenClaw gateway 或开启新会话
4. 通过 `openclaw skills list` 确认 `kuake_skill` 已加载
5. 发送测试消息，如“列出我的夸克网盘根目录”，验证 agent 使用 `kuake` 命令

## 兼容性

如果用户未在 PATH 中安装 `kuake`，skill 需要指导 agent 提示用户先安装或构建本地 `kuake`。这样可以避免 skill 在不可用环境中失败。
