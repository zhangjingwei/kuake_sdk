package tools

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/zhangjingwei/kuake_cli/internal/guard"
	"github.com/zhangjingwei/kuake_cli/sdk"
)

// ToolEntry pairs an MCP tool definition with its handler.
type ToolEntry struct {
	Tool    mcp.Tool
	Handler server.ToolHandlerFunc
}

// AllTools returns all 14 tool entries.
func AllTools(client *sdk.QuarkClient, g *guard.Guard) []ToolEntry {
	entries := FileTools(client, g)
	entries = append(entries, ShareTools(client, g)...)
	return entries
}

func clientError() (*mcp.CallToolResult, error) {
	return jsonResult(map[string]string{
		"error": "KUAKE_COOKIE not set; set the env var and restart kuake-mcp",
	})
}

func guardError(err error) (*mcp.CallToolResult, error) {
	return jsonResult(map[string]string{"error": err.Error()})
}

func sdkError(err error) (*mcp.CallToolResult, error) {
	return jsonResult(map[string]string{"error": err.Error()})
}

func jsonResult(v interface{}) (*mcp.CallToolResult, error) {
	b, _ := json.Marshal(v)
	return mcp.NewToolResultText(string(b)), nil
}
