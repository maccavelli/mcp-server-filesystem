package main

// Tests for the MCP filesystem server entrypoint, covering shutdown error
// classification, auto-flushing writer behaviour, and EOF detection.

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/maccavelli/mcplib"
)

func TestIsExpectedShutdownErr(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"EOF", io.EOF, true},
		{"unexpected EOF", io.ErrUnexpectedEOF, true},
		{"broken pipe", fmt.Errorf("broken pipe"), true},
		{"connection reset", fmt.Errorf("connection reset by peer"), true},
		{"use of closed", fmt.Errorf("use of closed network connection"), true},
		{"bad file descriptor", fmt.Errorf("bad file descriptor"), true},
		{"client is closing", fmt.Errorf("client is closing"), true},
		{"connection closed", fmt.Errorf("connection closed"), true},
		{"file already closed", fmt.Errorf("file already closed"), true},
		{"random error", fmt.Errorf("random error"), false},
		{"wrapped EOF", fmt.Errorf("wrapped: %w", io.EOF), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mcplib.IsExpectedShutdownErr(tc.err)
			if got != tc.want {
				t.Errorf("IsExpectedShutdownErr(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestRun_FS(t *testing.T) {
	exitFunc = func(int) {}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	reader := strings.NewReader("")
	var writer bytes.Buffer
	errChan := make(chan error, 1)

	go func() {
		dir := t.TempDir()
		os.Args = []string{"cmd", dir}
		err := run(ctx, []string{dir}, nil, mcplib.NopReadCloser{Reader: reader}, mcplib.NopWriteCloser{Writer: &writer})
		errChan <- err
		close(errChan)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-errChan:
	case <-time.After(1 * time.Second):
	}
}

func TestMain_Version(t *testing.T) {
	os.Args = []string{"cmd", "-version"}
	exited := false
	exitFunc = func(_ int) {
		exited = true
	}
	main()
	if !exited {
		t.Error("main should have exited")
	}
}
