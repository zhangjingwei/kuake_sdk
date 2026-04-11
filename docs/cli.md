# kuake CLI 使用说明

命令行工具的完整参考（配置、子命令、输出约定与示例）。若需在 **OpenClaw** 中集成，另见 [openclaw/references/cli-reference.md](../openclaw/references/cli-reference.md)。

## 配置说明

### 配置文件格式

```json
{
  "Quark": {
    "access_tokens": [
      "__pus=your_pus_value_here;"
    ]
  }
}
```

**重要说明**:
- `access_tokens` 字段是一个字符串数组，支持配置多个用户的 Cookie
- 每个字符串存储的是完整的 Cookie 字符串（所有 cookie 用分号和空格分隔）
- 从浏览器开发者工具中复制完整的 Cookie 值
- 示例格式：`cookie1=value1; cookie2=value2; cookie3=value3`
- 支持多用户配置（在数组中添加多个 Cookie 字符串）

**安全提示**:
- `config.json` 文件包含敏感信息，请不要将其提交到版本控制系统
- `.gitignore` 文件已包含 `config.json`，确保不会被意外提交
- 请妥善保管您的 Cookie，不要分享给他人

## CLI 工具使用

### 基本用法

```bash
kuake [options] <command> [arguments...]
kuake <command> [config.json] [arguments...]  (deprecated: use -c instead)
```

**选项**：
- `-c, --config <path>`: 指定配置文件路径（默认: config.json）
- `-cookies, --cookies <value>`: 直接指定 cookie 值（自动添加 `__pus=` 前缀，绕过配置文件）

### 可用命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `user` | 获取用户信息 | `kuake user` |
| `list [path] [--stream]` | 列出目录内容（默认: "/"），使用 `--stream` 输出流式 JSON 用于管道模式 | `kuake list "/"` 或 `kuake list "/" --stream` |
| `info <path>` | 获取文件/文件夹信息（支持管道模式） | `kuake info "/file.txt"` |
| `download <path> [dest]` | 获取文件下载链接或下载到本地（支持管道模式） | `kuake download "/file.txt"` 或 `kuake download "/file.txt" ./local` |
| `upload <file> <dest> [--max_upload_parallel N]` | 上传文件（上传进度输出到 stderr，支持并行上传） | `kuake upload "file.txt" "/file.txt"` 或 `kuake upload "file.txt" "/file.txt" --max_upload_parallel 4` |
| `create <name> <pdir>` | 创建文件夹（pdir 为父目录路径，根目录使用 "/"） | `kuake create "test_folder" "/"` |
| `move <src> <dest>` | 移动文件/文件夹 | `kuake move "/file.txt" "/folder/"` |
| `copy <src> <dest>` | 复制文件/文件夹 | `kuake copy "/file.txt" "/folder/"` |
| `rename <path> <newName>` | 重命名文件/文件夹 | `kuake rename "/file.txt" "new_name.txt"` |
| `delete <path>` | 删除文件/文件夹（支持管道模式） | `kuake delete "/file.txt"` |
| `share <path> <days> <passcode>` | 创建分享链接 | `kuake share "/file.txt" 7 "false"` |
| `share-delete <share_id_or_path> [share_id_or_path2] ...` | 取消分享（支持通过 share_id 或文件路径） | `kuake share-delete "fdd8bfd93f21491ab80122538bec310d"` 或 `kuake share-delete "/file.txt"` |
| `share-list [page] [size] [orderField] [orderType]` | 获取我的分享列表 | `kuake share-list` 或 `kuake share-list 1 50 "created_at" "desc"` |
| `share-save <share_link> [passcode] [dest_dir]` | 转存分享文件到自己的网盘 | `kuake share-save "https://pan.quark.cn/s/xxx"` 或 `kuake share-save "https://pan.quark.cn/s/xxx" "1234" "/folder"` |
| `help` | 显示帮助信息 | `kuake help` |

**重要提示**：
- 所有路径参数必须用引号包裹（`"path"`）
- 根目录使用 `"/"` 表示
- `days` 参数：`0`=永久，`1`=1天，`7`=7天，`30`=30天
- `passcode` 参数：`"true"`=需要提取码，`"false"`=不需要提取码
- `share-save` 命令说明：
  - `share_link`: 分享链接（如 `https://pan.quark.cn/s/xxx`），会自动提取 pwd_id
  - `passcode`: 提取码（可选），如果分享链接中包含提取码会自动提取
  - `dest_dir`: 目标目录（可选，默认 `"/"`），可以是路径或 FID
  - 默认会转存分享中的所有文件到指定目录
- **并行上传参数**：
  - `--max_upload_parallel N`：设置并行上传的分片数量（1-16，默认 4）
  - 也支持通过环境变量 `KUAKE_UPLOAD_PARALLEL` 设置
  - 并行上传仅在满足条件时启用（新上传、多分片文件等）
  - 断点续传时自动使用顺序上传，确保兼容性
- **管道模式**：
  - `list` 命令使用 `--stream` 选项输出流式 JSON（每行一个文件对象）
  - `delete`、`info`、`download` 命令支持从 stdin 读取 JSON 输入
  - 自动检测 stdin，有数据时自动进入管道模式
  - 每行输入应为 JSON 对象，包含 `path` 或 `fid` 字段
  - 支持与其他 Unix 工具组合使用，如 `jq`、`grep`、`head` 等

### 输出格式

所有命令的结果都以 JSON 格式输出到 stdout：

**成功响应**：
```json
{
  "success": true,
  "code": "OK",
  "message": "操作成功",
  "data": {
    ...
  }
}
```

**错误响应**：
```json
{
  "success": false,
  "code": "ERROR_CODE",
  "message": "错误描述",
  "error": "详细错误信息"
}
```

**注意**：
- 所有结果（包括成功和错误）都以 JSON 格式输出到 stdout
- 上传进度、帮助信息和序列化错误输出到 stderr
- 这样设计便于其他进程解析 JSON 结果，进度信息不会混入 JSON 输出

### 退出码

- `0`: 操作成功
- `1`: 操作失败

### 使用示例

```bash
# 获取用户信息（使用默认配置文件 config.json）
./kuake-{version}-{os}-{arch} user

# 获取用户信息（使用自定义配置文件）
./kuake-{version}-{os}-{arch} user custom.json

# 列出根目录
./kuake-{version}-{os}-{arch} list "/"

# 获取文件信息
./kuake-{version}-{os}-{arch} info "/file.txt"

# 获取文件下载链接
./kuake-{version}-{os}-{arch} download "/file.txt"

# 上传文件（使用默认并行度 4）
./kuake-{version}-{os}-{arch} upload "file.txt" "/file.txt"

# 上传文件（指定并行度为 8）
./kuake-{version}-{os}-{arch} upload "file.txt" "/file.txt" --max_upload_parallel 8

# 上传文件（通过环境变量设置并行度）
export KUAKE_UPLOAD_PARALLEL=8
./kuake-{version}-{os}-{arch} upload "file.txt" "/file.txt"

# 创建文件夹（根目录）
./kuake-{version}-{os}-{arch} create "test_folder" "/"

# 移动文件
./kuake-{version}-{os}-{arch} move "/file.txt" "/folder/"

# 复制文件
./kuake-{version}-{os}-{arch} copy "/file.txt" "/folder/"

# 重命名文件
./kuake-{version}-{os}-{arch} rename "/file.txt" "new_name.txt"

# 删除文件
./kuake-{version}-{os}-{arch} delete "/file.txt"

# 创建分享链接（7天，不需要提取码）
./kuake-{version}-{os}-{arch} share "/file.txt" 7 "false"

# 创建分享链接（30天，需要提取码，使用自定义配置文件）
./kuake-{version}-{os}-{arch} share "/file.txt" 30 "true" custom.json

# 取消分享（通过 share_id）
./kuake-{version}-{os}-{arch} share-delete "fdd8bfd93f21491ab80122538bec310d"

# 取消分享（通过文件路径，会自动查找对应的 share_id）
./kuake-{version}-{os}-{arch} share-delete "/file.txt"

# 同时取消多个分享
./kuake-{version}-{os}-{arch} share-delete "share_id1" "share_id2" "/file.txt"

# 获取我的分享列表（使用默认参数）
./kuake-{version}-{os}-{arch} share-list

# 获取我的分享列表（指定分页和排序参数）
./kuake-{version}-{os}-{arch} share-list 1 50 "created_at" "desc"

# 转存分享文件到根目录
./kuake-{version}-{os}-{arch} share-save "https://pan.quark.cn/s/xxx"

# 转存分享文件（指定提取码和目标目录）
./kuake-{version}-{os}-{arch} share-save "https://pan.quark.cn/s/xxx" "1234" "/folder"

# 查看帮助
./kuake-{version}-{os}-{arch} help

# 使用 -cookies 参数（绕过配置文件，只需提供 cookie 值）
./kuake-{version}-{os}-{arch} -cookies "your_cookie_value_here" user
./kuake-{version}-{os}-{arch} -cookies "your_cookie_value_here" upload "file.txt" "/folder/file.txt"

# 管道模式示例
# 列出文件并批量删除
./kuake-{version}-{os}-{arch} list "/photos" --stream | ./kuake-{version}-{os}-{arch} delete

# 列出文件并获取每个文件的信息
./kuake-{version}-{os}-{arch} list "/" --stream | ./kuake-{version}-{os}-{arch} info

# 列出文件并获取下载链接
./kuake-{version}-{os}-{arch} list "/documents" --stream | ./kuake-{version}-{os}-{arch} download

# 结合 jq 进行过滤：列出大文件并删除
./kuake-{version}-{os}-{arch} list "/" --stream | jq -r 'select(.size > 1000000) | .path' | ./kuake-{version}-{os}-{arch} delete

# 列出文件并下载到指定目录
./kuake-{version}-{os}-{arch} list "/videos" --stream | ./kuake-{version}-{os}-{arch} download "./downloads"
```

**注意**：
- 示例中的 `{version}`、`{os}`、`{arch}` 需要替换为实际值
- Windows 用户需要添加 `.exe` 扩展名并使用 `.\` 前缀
- 如果已添加到 PATH，可以直接使用 `kuake` 命令

## 注意事项

- **文件名格式**：二进制文件名包含版本号，格式为 `kuake-{version}-{os}-{arch}` 或 `kuake-{version}-{os}-{arch}.exe`（Windows）
- **执行权限**：Linux/macOS 二进制文件已包含执行权限，可直接使用 `./` 前缀执行
- **路径参数**：所有路径参数必须用引号包裹（包含空格或特殊字符时），例如：`"./file name.txt"`、`"/path/to/file"`
- **跨平台路径支持**：
  - Windows 用户可以使用 Windows 风格的路径（`d:\a\b\c`），会自动转换为 Unix 风格
  - Linux/macOS 用户继续使用标准 Unix 路径格式（`/a/b/c`）
  - 所有路径最终都会标准化为 Unix 风格，确保跨平台一致性
- **配置文件**：
  - 默认配置文件路径：`config.json`（当前目录）
  - 配置文件参数是可选的，放在命令之后、其他参数之前
  - 配置文件参数必须是 `.json` 扩展名
  - 示例：`kuake user custom.json`（使用自定义配置文件）
- **Cookie 参数**：
  - 使用 `-cookies` 或 `--cookies` 参数可直接指定 cookie 值，无需配置文件
  - 只需提供 cookie 值，工具会自动添加 `__pus=` 前缀和末尾分号
  - 使用 `-cookies` 参数时，不会读取配置文件，提高效率并避免不一致
  - 示例：`kuake -cookies "your_cookie_value" user`
- **操作说明**：
  - 所有操作都通过夸克网盘 API 进行
  - 需要有效的 Cookie（access_token）才能使用
  - 上传操作支持进度显示（输出到 stderr）
  - 上传操作支持并行上传，可通过 `--max_upload_parallel` 参数或 `KUAKE_UPLOAD_PARALLEL` 环境变量配置（1-16，默认 4）
  - 删除目录会递归删除所有子文件和子目录
- **输出格式**：
  - CLI 工具的所有结果以 JSON 格式输出到 stdout，方便其他进程解析
  - 上传进度、帮助信息和序列化错误输出到 stderr，不会混入 JSON 输出
  - 成功时退出码为 0，失败时为 1
