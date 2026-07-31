// Package engine provides core filesystem operations including file reading,
// writing, editing with diff generation, directory traversal, and search.
package engine

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/maccavelli/mcp-server-filesystem/internal/config"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

const (
	defaultFilePerm = 0o644
	mimeImagePNG    = "image/png"
	mimeOctetStream = "application/octet-stream"
)

// FormatSize returns a human-readable byte size string.
func FormatSize(sizeBytes int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	if sizeBytes == 0 {
		return "0 B"
	}
	size := float64(sizeBytes)
	idx := 0
	for size >= 1024 && idx < len(units)-1 {
		size /= 1024
		idx++
	}
	if idx == 0 {
		return fmt.Sprintf("%d %s", sizeBytes, units[0])
	}
	return fmt.Sprintf("%.2f %s", size, units[idx])
}

// ReadFileContent reads an entire file as a UTF-8 string.
// Files larger than config.MaxReadFileSize are rejected to prevent OOM.
func ReadFileContent(filePath string) (string, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		slog.Error("stat file", "error", err)
		return "", err
	}
	if info.Size() > config.MaxReadFileSize {
		return "", fmt.Errorf("file too large (%s, limit %s): %s",
			FormatSize(info.Size()), FormatSize(config.MaxReadFileSize), filePath)
	}
	data, err := os.ReadFile(filePath) //nolint:gosec // path validated against allowed directories upstream
	if err != nil {
		slog.Error("reading file", "error", err)
		return "", err
	}
	return string(data), nil
}

// ReadFileBase64 reads a file and returns its base64-encoded content.
func ReadFileBase64(filePath string) (string, error) {
	f, err := os.Open(filePath) //nolint:gosec // path validated against allowed directories upstream
	if err != nil {
		slog.Error("opening file", "error", err)
		return "", err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			slog.Error("closing file", "error", closeErr)
		}
	}()

	var buf strings.Builder
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	if _, err := io.Copy(encoder, f); err != nil {
		slog.Error("encoding file", "error", err)
		return "", err
	}
	if err := encoder.Close(); err != nil {
		slog.Error("closing encoder", "error", err)
		return "", err
	}
	return buf.String(), nil
}

// WriteFileContent writes content to a file using atomic rename to prevent
// race conditions and symlink attacks.
func WriteFileContent(filePath, content string) error {
	// Resolve symlink so we replace the target, not the symlink itself.
	if realPath, err := filepath.EvalSymlinks(filePath); err == nil {
		filePath = realPath
	}

	// Try exclusive creation first.
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, defaultFilePerm) //nolint:gosec // path validated upstream; user-facing files are group-readable by design
	if err == nil {
		n, writeErr := f.WriteString(content)
		if writeErr == nil && n != len(content) {
			writeErr = io.ErrShortWrite
		}
		if writeErr == nil {
			writeErr = f.Sync()
		}
		closeErr := f.Close()
		if writeErr != nil {
			return fmt.Errorf("writing new file: %w", writeErr)
		}
		return closeErr
	}

	if !os.IsExist(err) {
		slog.Error("creating file", "error", err)
		return err
	}

	// File exists — use atomic temp+rename.
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		slog.Error("generating random suffix", "error", err)
		return err
	}
	tmpPath := filePath + "." + hex.EncodeToString(randBytes) + ".tmp"

	f, err = os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, defaultFilePerm) //nolint:gosec // path validated upstream; user-facing files are group-readable by design
	if err != nil {
		slog.Error("creating temp file", "error", err)
		return err
	}
	if _, err := f.WriteString(content); err != nil {
		if closeErr := f.Close(); closeErr != nil {
			slog.Error("closing temp file", "error", closeErr)
		}
		_ = os.Remove(tmpPath) //nolint:errcheck // best-effort temp cleanup
		slog.Error("writing temp file", "error", err)
		return err
	}
	if err := f.Sync(); err != nil {
		if closeErr := f.Close(); closeErr != nil {
			slog.Error("closing temp file", "error", closeErr)
		}
		_ = os.Remove(tmpPath) //nolint:errcheck // best-effort temp cleanup
		slog.Error("syncing temp file", "error", err)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath) //nolint:errcheck // best-effort temp cleanup
		slog.Error("closing temp file", "error", err)
		return err
	}

	if err := os.Rename(tmpPath, filePath); err != nil {
		_ = os.Remove(tmpPath) //nolint:errcheck // best-effort temp cleanup after failed rename
		slog.Error("atomic rename", "error", err)
		return err
	}
	return nil
}

// FileInfo holds metadata about a file or directory.
type FileInfo struct {
	Size        int64  `json:"size"`
	Created     string `json:"created"`
	Modified    string `json:"modified"`
	Accessed    string `json:"accessed"`
	IsDirectory bool   `json:"isDirectory"`
	IsFile      bool   `json:"isFile"`
	Permissions string `json:"permissions"`
}

// GetFileStats returns metadata for a path.
func GetFileStats(filePath string) (*FileInfo, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		slog.Error("stat", "error", err)
		return nil, err
	}
	return &FileInfo{
		Size:        info.Size(),
		Created:     info.ModTime().String(), // Go doesn't expose birth time portably
		Modified:    info.ModTime().String(),
		Accessed:    info.ModTime().String(),
		IsDirectory: info.IsDir(),
		IsFile:      info.Mode().IsRegular(),
		Permissions: fmt.Sprintf("%o", info.Mode().Perm()),
	}, nil
}

// HeadFile returns the first n lines of a file.
func HeadFile(filePath string, n int) (string, error) {
	f, err := os.Open(filePath) //nolint:gosec // path validated against allowed directories upstream
	if err != nil {
		slog.Error("opening file", "error", err)
		return "", err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			slog.Error("closing file", "error", closeErr)
		}
	}()

	scanner := bufio.NewScanner(f)
	lines := make([]string, 0, n)
	for scanner.Scan() && len(lines) < n {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		slog.Error("reading lines", "error", err)
		return "", err
	}
	return strings.Join(lines, "\n"), nil
}

// TailFile returns the last n lines of a file using reverse chunked reads.
func TailFile(filePath string, n int) (string, error) {
	f, err := os.Open(filePath) //nolint:gosec // path validated against allowed directories upstream
	if err != nil {
		slog.Error("opening file", "error", err)
		return "", err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			slog.Error("closing file", "error", closeErr)
		}
	}()

	info, err := f.Stat()
	if err != nil {
		slog.Error("stat", "error", err)
		return "", err
	}
	fileSize := info.Size()
	if fileSize == 0 {
		return "", nil
	}

	const chunkSize = 1024
	var lines []string
	position := fileSize
	var remaining []byte
	buf := make([]byte, chunkSize) // Reuse buffer across iterations

	for position > 0 && len(lines) < n {
		size := min(int64(chunkSize), position)
		position -= size

		_, err := f.ReadAt(buf[:size], position)
		if err != nil && err != io.EOF {
			slog.Error("reading chunk", "error", err)
			return "", err
		}

		chunkBytes := buf[:size:size]
		chunkBytes = append(chunkBytes, remaining...)
		chunkBytes = bytes.ReplaceAll(chunkBytes, []byte("\r\n"), []byte("\n"))
		chunkLines := bytes.Split(chunkBytes, []byte("\n"))

		if position > 0 {
			remaining = chunkLines[0]
			chunkLines = chunkLines[1:]
		} else {
			remaining = nil
		}

		for i := len(chunkLines) - 1; i >= 0 && len(lines) < n; i-- {
			lines = append(lines, string(chunkLines[i]))
		}
	}

	// If there's remaining text and we still need lines, prepend it.
	if len(remaining) > 0 && len(lines) < n {
		lines = append(lines, string(remaining))
	}

	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}

	return strings.Join(lines, "\n"), nil
}

// normalizeLineEndings converts \r\n to \n.
func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// FileEdit describes a single text replacement operation.
type FileEdit struct {
	OldText string `json:"oldText" jsonschema:"Exact block of text to replace. Must match the target file exactly including leading/trailing whitespace."`
	NewText string `json:"newText" jsonschema:"Replacement text block."`
}

// ApplyFileEdits applies a sequence of text replacements to a file and
// returns a unified diff. If dryRun is true, no changes are written.
func ApplyFileEdits(filePath string, edits []FileEdit, dryRun bool) (string, error) {
	raw, err := os.ReadFile(filePath) //nolint:gosec // path validated against allowed directories upstream
	if err != nil {
		slog.Error("reading file", "error", err)
		return "", err
	}
	content := normalizeLineEndings(string(raw))
	modified := content

	for _, edit := range edits {
		normalizedOld := normalizeLineEndings(edit.OldText)
		normalizedNew := normalizeLineEndings(edit.NewText)

		if strings.Contains(modified, normalizedOld) {
			modified = strings.Replace(modified, normalizedOld, normalizedNew, 1)
			continue
		}

		// Fallback: whitespace-flexible line matching.
		var matchFound bool
		modified, matchFound = applyFlexibleMatch(modified, normalizedOld, normalizedNew)
		if !matchFound {
			return "", fmt.Errorf("could not find exact match for edit:\n%s", edit.OldText)
		}
	}

	// Generate unified diff.
	editsResult := myers.ComputeEdits(span.URIFromPath(filePath), content, modified)
	diff := fmt.Sprint(gotextdiff.ToUnified(filePath, filePath, content, editsResult))

	if !dryRun {
		if err := atomicWrite(filePath, modified); err != nil {
			slog.Error("writing edits", "error", err)
			return "", err
		}
	}

	return diff, nil
}

// applyFlexibleMatch handles whitespace-flexible line matching and replacement.
func applyFlexibleMatch(modified, normalizedOld, normalizedNew string) (string, bool) {
	oldLines := strings.Split(normalizedOld, "\n")
	contentLines := strings.Split(modified, "\n")

	for i := 0; i <= len(contentLines)-len(oldLines); i++ {
		isMatch := true
		for j, oldLine := range oldLines {
			if strings.TrimSpace(oldLine) != strings.TrimSpace(contentLines[i+j]) {
				isMatch = false
				break
			}
		}
		if isMatch {
			newLines := strings.Split(normalizedNew, "\n")
			// Preserve original indentation.
			indent := ""
			if idx := strings.IndexFunc(contentLines[i], func(r rune) bool { return r != ' ' && r != '\t' }); idx > 0 {
				indent = contentLines[i][:idx]
			}
			for k, line := range newLines {
				if k == 0 {
					newLines[k] = indent + strings.TrimLeft(line, " \t")
				}
			}
			result := make([]string, 0, len(contentLines)-len(oldLines)+len(newLines))
			result = append(result, contentLines[:i]...)
			result = append(result, newLines...)
			result = append(result, contentLines[i+len(oldLines):]...)
			return strings.Join(result, "\n"), true
		}
	}
	return modified, false
}

// atomicWrite writes content via a temp file + rename.
func atomicWrite(filePath, content string) error {
	if realPath, err := filepath.EvalSymlinks(filePath); err == nil {
		filePath = realPath
	}
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return err
	}
	tmpPath := filePath + "." + hex.EncodeToString(randBytes) + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, defaultFilePerm) //nolint:gosec // path validated upstream; user-facing files are group-readable by design
	if err != nil {
		return err
	}
	if _, err := f.WriteString(content); err != nil {
		if closeErr := f.Close(); closeErr != nil {
			slog.Error("closing temp file", "error", closeErr)
		}
		_ = os.Remove(tmpPath) //nolint:errcheck // best-effort temp cleanup
		return err
	}
	if err := f.Sync(); err != nil {
		if closeErr := f.Close(); closeErr != nil {
			slog.Error("closing temp file", "error", closeErr)
		}
		_ = os.Remove(tmpPath) //nolint:errcheck // best-effort temp cleanup
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath) //nolint:errcheck // best-effort temp cleanup
		return err
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		_ = os.Remove(tmpPath) //nolint:errcheck // best-effort temp cleanup
		return err
	}
	return nil
}

// TreeEntry represents a node in a directory tree.
type TreeEntry struct {
	Name     string       `json:"name"`
	Type     string       `json:"type"` // "file" or "directory"
	Children []*TreeEntry `json:"children,omitempty,omitzero"`
}

// BuildDirectoryTree recursively builds a JSON-serializable tree.
// Recursion is capped at config.MaxTreeDepth to prevent stack overflow.
func BuildDirectoryTree(ctx context.Context, rootPath string, excludePatterns []string) ([]*TreeEntry, error) {
	return buildTree(ctx, rootPath, rootPath, excludePatterns, 0)
}

func buildTree(ctx context.Context, currentPath, rootPath string, excludePatterns []string, depth int) ([]*TreeEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if depth > config.MaxTreeDepth {
		return nil, nil // Silently stop recursion at max depth
	}

	entries, err := os.ReadDir(currentPath)
	if err != nil {
		slog.Error("reading directory", "error", err)
		return nil, err
	}

	var result []*TreeEntry
	for _, entry := range entries {
		relPath, err := filepath.Rel(rootPath, filepath.Join(currentPath, entry.Name()))
		if err != nil {
			continue
		}

		if shouldExclude(relPath, excludePatterns) {
			continue
		}

		node := &TreeEntry{
			Name: entry.Name(),
			Type: "file",
		}
		if entry.IsDir() {
			node.Type = "directory"
			children, err := buildTree(ctx, filepath.Join(currentPath, entry.Name()), rootPath, excludePatterns, depth+1)
			if err != nil {
				return nil, err
			}
			node.Children = children
		}
		result = append(result, node)
	}
	return result, nil
}

func shouldExclude(relPath string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, err := doublestar.Match(pattern, relPath); err == nil && matched {
			return true
		}
		if matched, err := doublestar.Match("**/"+pattern, relPath); err == nil && matched {
			return true
		}
		if matched, err := doublestar.Match("**/"+pattern+"/**", relPath); err == nil && matched {
			return true
		}
	}
	return false
}

// SearchFiles recursively searches for files matching a glob pattern.
// Results are capped at config.MaxSearchResults to prevent memory exhaustion.
func SearchFiles(ctx context.Context, rootPath, pattern string, excludePatterns []string) ([]string, error) {
	var results []string

	err := filepath.WalkDir(rootPath, func(fullPath string, d fs.DirEntry, walkErr error) error {
		if walkErr == nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			if len(results) >= config.MaxSearchResults {
				return filepath.SkipAll
			}

			relPath, relErr := filepath.Rel(rootPath, fullPath)
			if relErr == nil && relPath != "." {
				skip := false
				for _, ep := range excludePatterns {
					if matched, matchErr := doublestar.Match(ep, relPath); matchErr == nil && matched {
						if d.IsDir() {
							return filepath.SkipDir
						}
						skip = true
						break
					}
				}
				if !skip {
					if matched, matchErr := doublestar.Match(pattern, relPath); matchErr == nil && matched {
						results = append(results, fullPath)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		slog.Error("searching files", "error", err)
		return nil, err
	}
	return results, nil
}

// DirEntry holds details for list_directory_with_sizes.
type DirEntry struct {
	Name        string
	IsDirectory bool
	Size        int64
}

// ListDirectoryWithSizes lists entries in a directory with their sizes.
func ListDirectoryWithSizes(dirPath, sortBy string) ([]DirEntry, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		slog.Error("reading directory", "error", err)
		return nil, err
	}

	result := make([]DirEntry, 0, len(entries))
	for _, entry := range entries {
		de := DirEntry{
			Name:        entry.Name(),
			IsDirectory: entry.IsDir(),
		}
		info, err := entry.Info()
		if err == nil {
			de.Size = info.Size()
		}
		result = append(result, de)
	}

	switch sortBy {
	case "size":
		sort.Slice(result, func(i, j int) bool {
			return result[i].Size > result[j].Size
		})
	default: // "name"
		sort.Slice(result, func(i, j int) bool {
			return result[i].Name < result[j].Name
		})
	}
	return result, nil
}

// mimeTypes maps file extensions to MIME types. Hoisted to package level
// to avoid re-creating the map on every call.
var mimeTypes = map[string]string{
	".png":  mimeImagePNG,
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".bmp":  "image/bmp",
	".svg":  "image/svg+xml",
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".ogg":  "audio/ogg",
	".flac": "audio/flac",
}

// MIMEType returns the MIME type for a file extension.
func MIMEType(ext string) string {
	if mt, ok := mimeTypes[strings.ToLower(ext)]; ok {
		return mt
	}
	return mimeOctetStream
}

// TreeToJSON is a convenience that marshals a tree to indented JSON.
func TreeToJSON(tree []*TreeEntry) (string, error) {
	data, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		slog.Error("marshaling tree", "error", err)
		return "", err
	}
	return string(data), nil
}

// CopyPath copies a file or directory recursively from src to dst.
func CopyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		slog.Error("stat source", "error", err)
		return err
	}

	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src) //nolint:gosec // path validated against allowed directories upstream
	if err != nil {
		slog.Error("open source", "error", err)
		return err
	}
	defer func() {
		if closeErr := in.Close(); closeErr != nil {
			slog.Error("closing source file", "error", closeErr)
		}
	}()

	out, err := os.Create(dst) //nolint:gosec // path validated against allowed directories upstream
	if err != nil {
		slog.Error("create destination", "error", err)
		return err
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			slog.Error("closing destination file", "error", closeErr)
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		slog.Error("copy data", "error", err)
		return err
	}
	return out.Sync()
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o750); err != nil { //nolint:gosec // workspace directories use standard permissions
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// RemovePath forcefully removes a file or an entire directory recursively.
func RemovePath(path string) error {
	if err := os.RemoveAll(path); err != nil {
		slog.Error("remove path", "error", err)
		return err
	}
	return nil
}

// AppendFileContent appends a string to the end of a file.
// If the file does not exist, it creates it.
func AppendFileContent(filePath, content string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, defaultFilePerm) //nolint:gosec // path validated upstream; user-facing files are group-readable by design
	if err != nil {
		slog.Error("open file for append", "error", err)
		return err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			slog.Error("closing file", "error", closeErr)
		}
	}()

	if _, err := f.WriteString(content); err != nil {
		slog.Error("write append content", "error", err)
		return err
	}
	return f.Sync()
}

// GetFileHash computes the SHA-256 hash of a file.
func GetFileHash(filePath string) (string, error) {
	f, err := os.Open(filePath) //nolint:gosec // path validated against allowed directories upstream
	if err != nil {
		slog.Error("open file for hash", "error", err)
		return "", err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			slog.Error("closing file", "error", closeErr)
		}
	}()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		slog.Error("hash copy", "error", err)
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
