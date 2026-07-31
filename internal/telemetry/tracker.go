package telemetry

import "github.com/maccavelli/mcplib"

// GlobalRingBuffer is the canonical type alias for downstream consumers
type GlobalRingBuffer = mcplib.LogBuffer

// Tracker holds telemetry state for the filesystem MCP server.
type Tracker struct {
	RingBuffer *GlobalRingBuffer
}

// GlobalTracker is the process-wide telemetry instance.
var GlobalTracker = &Tracker{
	RingBuffer: mcplib.NewLogBuffer(),
}
