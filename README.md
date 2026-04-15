# Kuake CLI

[License](LICENSE)

夸克网盘文件管理 CLI 工具。

## 开源说明

本项目采用 **AGPL-3.0** 开源；**商业使用**（含 SaaS、商业产品集成等）须另行取得授权。源代码可自本仓库获取，使用 `build.sh` 可本地编译各平台二进制。欢迎通过 Issue / Pull Request 参与。

## 目录

- [功能特性](#功能特性)
- [系统要求](#系统要求)
- [安装](#安装)
- [快速开始](#快速开始)
- [文档与变更](#文档与变更)
- [参与开发](#参与开发)
- [免责声明](#免责声明)
- [许可证](#许可证)
- [Star History](#star-history)
- [贡献者](#贡献者)

## 功能特性

- 用户信息与网盘目录列表、文件详情、上传/下载、创建目录、移动/复制/重命名/删除
- 分享创建与取消、分享列表、转存他人分享
- JSON 输出、管道模式（与 `jq` 等组合）；可选 **OpenClaw** 技能（见 [openclaw/kuake_skill/](openclaw/kuake_skill/)）：普通用户只需安装 [Releases](https://github.com/zhangjingwei/kuake_cli/releases) 中的 `kuake`、配置 `PATH` 与 `KUAKE_COOKIE`（说明见 [openclaw/kuake_skill/SKILL.md](openclaw/kuake_skill/SKILL.md) 与 [docs/cli.md](docs/cli.md)）

更多用法见 [docs/cli.md](docs/cli.md)。环境变量模板见 [.env.example](.env.example)；存在 `.env` 时 `kuake` 会从当前目录及 `-c` 配置所在目录自动加载（可用 `KUAKE_LOAD_DOTENV=0` 关闭），不覆盖已 export 的变量。

更多文档见：

- [specs/architecture/spec.md](specs/architecture/spec.md)
- [openclaw/kuake_skill/SKILL.md](openclaw/kuake_skill/SKILL.md)（OpenClaw 技能说明）

## 系统要求

- Linux / macOS / Windows
- 有效的夸克网盘账号与 Cookie

## 安装

### 从源码构建

需要 **Go 1.21+**（与 `go.mod` 一致）与 Git。

```bash
git clone https://github.com/zhangjingwei/kuake_cli.git
cd kuake_cli
chmod +x build.sh
./build.sh
```

构建产物位于 `dist/`。

### 预编译二进制

从 [Releases](https://github.com/zhangjingwei/kuake_cli/releases) 下载对应平台文件，文件名与版本以 Release 页为准。

## 快速开始

1. 复制 `[.env.example](.env.example)` 为 `.env`，按文件内说明填入夸克网盘 Cookie（浏览器 F12 → Network 复制）。**勿提交 `.env`。**
2. 在项目目录执行（二进制名以你本机为准）：

```bash
./kuake user
./kuake list "/"
./kuake upload "file.txt" "/file.txt"
```

更多参数与凭证回退方式见 [docs/cli.md](docs/cli.md)；从源码运行见下方「参与开发」。

## 文档与变更


| 文档                                             | 说明                                                                 |
| ---------------------------------------------- | ------------------------------------------------------------------ |
| [docs/cli.md](docs/cli.md)                     | CLI 配置、命令表、JSON 约定与示例                                              |
| [docs/CHANGELOG.md](docs/CHANGELOG.md)         | 版本变更记录                                                             |
| [docs/DISCLAIMER.md](docs/DISCLAIMER.md)       | 完整免责声明                                                             |
| [openclaw/kuake_skill/](openclaw/kuake_skill/) | 给 **OpenClaw 普通用户**：把内含 `SKILL.md` 的文件夹配进 OpenClaw 的技能目录；另从 [Releases](https://github.com/zhangjingwei/kuake_cli/releases) 安装 `kuake` 并加入 `PATH`，按 [docs/cli.md](docs/cli.md) 配置 `KUAKE_COOKIE` 等即可，无需本仓库其它文件 |


## 参与开发

**提交 PR 前**（与 `[.github/workflows/ci.yml](.github/workflows/ci.yml)` 中 **lint** / **test** 一致）：

```bash
golangci-lint run ./...
go test ./... -count=1
```

`golangci-lint` 规则见 `[.golangci.yml](.golangci.yml)`；未安装时可用 `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4`，并将 `$(go env GOPATH)/bin` 加入 `PATH`。

**作为其它项目的 Go 模块依赖**：

```bash
go get github.com/zhangjingwei/kuake_cli@latest
```

## 免责声明

本工具为**非官方**第三方项目，与夸克网盘官方无关；使用可能导致账号风险、数据丢失或 API 变更导致不可用等，**风险自负**。使用即表示您已阅读并同意完整条款，详见 [docs/DISCLAIMER.md](docs/DISCLAIMER.md)。

## 许可证

本项目采用 **AGPL-3.0**，详见 [LICENSE](LICENSE)。衍生作品须以相同协议开源。**商业使用**请联系维护者取得授权。

## Star History

[Star History Chart](https://star-history.com/#zhangjingwei/kuake_cli&Date)

## 贡献者

感谢通过 Issue、Pull Request 等形式参与本项目的所有人。完整贡献者统计见仓库 [Contributors](https://github.com/zhangjingwei/kuake_cli/graphs/contributors)。

欢迎提交 Issue 与 Pull Request。