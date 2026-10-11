package targets

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "urls.txt")
	content := `#Production
		https://example.com
		https://go.dev
		# Another comment
		https://github.com
	`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	want := []string{
		"https://example.com",
		"https://go.dev",
		"https://github.com",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("LoadFile() = %v, want %v", got, want)
	}
}

func TestLoadFileMissingFile(t *testing.T) {
	_, err := LoadFile(filepath.Join(t.TempDir(), "missing.txt"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadFileEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.txt")

	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}

	got, err := LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("got %v, want no URLs", got)
	}
}
