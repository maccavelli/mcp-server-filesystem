// Package main implements the MCP filesystem server, providing sandboxed
// file-system operations over the Model Context Protocol stdio transport.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maccavelli/mcplib"

	"github.com/maccavelli/mcp-server-filesystem/internal/config"
	"github.com/maccavelli/mcp-server-filesystem/internal/handler/filesystem"
	"github.com/maccavelli/mcp-server-filesystem/internal/handler/system"
	"github.com/maccavelli/mcp-server-filesystem/internal/pathutil"
	"github.com/maccavelli/mcp-server-filesystem/internal/registry"
	"github.com/maccavelli/mcp-server-filesystem/internal/telemetry"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var exitFunc = os.Exit

func main() {
	// Defense-in-depth: Unmanaged Standalone Fallbacks
	if _, exists := os.LookupEnv("GOMEMLIMIT"); !exists {
		os.Setenv("GOMEMLIMIT", "1024MiB")
	}
	if _, exists := os.LookupEnv("GOMAXPROCS"); !exists {
		os.Setenv("GOMAXPROCS", "2")
	}
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		printVersion()
		exitFunc(0)
	}

	// Allowed directories from arguments.
	dirs := flag.Args()
	if len(dirs) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: mcp-server-filesystem [allowed-directory] [additional-directories...]")
		fmt.Fprintln(os.Stderr, "Note: Allowed directories can also be provided via MCP roots protocol.")
	}

	// Redirect os.Stdout to stderr so only MCP JSON goes to real stdout.
	realStdout := os.Stdout
	os.Stdout = os.Stderr

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	cleanupLogs := mcplib.SetupStandardLogging("filesystem", telemetry.GlobalTracker.RingBuffer)
	defer cleanupLogs()

	logger := slog.Default()
	slog.Info("[BACKPLANE] SPAWN "+config.Name, "version", Version, "allowed_dirs", dirs)

	pipeline := mcplib.NewStdioPipeline(os.Stdin, realStdout, stop)

	if err := run(ctx, dirs, logger, pipeline.Reader, pipeline.Writer); err != nil {
		if mcplib.IsExpectedShutdownErr(err) {
			slog.Info("server shut down gracefully", "error", err)
			_ = pipeline.Flush()
			exitFunc(0)
			return
		}
		slog.Error("server fatal error", "error", err)
		exitFunc(1)
	}
	_ = pipeline.Flush()
}

func run(ctx context.Context, dirs []string, logger *slog.Logger, reader io.ReadCloser, writer io.WriteCloser) error {
	pm := pathutil.NewManager(dirs)

	// Filter to accessible directories.
	accessible := filterAccessible(pm.Allowed())
	if len(accessible) == 0 && len(dirs) > 0 {
		slog.Warn("none of the specified directories are accessible, waiting for MCP roots")
	}

	filesystem.Register(pm)
	system.Register()

	mcpServer := mcplib.NewMCPServer(config.Platform, Version, logger)

	for _, t := range registry.Global.List() {
		t.Register(mcpServer.MCPServer())
	}

	// Log resource.
	mcpServer.MCPServer().AddResource(&mcp.Resource{
		Name:        "Logs",
		URI:         "filesystem://logs",
		Description: "[DIRECTIVE: Audit Streaming] Server logs natively. Keywords: debug, errors, daemon-logs, traces, faults",
		MIMEType:    "text/plain",
	}, mcplib.HardenedResourceHandler(func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      req.Params.URI,
					Text:     telemetry.GlobalTracker.RingBuffer.String(),
					MIMEType: "text/plain",
				},
			},
		}, nil
	}))

	errChan := make(chan error, 1)
	go func() {
		if err := mcpServer.Serve(ctx, writer, reader); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("context cancelled; initiating graceful shutdown")
	case err := <-errChan:
		if mcplib.IsExpectedShutdownErr(err) {
			slog.Info("stdio transport closed gracefully", "reason", err.Error())
			return nil
		}
		return fmt.Errorf("server error: %w", err)
	}

	// 🛡️ DRAIN: Allow 2 seconds for clean transport closure before process exit.
	drainCtx, drainCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer drainCancel()
	<-drainCtx.Done()
	slog.Info("shutdown drain period complete")
	return nil
}

// filterAccessible returns only the directories that exist and are accessible.
func filterAccessible(dirs []string) []string {
	var result []string
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			slog.Warn("cannot access directory, skipping", "dir", dir, "error", err)
			continue
		}
		if !info.IsDir() {
			slog.Warn("path is not a directory, skipping", "dir", dir)
			continue
		}
		result = append(result, dir)
	}
	return result
}
