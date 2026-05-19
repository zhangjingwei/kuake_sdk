# MCP Server Design — kuake-mcp

Date: 2026-05-19  
Status: Approved

## Overview

Build a standalone MCP server binary (`kuake-mcp`) on top of the existing `sdk/` package, exposing all Quark Drive file operations as MCP tools for local AI agents (Claude Code, Cursor). Transport: stdio only.

---

## 1. Repository Structure

```
kuake_cli/
├── cmd/                    ← existing CLI (unchanged except -c flag removal)
├── sdk/                    ← existing SDK (config.json support removed)
├── internal/
│   └── guard/
│       ├── guard.go        ← blacklist logic
│       └── guard_test.go
├── mcp/
│   ├── main.go             ← kuake-mcp binary entry point
│   ├── server.go           ← tool registration + stdio server
│   └── tools/
│       ├── file.go         ← 9 file operation tools
│       └── share.go        ← 5 share + user tools
└── build.sh                ← extended with kuake-mcp targets
```

**Dependency graph (unidirectional, no cycles):**

```
mcp/        → internal/guard/
mcp/        → sdk/
cmd/        → sdk/
internal/guard/ ↛ sdk/
cmd/        ↛ mcp/
```

**MCP protocol library:** `github.com/mark3labs/mcp-go`

---

## 2. SDK Changes — Remove config.json

Remove from `sdk/config.go`:
- JSON file parsing branch in `LoadConfig()`
- `SaveConfig()` function entirely

Auth priority chain after removal:
```
KUAKE_COOKIE env → KUAKE_PUS+KUAKE_PUUS env → -cookies CLI flag → .env file
```

The `Config` struct is retained for internal SDK use. Remove the `-c <config-path>` flag from `cmd/main.go`.

---

## 3. Guard — Blacklist System

### Environment Variables

| Variable | Separator | Example | Description |
|----------|-----------|---------|-------------|
| `KUAKE_DENY_OPS` | `:` | `delete:move` | Disable named operations |
| `KUAKE_DENY_PATHS` | `:` | `/备份:/重要资料` | Protect remote path prefixes |
| `KUAKE_DENY_EXTS` | `:` | `.env:.key:.pem` | Block upload by file extension |
| `KUAKE_MAX_UPLOAD_MB` | — | `200` | Max upload size, 0 = unlimited |
| `KUAKE_DOWNLOAD_DIR` | — | `/tmp/quark` | Sandbox all downloads to this directory; if unset, defaults to current working directory |

Multi-value separator is `:` throughout. Quark remote paths never contain `:`.

### Interface

```go
type Guard struct { /* loaded from env at startup */ }

func NewGuard() *Guard

func (g *Guard) CheckOp(op string) error
func (g *Guard) CheckPath(remotePath string) error
func (g *Guard) CheckUpload(filename string, sizeBytes int64) error
func (g *Guard) CheckDownloadDir(localSubPath string) error
```

### Path Matching

Uses `path.Clean` + explicit suffix check to avoid sibling path false positives:

```go
func isPathDenied(target, deny string) bool {
    t := path.Clean(target)
    d := path.Clean(deny)
    return t == d || strings.HasPrefix(t, d+"/")
}
```

`/备份2` is NOT blocked by a `/备份` rule.

### Tool → Guard Mapping

| Tool | CheckOp | CheckPath | CheckUpload | CheckDownloadDir |
|------|:-------:|:---------:|:-----------:|:----------------:|
| `quark_list` | — | — | — | — |
| `quark_info` | — | — | — | — |
| `quark_user` | ✓ | — | — | — |
| `quark_download` | — | ✓ (remote) | — | ✓ |
| `quark_upload` | ✓ | ✓ (remote) | ✓ | — |
| `quark_create` | ✓ | ✓ | — | — |
| `quark_move` | ✓ | ✓ (src+dst) | — | — |
| `quark_copy` | ✓ | ✓ (src+dst) | — | — |
| `quark_rename` | ✓ | ✓ | — | — |
| `quark_delete` | ✓ | ✓ | — | — |
| `quark_share_create` | ✓ | ✓ | — | — |
| `quark_share_delete` | ✓ | — | — | — |
| `quark_share_list` | — | — | — | — |
| `quark_share_save` | ✓ | — | — | — |

`quark_list` and `quark_info` are read-only with no side effects; no guard checks required.  
`quark_user` exposes PII (phone, membership); gated by `CheckOp("user")`.

---

## 4. MCP Tools

All tools use MCP's native JSON input schema (named fields, explicit types). No concatenated string parameters.

### tools/file.go

**`quark_list`**
- `path`: string, required — remote directory path

**`quark_info`**
- `path`: string, required — remote file or folder path

**`quark_download`**
- `remote_path`: string, required
- `local_sub_dir`: string, required — subdirectory relative to `KUAKE_DOWNLOAD_DIR`; empty string `""` means download directly to `KUAKE_DOWNLOAD_DIR`; must not contain `..`

**`quark_upload`**
- `local_path`: string, required — absolute local file path
- `remote_dir`: string, required — destination directory on Quark
- `policy`: enum `["skip","overwrite","rename"]`, required, default `"skip"`

**`quark_create`**
- `remote_path`: string, required — full path of folder to create

**`quark_move`**
- `src`: string, required
- `dst`: string, required

**`quark_copy`**
- `src`: string, required
- `dst`: string, required

**`quark_rename`**
- `path`: string, required — full remote path of file/folder
- `new_name`: string, required — new name only; pattern `^[^/]+$` (no path separators)
- If `new_name` contains `/`: return error `"new_name must be a plain name, not a path; use quark_move to relocate"`

**`quark_delete`**
- `path`: string, required

### tools/share.go

**`quark_share_create`**
- `path`: string, required
- `expire_days`: integer, required, enum `[1, 7, 30, 0]`, default `7` (`0` = permanent)
- `passcode`: string, required — empty string `""` means no passcode

**`quark_share_delete`**
- `share_id`: string, required

**`quark_share_list`**
- `page`: integer, required, default `1`
- `size`: integer, required, default `30`

**`quark_share_save`**
- `share_url`: string, required
- `passcode`: string, required — empty string `""` means no passcode
- `dst`: string, required — destination directory on Quark

**`quark_user`**  
No parameters. Returns account info (nickname, capacity, membership). Gated by `CheckOp("user")`.

### Return Format

All tools return a JSON string:
- Success: `{"data": ...}`
- Error: `{"error": "human-readable message"}`

---

## 5. Server Entry Point

### Startup

`mcp/main.go`:

```go
func main() {
    // 1. Redirect all logging to stderr — must be first line
    log.SetOutput(os.Stderr)

    // 2. Capture real stdout fd for MCP protocol, redirect fd1 → fd2
    //    so no library can accidentally pollute the JSON-RPC stream
    mcpStdout := os.Stdout
    os.Stdout = os.Stderr

    // 3. Resolve cookie from env (sdk.ResolveEnvCookieString)
    //    client may be nil if no cookie — tools return MCP errors at call time

    // 4. Init Guard from env

    // 5. Register tools, start stdio server using mcpStdout
}
```

**Startup failure policy:** The server process always starts. If `KUAKE_COOKIE` is not set, each tool call returns a protocol-compliant MCP error: `"KUAKE_COOKIE not set; configure the environment variable and restart"`. The process never calls `os.Exit` on startup — a sudden exit crashes the MCP client (Claude Code / Cursor).

### stdout Discipline

- fd 1 (stdout) is reserved exclusively for MCP JSON-RPC protocol messages
- All `log.*`, `fmt.Print*`, and third-party library output must go to stderr
- mcp-go is initialized with an explicit `io.Writer` pointing to the captured stdout fd, not `os.Stdout`
- This prevents any library that writes to `os.Stdout` from polluting the protocol stream

### build.sh Additions

```bash
GOOS=linux   GOARCH=amd64  go build -trimpath -ldflags="-s -w" -o dist/kuake-mcp-linux-amd64    ./mcp
GOOS=linux   GOARCH=arm64  go build -trimpath -ldflags="-s -w" -o dist/kuake-mcp-linux-arm64    ./mcp
GOOS=darwin  GOARCH=amd64  go build -trimpath -ldflags="-s -w" -o dist/kuake-mcp-darwin-amd64   ./mcp
GOOS=darwin  GOARCH=arm64  go build -trimpath -ldflags="-s -w" -o dist/kuake-mcp-darwin-arm64   ./mcp
GOOS=windows GOARCH=amd64  go build -trimpath -ldflags="-s -w" -o dist/kuake-mcp-windows-amd64.exe ./mcp
```

### Claude Code Integration (`.mcp.json`)

```json
{
  "mcpServers": {
    "quark": {
      "command": "./dist/kuake-mcp-darwin-arm64",
      "env": {
        "KUAKE_COOKIE": "...",
        "KUAKE_DENY_OPS": "delete",
        "KUAKE_DENY_PATHS": "/备份",
        "KUAKE_DENY_EXTS": ".env:.key:.pem",
        "KUAKE_DOWNLOAD_DIR": "/tmp/quark-downloads"
      }
    }
  }
}
```

---

## 6. What Is Not In Scope

- Baidu Pan support (separate future effort)
- HTTP/SSE transport
- Multi-tenant auth
- Tool-level audit logging
