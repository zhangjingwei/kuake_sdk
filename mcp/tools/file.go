package tools

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/zhangjingwei/kuake_cli/internal/guard"
	"github.com/zhangjingwei/kuake_cli/sdk"
)

var newNameRe = regexp.MustCompile(`^[^/]+$`)

// FileTools returns the 9 file operation tool entries.
func FileTools(client *sdk.QuarkClient, g *guard.Guard) []ToolEntry {
	return []ToolEntry{
		quarkList(client),
		quarkInfo(client),
		quarkDownload(client, g),
		quarkUpload(client, g),
		quarkCreate(client, g),
		quarkMove(client, g),
		quarkCopy(client, g),
		quarkRename(client, g),
		quarkDelete(client, g),
	}
}

func quarkList(client *sdk.QuarkClient) ToolEntry {
	tool := mcp.NewTool("quark_list",
		mcp.WithDescription("List files and folders in a Quark Drive directory."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Remote directory path, e.g. / or /documents"),
		),
	)
	return ToolEntry{Tool: tool, Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return clientError()
		}
		p := req.GetString("path", "")
		resp, err := client.List(p)
		if err != nil {
			return sdkError(err)
		}
		return jsonResult(resp)
	}}
}

func quarkInfo(client *sdk.QuarkClient) ToolEntry {
	tool := mcp.NewTool("quark_info",
		mcp.WithDescription("Get metadata for a file or folder in Quark Drive."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Remote file or folder path"),
		),
	)
	return ToolEntry{Tool: tool, Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return clientError()
		}
		p := req.GetString("path", "")
		resp, err := client.GetFileInfo(p)
		if err != nil {
			return sdkError(err)
		}
		return jsonResult(resp)
	}}
}

func quarkDownload(client *sdk.QuarkClient, g *guard.Guard) ToolEntry {
	tool := mcp.NewTool("quark_download",
		mcp.WithDescription("Download a file from Quark Drive into the local sandbox directory."),
		mcp.WithString("remote_path",
			mcp.Required(),
			mcp.Description("Remote file path to download"),
		),
		mcp.WithString("local_sub_dir",
			mcp.Required(),
			mcp.Description("Subdirectory under KUAKE_DOWNLOAD_DIR to save into. Use \"\" to save directly to KUAKE_DOWNLOAD_DIR. Must not contain '..'"),
		),
	)
	return ToolEntry{Tool: tool, Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return clientError()
		}
		remotePath := req.GetString("remote_path", "")
		localSubDir := req.GetString("local_sub_dir", "")

		if err := g.CheckDownloadDir(localSubDir); err != nil {
			return guardError(err)
		}
		if err := g.CheckPath(remotePath); err != nil {
			return guardError(err)
		}

		info, err := client.GetFileInfo(remotePath)
		if err != nil {
			return sdkError(err)
		}
		if !info.Success {
			return jsonResult(map[string]string{"error": info.Message})
		}
		fid, _ := info.Data["fid"].(string)
		fileName, _ := info.Data["file_name"].(string)

		destDir := filepath.Join(g.DownloadDir(), localSubDir)
		if err := client.DownloadFile(fid, destDir, fileName, nil); err != nil {
			return sdkError(err)
		}
		return jsonResult(map[string]string{
			"saved_to": filepath.Join(destDir, fileName),
		})
	}}
}

func quarkUpload(client *sdk.QuarkClient, g *guard.Guard) ToolEntry {
	tool := mcp.NewTool("quark_upload",
		mcp.WithDescription("Upload a local file to a Quark Drive directory."),
		mcp.WithString("local_path",
			mcp.Required(),
			mcp.Description("Absolute local file path to upload"),
		),
		mcp.WithString("remote_dir",
			mcp.Required(),
			mcp.Description("Destination directory on Quark Drive, e.g. /documents"),
		),
		mcp.WithString("policy",
			mcp.Required(),
			mcp.Description("Conflict policy: \"skip\" skips if same-name exists, \"overwrite\" replaces it, \"rsync\" overwrites only if size differs"),
			mcp.Enum("skip", "overwrite", "rsync"),
		),
	)
	return ToolEntry{Tool: tool, Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return clientError()
		}
		localPath := req.GetString("local_path", "")
		remoteDir := req.GetString("remote_dir", "")
		policy := req.GetString("policy", "skip")
		if policy == "" {
			policy = "skip"
		}

		if err := g.CheckOp("upload"); err != nil {
			return guardError(err)
		}
		if err := g.CheckPath(remoteDir); err != nil {
			return guardError(err)
		}
		fi, err := os.Stat(localPath)
		if err != nil {
			return sdkError(err)
		}
		if err := g.CheckUpload(filepath.Base(localPath), fi.Size()); err != nil {
			return guardError(err)
		}

		resp, err := client.UploadFile(localPath, remoteDir, nil, &sdk.UploadOptions{Policy: sdk.UploadPolicy(policy)})
		if err != nil {
			return sdkError(err)
		}
		return jsonResult(resp)
	}}
}

func quarkCreate(client *sdk.QuarkClient, g *guard.Guard) ToolEntry {
	tool := mcp.NewTool("quark_create",
		mcp.WithDescription("Create a new folder in Quark Drive."),
		mcp.WithString("remote_path",
			mcp.Required(),
			mcp.Description("Full path of the folder to create, e.g. /documents/new-folder"),
		),
	)
	return ToolEntry{Tool: tool, Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return clientError()
		}
		remotePath := req.GetString("remote_path", "")

		if err := g.CheckOp("create"); err != nil {
			return guardError(err)
		}
		if err := g.CheckPath(remotePath); err != nil {
			return guardError(err)
		}

		folderName := path.Base(remotePath)
		parentPath := path.Dir(remotePath)

		var pdirFid string
		if parentPath == "/" || parentPath == "." || parentPath == "" {
			pdirFid = "/"
		} else {
			info, err := client.GetFileInfo(parentPath)
			if err != nil {
				return sdkError(err)
			}
			if !info.Success {
				return jsonResult(map[string]string{"error": "parent directory not found: " + info.Message})
			}
			fid, ok := info.Data["fid"].(string)
			if !ok || fid == "" {
				return jsonResult(map[string]string{"error": "parent directory fid not found"})
			}
			pdirFid = fid
		}

		resp, err := client.CreateFolder(folderName, pdirFid)
		if err != nil {
			return sdkError(err)
		}
		return jsonResult(resp)
	}}
}

func quarkMove(client *sdk.QuarkClient, g *guard.Guard) ToolEntry {
	tool := mcp.NewTool("quark_move",
		mcp.WithDescription("Move a file or folder to a new location in Quark Drive."),
		mcp.WithString("src", mcp.Required(), mcp.Description("Source remote path")),
		mcp.WithString("dst", mcp.Required(), mcp.Description("Destination remote path")),
	)
	return ToolEntry{Tool: tool, Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return clientError()
		}
		src := req.GetString("src", "")
		dst := req.GetString("dst", "")

		if err := g.CheckOp("move"); err != nil {
			return guardError(err)
		}
		if err := g.CheckPath(src); err != nil {
			return guardError(err)
		}
		if err := g.CheckPath(dst); err != nil {
			return guardError(err)
		}
		resp, err := client.Move(src, dst)
		if err != nil {
			return sdkError(err)
		}
		return jsonResult(resp)
	}}
}

func quarkCopy(client *sdk.QuarkClient, g *guard.Guard) ToolEntry {
	tool := mcp.NewTool("quark_copy",
		mcp.WithDescription("Copy a file or folder to a new location in Quark Drive."),
		mcp.WithString("src", mcp.Required(), mcp.Description("Source remote path")),
		mcp.WithString("dst", mcp.Required(), mcp.Description("Destination remote path")),
	)
	return ToolEntry{Tool: tool, Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return clientError()
		}
		src := req.GetString("src", "")
		dst := req.GetString("dst", "")

		if err := g.CheckOp("copy"); err != nil {
			return guardError(err)
		}
		if err := g.CheckPath(src); err != nil {
			return guardError(err)
		}
		if err := g.CheckPath(dst); err != nil {
			return guardError(err)
		}
		resp, err := client.Copy(src, dst)
		if err != nil {
			return sdkError(err)
		}
		return jsonResult(resp)
	}}
}

func quarkRename(client *sdk.QuarkClient, g *guard.Guard) ToolEntry {
	tool := mcp.NewTool("quark_rename",
		mcp.WithDescription("Rename a file or folder. Changes name only — use quark_move to change location."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Full remote path of the file or folder to rename"),
		),
		mcp.WithString("new_name",
			mcp.Required(),
			mcp.Description("New name only (no path separators). To move, use quark_move instead."),
		),
	)
	return ToolEntry{Tool: tool, Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return clientError()
		}
		p := req.GetString("path", "")
		newName := req.GetString("new_name", "")

		if !newNameRe.MatchString(newName) {
			return jsonResult(map[string]string{
				"error": "new_name must be a plain name, not a path; use quark_move to relocate",
			})
		}
		if err := g.CheckOp("rename"); err != nil {
			return guardError(err)
		}
		if err := g.CheckPath(p); err != nil {
			return guardError(err)
		}
		resp, err := client.Rename(p, newName)
		if err != nil {
			return sdkError(err)
		}
		return jsonResult(resp)
	}}
}

func quarkDelete(client *sdk.QuarkClient, g *guard.Guard) ToolEntry {
	tool := mcp.NewTool("quark_delete",
		mcp.WithDescription("Delete a file or folder from Quark Drive."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Remote path of the file or folder to delete"),
		),
	)
	return ToolEntry{Tool: tool, Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return clientError()
		}
		p := req.GetString("path", "")

		if err := g.CheckOp("delete"); err != nil {
			return guardError(err)
		}
		if err := g.CheckPath(p); err != nil {
			return guardError(err)
		}
		resp, err := client.Delete(p)
		if err != nil {
			return sdkError(err)
		}
		return jsonResult(resp)
	}}
}
