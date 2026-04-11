# 项目架构规范

## 1. 系统架构总览

Kuake 是一个以 Go 构建的命令行工具，提供夸克网盘操作能力。项目整体架构分为如下层次：

- CLI 层：`cmd/`，负责解析命令、接收用户输入、呈现输出。
- SDK 层：`sdk/`，封装夸克网盘 API 调用、请求构建、响应解析、错误处理等业务逻辑。
- 传输层：基于 Go `net/http`，统一请求、Cookie 处理、超时、上传/下载时限控制。
- 外部 API 层：夸克网盘官方/逆向 API 接口，主要通过 `drive-pc.quark.cn`、`pan.quark.cn` 等域名访问。

### 1.1 分层架构图

```
[CLI] <--> [SDK] <--> [Transport / HTTP] <--> [Quark API]
            |           
            |-- Config
            |-- Request / Response
            |-- Error handling
```

## 2. 技术栈与依赖说明

- 语言：Go 1.21
- 标准库：`net/http`, `encoding/json`, `sync`, `time`, `os`, `fmt`, `io`, `bytes`。
- 构建：`build.sh` 负责本地编译命令行二进制。发布产物存放 `dist/`。
- 文档：`README.md`, `docs/`, `openspec/`。

本项目避免引入大型第三方库，优先使用标准库完成 HTTP 请求、JSON 编解码、文件读写、并发控制等核心功能。

## 3. 目录结构与模块划分

```
/               # 仓库根目录
  cmd/          # CLI 入口和命令解析
  sdk/          # 核心 SDK，包括 API 客户端、数据结构、业务操作
  docs/         # 用户文档与 CLI 使用说明
  openspec/     # 规格、设计、变更管理文档
  specs/        # 项目架构与规范说明文档
  dist/         # 构建产物（不纳入版本控制）
```

### 核心目录说明

- `cmd/`：命令行入口，负责命令解析、参数校验、与 SDK 的交互。
- `sdk/`：实现夸克网盘客户端和业务操作逻辑。
- `docs/`：用户使用文档，包含 CLI 说明与示例。
- `openspec/`：OpenSpec 变更管理与规范文档。
- `specs/architecture/`：项目长期架构规范和设计文档。

## 4. 核心组件设计

### 4.1 QuarkClient

`sdk.QuarkClient` 是核心 API 客户端，负责：

- 维护 `baseURL`, `accessTokens`, `cookies`。
- 统一构建请求、设置默认请求头。
- 处理认证检查、token 轮换、失败 token 标记。
- 提供上传、下载、文件操作、分享操作等高层业务接口。

### 4.2 FileOps

文件操作模块负责：

- 文件列表、目录遍历、文件信息获取。
- 上传前置验证、上传授权、上传完成确认。
- 文件移动、复制、重命名、删除、创建目录。
- 下载请求构建与传输控制。

### 4.3 UserOps

用户操作模块负责：

- 初始化用户身份信息。
- 读取 `config.json` 中的 access token。
- 获取账户基本信息与会员容量信息。
- 管理认证有效性与重试策略。

### 4.4 ShareOps

分享操作模块负责：

- 创建、删除、查询分享链接。
- 管理分享密码与分享页面令牌。
- 保存他人分享内容到自己的网盘。

### 4.5 Config

配置模块负责：

- 解析并加载本地 `config.json`。
- 提供默认配置路径 `config.json`。
- 校验必需配置项（如 `access_tokens`）。
- 提供测试环境下的临时配置加载机制。

## 5. 核心组件交互关系

```
cmd/ --> sdk.NewQuarkClient --> QuarkClient
            |                    |-- Config
            |                    |-- Request/Response
            |                    |-- Error handling
            |                    |-- Concurrency management
```

`cmd/` 调用 `sdk/` 提供的高层接口，`sdk/` 负责所有网络请求与业务逻辑。`sdk/` 内部使用 `QuarkClient` 做为统一入口，`QuarkClient` 再调用具体的文件、用户、分享业务方法。

## 6. API 路径索引表

### 6.1 用户信息

- `https://pan.quark.cn/account/info`
- `https://drive-pc.quark.cn/1/clouddrive/member`

### 6.2 文件上传

- `https://drive-pc.quark.cn/1/clouddrive/file/upload/pre`
- `https://drive-pc.quark.cn/1/clouddrive/file/update/hash`
- `https://drive-pc.quark.cn/1/clouddrive/file/upload/auth`
- `https://drive-pc.quark.cn/1/clouddrive/file/upload/finish`

### 6.3 文件下载

- `https://drive-pc.quark.cn/1/clouddrive/file/download`

### 6.4 文件操作

- `https://drive-pc.quark.cn/1/clouddrive/file/sort`
- `https://drive-pc.quark.cn/1/clouddrive/file/move`
- `https://drive-pc.quark.cn/1/clouddrive/file/copy`
- `https://drive-pc.quark.cn/1/clouddrive/file/rename`
- `https://drive-pc.quark.cn/1/clouddrive/file/delete`
- `https://drive-pc.quark.cn/1/clouddrive/file`

### 6.5 分享相关

- `https://drive-pc.quark.cn/1/clouddrive/share`
- `https://drive-pc.quark.cn/1/clouddrive/share/password`
- `https://drive-pc.quark.cn/1/clouddrive/share/delete`
- `https://drive-pc.quark.cn/1/clouddrive/share/mypage/detail`
- `https://drive-pc.quark.cn/1/clouddrive/share/sharepage/token`
- `https://drive-pc.quark.cn/1/clouddrive/share/sharepage/detail`
- `https://drive-pc.quark.cn/1/clouddrive/share/sharepage/save`

## 7. 并发模型规范

- SDK 使用 Go 的并发原语进行安全访问。
- `QuarkClient` 内部通过 `sync.RWMutex` 保护认证检查和失败 token 状态。
- 上传相关流程应保持单个请求超时和并发控制，避免对 Quark API 产生过多并发压力。
- 读取和写入共享状态时，应优先使用 `RLock`/`RUnlock` 进行并发安全读取。

## 8. 错误处理策略

- 可重试错误：网络中断、API 超时、token 轮换失败前的临时认证问题。
- 不可重试错误：配置缺失、参数校验失败、API 返回明确业务错误（例如文件不存在、权限不足）。
- 统一将 API 返回映射为 `StandardResponse` 或 Go `error`，上层调用者负责选择是否重试。
- 关键业务流程应区分“重试后可恢复”和“禁止重试”的错误类型。

## 9. 开发环境搭建指南

- 安装 Go 1.21
- 克隆仓库后，运行 `./build.sh` 进行本地构建
- 在项目根目录创建 `config.json` 用于测试和运行
- 运行 `go test ./...` 进行单元测试

## 10. 代码规范说明

- 优先使用标准库，避免额外依赖。
- Go 代码遵循 `gofmt` 格式化。
- 包名应简洁明了，SDK 中保持 `sdk` 包作为业务实现入口。
- 配置读取、错误检查、请求构造等逻辑应保持清晰分层。
- 文档目录与规范目录应按职责组织，不将架构文档与代码混合。
