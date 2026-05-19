package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/zhangjingwei/kuake_cli/internal/guard"
	"github.com/zhangjingwei/kuake_cli/sdk"
)

// ShareTools returns the 5 share and user tool entries.
func ShareTools(client *sdk.QuarkClient, g *guard.Guard) []ToolEntry {
	return []ToolEntry{
		quarkUser(client, g),
		quarkShareCreate(client, g),
		quarkShareDelete(client, g),
		quarkShareList(client),
		quarkShareSave(client, g),
	}
}

func quarkUser(client *sdk.QuarkClient, g *guard.Guard) ToolEntry {
	tool := mcp.NewTool("quark_user",
		mcp.WithDescription("Get Quark Drive account info: nickname, storage capacity, membership status."),
	)
	return ToolEntry{Tool: tool, Handler: func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return clientError()
		}
		if err := g.CheckOp("user"); err != nil {
			return guardError(err)
		}
		resp, err := client.GetUserInfo()
		if err != nil {
			return sdkError(err)
		}
		return jsonResult(resp)
	}}
}

func quarkShareCreate(client *sdk.QuarkClient, g *guard.Guard) ToolEntry {
	tool := mcp.NewTool("quark_share_create",
		mcp.WithDescription("Create a share link for a file or folder."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Remote path of the file or folder to share"),
		),
		mcp.WithNumber("expire_days",
			mcp.Required(),
			mcp.Description("Expiry in days: 1, 7, 30, or 0 for permanent"),
		),
		mcp.WithString("passcode",
			mcp.Required(),
			mcp.Description("Set to \"\" for no passcode. Any non-empty string enables an auto-generated passcode returned in the response."),
		),
	)
	return ToolEntry{Tool: tool, Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return clientError()
		}
		p := req.GetString("path", "")
		expireDays := req.GetInt("expire_days", 0)
		passcode := req.GetString("passcode", "")

		if err := g.CheckOp("share_create"); err != nil {
			return guardError(err)
		}
		if err := g.CheckPath(p); err != nil {
			return guardError(err)
		}

		shareInfo, err := client.CreateShare(p, expireDays, passcode != "")
		if err != nil {
			return sdkError(err)
		}
		return jsonResult(map[string]interface{}{
			"share_url":  shareInfo.ShareURL,
			"pwd_id":     shareInfo.PwdID,
			"passcode":   shareInfo.Passcode,
			"expires_at": shareInfo.ExpiresAt,
		})
	}}
}

func quarkShareDelete(client *sdk.QuarkClient, g *guard.Guard) ToolEntry {
	tool := mcp.NewTool("quark_share_delete",
		mcp.WithDescription("Cancel a share link by its share ID (pwd_id)."),
		mcp.WithString("share_id",
			mcp.Required(),
			mcp.Description("Share ID (pwd_id) as returned by quark_share_create or quark_share_list"),
		),
	)
	return ToolEntry{Tool: tool, Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return clientError()
		}
		shareID := req.GetString("share_id", "")

		if err := g.CheckOp("share_delete"); err != nil {
			return guardError(err)
		}
		if err := client.DeleteShare([]string{shareID}); err != nil {
			return sdkError(err)
		}
		return jsonResult(map[string]string{"deleted": shareID})
	}}
}

func quarkShareList(client *sdk.QuarkClient) ToolEntry {
	tool := mcp.NewTool("quark_share_list",
		mcp.WithDescription("List your active Quark Drive share links."),
		mcp.WithNumber("page",
			mcp.Required(),
			mcp.Description("Page number starting at 1"),
		),
		mcp.WithNumber("size",
			mcp.Required(),
			mcp.Description("Results per page"),
		),
	)
	return ToolEntry{Tool: tool, Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return clientError()
		}
		page := req.GetInt("page", 1)
		size := req.GetInt("size", 50)

		resp, err := client.GetMyShareList(page, size, "created_at", "desc")
		if err != nil {
			return sdkError(err)
		}
		return jsonResult(resp)
	}}
}

func quarkShareSave(client *sdk.QuarkClient, g *guard.Guard) ToolEntry {
	tool := mcp.NewTool("quark_share_save",
		mcp.WithDescription("Save all files from a share link to your Quark Drive."),
		mcp.WithString("share_url",
			mcp.Required(),
			mcp.Description("Quark Drive share URL, e.g. https://pan.quark.cn/s/xxxxxx"),
		),
		mcp.WithString("passcode",
			mcp.Required(),
			mcp.Description("Share passcode. Use \"\" if no passcode is required."),
		),
		mcp.WithString("dst",
			mcp.Required(),
			mcp.Description("Destination directory on your Quark Drive, e.g. /documents. Use / for root."),
		),
	)
	return ToolEntry{Tool: tool, Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return clientError()
		}
		shareURL := req.GetString("share_url", "")
		passcode := req.GetString("passcode", "")
		dst := req.GetString("dst", "")

		if err := g.CheckOp("share_save"); err != nil {
			return guardError(err)
		}
		if err := g.CheckPath(dst); err != nil {
			return guardError(err)
		}

		shareInfo, err := client.GetShareInfo(shareURL)
		if err != nil {
			return sdkError(fmt.Errorf("invalid share link: %w", err))
		}

		stokenData, err := client.GetShareStoken(shareInfo.PwdID, passcode)
		if err != nil {
			return sdkError(fmt.Errorf("failed to get share token: %w", err))
		}
		stoken, ok := stokenData["stoken"].(string)
		if !ok || stoken == "" {
			return jsonResult(map[string]string{"error": "stoken not found in response"})
		}

		toPdirFid := "0" // root
		if dst != "" && dst != "/" {
			info, err := client.GetFileInfo(dst)
			if err != nil {
				return sdkError(err)
			}
			if !info.Success {
				return jsonResult(map[string]string{"error": "destination directory not found: " + info.Message})
			}
			fid, ok := info.Data["fid"].(string)
			if !ok || fid == "" {
				return jsonResult(map[string]string{"error": "destination fid not found"})
			}
			toPdirFid = fid
		}

		result, err := client.SaveShareFile(shareInfo.PwdID, stoken, []string{}, []string{}, toPdirFid, true)
		if err != nil {
			return sdkError(err)
		}
		return jsonResult(map[string]interface{}{
			"dst":       dst,
			"save_data": result,
		})
	}}
}
