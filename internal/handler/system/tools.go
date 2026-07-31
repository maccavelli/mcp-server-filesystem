package system

import (
	"github.com/maccavelli/mcplib"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/maccavelli/mcp-server-filesystem/internal/registry"
	"github.com/maccavelli/mcp-server-filesystem/internal/telemetry"
)

// Register adds the system tools to the registry.
func Register() {
	registry.Global.Register(&diagnosticToolShim{})
}

type diagnosticToolShim struct{}

func (t *diagnosticToolShim) Name() string { return "get_internal_logs" }

func (t *diagnosticToolShim) Register(s *mcp.Server) {
	mcplib.RegisterDiagnosticTool(s, telemetry.GlobalTracker.RingBuffer, "filesystem operations server")
}
