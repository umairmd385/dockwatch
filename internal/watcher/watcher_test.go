package watcher

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateHash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	err := os.WriteFile(testFile, []byte("hello world"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	hash1, err := CalculateHash(testFile)
	if err != nil {
		t.Fatalf("CalculateHash failed: %v", err)
	}

	err = os.WriteFile(testFile, []byte("hello world 2"), 0644)
	if err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}

	hash2, err := CalculateHash(testFile)
	if err != nil {
		t.Fatalf("CalculateHash failed: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("Expected hashes to differ but both were %s", hash1)
	}
}
