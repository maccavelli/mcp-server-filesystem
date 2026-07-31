package util

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/maccavelli/mcplib"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// UniversalBaseInput is the shared optional telemetry base type from mcplib.
type UniversalBaseInput = mcplib.UniversalBaseInput

// AbortIfCancelled returns an MCP error result when ctx has been canceled.
func AbortIfCancelled(ctx context.Context) (*mcp.CallToolResult, bool) {
	if err := ctx.Err(); err != nil {
		res := &mcp.CallToolResult{}
		res.SetError(err)
		return res, true
	}
	return nil, false
}

// HardenedResourceHandler wraps an MCP resource handler with panic recovery
// to prevent unhandled errors from crashing the server process.
func HardenedResourceHandler(
	handler func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error),
) func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (res *mcp.ReadResourceResult, err error) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("resource handler panic", "panic", r, "trace", string(debug.Stack()))
				err = fmt.Errorf("internal error: %v", r)
			}
		}()
		return handler(ctx, req)
	}
}
