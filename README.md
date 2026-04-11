# Kuake CLI

[![License](https://img.shields.io/badge/License-AGPL--3.0-green.svg)](LICENSE)

夸克网盘文件管理 CLI 工具。

## 开源说明

本项目采用 **AGPL-3.0** 开源；**商业使用**（含 SaaS、商业产品集成等）须另行取得授权。源代码可自本仓库获取，使用 `build.sh` 可本地编译各平台二进制。欢迎通过 Issue / Pull Request 参与。

## 目录

- [功能特性](#功能特性)
- [系统要求](#系统要求)
- [安装](#安装)
- [快速开始](#快速开始)
- [文档与变更](#文档与变更)
- [免责声明](#免责声明)
- [许可证](#许可证)
- [贡献者](#贡献者)

## 功能特性

- 用户信息与网盘目录列表、文件详情、上传/下载、创建目录、移动/复制/重命名/删除
- 分享创建与取消、分享列表、转存他人分享
- JSON 输出、管道模式（与 `jq` 等组合）；可选 **OpenClaw** 技能（见 [openclaw/](openclaw/)，`KUAKE_COOKIE` / `KUAKE_PATH`）

更多用法见 [docs/cli.md](docs/cli.md)。

更多文档见：
- [specs/architecture/spec.md](specs/architecture/spec.md)
- [docs/agent-skill.md](docs/agent-skill.md)

## 系统要求

- Linux / macOS / Windows
- 有效的夸克网盘账号与 Cookie

## 安装

### 从源码构建

需要 Go 1.18+ 与 Git。

```bash
git clone https://github.com/zhangjingwei/kuake_sdk.git
cd kuake_sdk
chmod +x build.sh
./build.sh
```

构建产物位于 `dist/`。

### 预编译二进制

从 [Releases](https://github.com/zhangjingwei/kuake_sdk/releases) 下载对应平台文件，文件名与版本以 Release 页为准。

## 快速开始

1. 在项目目录创建 `config.json`：

```json
{
  "Quark": {
    "access_tokens": [
      "__pus=your_pus_value_here;"
    ]
  }
}
```

在浏览器登录夸克网盘，开发者工具（F12）→ Network 中复制请求的 Cookie，将完整 Cookie 字符串写入 `access_tokens`（支持多账号多条）。**勿将 `config.json` 提交到 Git。**

2. 运行（示例，请替换为实际二进制名或 `kuake`）：

```bash
./kuake user
./kuake list "/"
./kuake upload "file.txt" "/file.txt"
```

子命令表、`-c` / `-cookies`、管道模式与完整示例见 [docs/cli.md](docs/cli.md)。

## 文档与变更

| 文档 | 说明 |
|------|------|
| [docs/cli.md](docs/cli.md) | CLI 配置、命令表、JSON 约定与示例 |
| [docs/CHANGELOG.md](docs/CHANGELOG.md) | 版本变更记录 |
| [docs/DISCLAIMER.md](docs/DISCLAIMER.md) | 完整免责声明 |
| [openclaw/kuake_skill/](openclaw/kuake_skill/) | OpenClaw skill 包，用户可将项目工作区作为 OpenClaw workspace 加载以启用 kuake CLI 能力 |

## 免责声明

本工具为**非官方**第三方项目，与夸克网盘官方无关；使用可能导致账号风险、数据丢失或 API 变更导致不可用等，**风险自负**。使用即表示您已阅读并同意完整条款，详见 [docs/DISCLAIMER.md](docs/DISCLAIMER.md)。

## 许可证

本项目采用 **AGPL-3.0**，详见 [LICENSE](LICENSE)。衍生作品须以相同协议开源。**商业使用**请联系维护者取得授权。

## 贡献者

感谢所有贡献者，包括但不限于：

- [@Cody292](https://github.com/Cody292) — 并行上传（PR #13）、`--policy` 上传策略（PR #16）、user 容量与 `--version`（PR #17）、并行上传优化（PR #18）

欢迎提交 Issue 与 Pull Request。
