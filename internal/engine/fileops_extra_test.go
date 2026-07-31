package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyPath(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	dst := filepath.Join(tmp, "dst.txt")

	os.WriteFile(src, []byte("hello copy"), 0644)

	err := CopyPath(src, dst)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	b, _ := os.ReadFile(dst)
	if string(b) != "hello copy" {
		t.Error("content mismatch")
	}

	srcDir := filepath.Join(tmp, "srcDir")
	dstDir := filepath.Join(tmp, "dstDir")
	os.Mkdir(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("hello dir"), 0644)

	err = CopyPath(srcDir, dstDir)
	if err != nil {
		t.Errorf("expected no error for dir copy, got %v", err)
	}

	b, _ = os.ReadFile(filepath.Join(dstDir, "file.txt"))
	if string(b) != "hello dir" {
		t.Error("content mismatch in dir copy")
	}

	err = CopyPath("nonexistent", filepath.Join(tmp, "nonexistent"))
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestAppendFileContent(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "append.txt")

	err := AppendFileContent(file, "hello")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = AppendFileContent(file, " world")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	b, _ := os.ReadFile(file)
	if string(b) != "hello world" {
		t.Error("content mismatch")
	}
}

func TestRemovePath(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "remove.txt")

	os.WriteFile(file, []byte("remove me"), 0644)

	err := RemovePath(file)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Error("expected file to be removed")
	}

	err = RemovePath("nonexistent")
	// os.RemoveAll returns nil if path does not exist, so it should be nil
	if err != nil {
		t.Errorf("expected no error for nonexistent file, got %v", err)
	}
}

func TestGetFileHash(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "hash.txt")

	os.WriteFile(file, []byte("hello hash"), 0644)

	hash, err := GetFileHash(file)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(hash) == 0 {
		t.Error("expected non-empty hash")
	}
}

func TestReadFileBase64(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "base64.txt")

	os.WriteFile(file, []byte("hello base64"), 0644)

	b64, err := ReadFileBase64(file)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if b64 != "aGVsbG8gYmFzZTY0" {
		t.Errorf("unexpected base64: %s", b64)
	}
}
