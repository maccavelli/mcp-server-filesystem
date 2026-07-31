package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListDirectoryWithSizes(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "file1.txt"), []byte("hello"), 0644)
	os.Mkdir(filepath.Join(tmp, "dir1"), 0755)

	entries, err := ListDirectoryWithSizes(tmp, "size")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	for _, entry := range entries {
		switch entry.Name {
		case "file1.txt":
			if entry.IsDirectory {
				t.Error("expected IsDirectory false")
			}
			if entry.Size != 5 {
				t.Errorf("expected size 5, got %d", entry.Size)
			}
		case "dir1":
			if !entry.IsDirectory {
				t.Error("expected IsDirectory true")
			}
		default:
			t.Errorf("unexpected entry %s", entry.Name)
		}
	}
}

func TestTreeToJSON(t *testing.T) {
	tree := []*TreeEntry{
		{
			Name: "dir1",
			Type: "directory",
			Children: []*TreeEntry{
				{
					Name: "file.txt",
					Type: "file",
				},
			},
		},
	}

	jsonStr, err := TreeToJSON(tree)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if jsonStr == "" {
		t.Error("expected non-empty JSON string")
	}

	if len(jsonStr) < 20 {
		t.Errorf("JSON string too short: %s", jsonStr)
	}
}

func TestMIMETypeExtra(t *testing.T) {
	if MIMEType(".png") != "image/png" {
		t.Error("expected image/png for .png")
	}
	if MIMEType(".unknown") != "application/octet-stream" {
		t.Error("expected application/octet-stream for .unknown")
	}
}

func TestNormalizeLineEndings(t *testing.T) {
	input := "line1\r\nline2\r\nline3\nline4"
	expected := "line1\nline2\nline3\nline4"

	if normalizeLineEndings(input) != expected {
		t.Errorf("expected %q, got %q", expected, normalizeLineEndings(input))
	}
}
