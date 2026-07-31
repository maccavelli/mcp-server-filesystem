package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadWriteFileContentExtra(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "test.txt")

	err := WriteFileContent(file, "hello")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	content, err := ReadFileContent(file)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if content != "hello" {
		t.Errorf("expected hello, got %s", content)
	}

	err = WriteFileContent(file, "world")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	content, err = ReadFileContent(file)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if content != "world" {
		t.Errorf("expected world, got %s", content)
	}

	_, err = ReadFileContent("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestApplyFileEditsExtra(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "edits.txt")

	os.WriteFile(file, []byte("line1\nline2\nline3\n"), 0644)

	edits := []FileEdit{
		{
			OldText: "line2",
			NewText: "line2_edited",
		},
	}

	res, err := ApplyFileEdits(file, edits, false)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if res == "" {
		t.Error("expected result")
	}

	b, _ := os.ReadFile(file)
	if string(b) != "line1\nline2_edited\nline3\n" {
		t.Errorf("unexpected content: %s", string(b))
	}
}
