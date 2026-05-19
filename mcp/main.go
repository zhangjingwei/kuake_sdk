package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/zhangjingwei/kuake_cli/internal/guard"
	"github.com/zhangjingwei/kuake_cli/sdk"
)

func main() {
	// Redirect all Go-level logging to stderr first.
	// This must be the very first line before any library init.
	log.SetOutput(os.Stderr)

	// Capture the real stdout fd for the MCP JSON-RPC stream, then
	// redirect os.Stdout to stderr so accidental fmt.Print* calls from
	// any library cannot pollute the protocol channel.
	realStdout := os.Stdout
	os.Stdout = os.Stderr

	// Build client from env. If KUAKE_COOKIE is not set, client stays
	// nil and each tool call returns a clear error — server does not exit.
	var client *sdk.QuarkClient
	if cookie := sdk.ResolveEnvCookieString(); cookie != "" {
		client = sdk.NewQuarkClient(cookie)
	} else {
		fmt.Fprintln(os.Stderr, "[kuake-mcp] KUAKE_COOKIE not set — tool calls will return errors until configured")
	}

	g := guard.NewGuard()
	mcpServer := newMCPServer(client, g)

	stdioSrv := server.NewStdioServer(mcpServer)
	if err := stdioSrv.Listen(context.Background(), os.Stdin, realStdout); err != nil {
		fmt.Fprintln(os.Stderr, "[kuake-mcp] server error:", err)
	}
}
