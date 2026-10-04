package cmd

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractBinaryStaysInDestDir(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "release.tar.gz")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	payload := []byte("binary")
	if err := tw.WriteHeader(&tar.Header{Name: "../../nginxpm-cli", Mode: 0o755, Size: int64(len(payload)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(payload); err != nil {
		t.Fatal(err)
	}
	for _, c := range []interface{ Close() error }{tw, gz, f} {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}

	destDir := filepath.Join(dir, "out")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := extractBinary(archivePath, destDir)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(destDir, "nginxpm"); got != want {
		t.Fatalf("extractBinary path = %q, want %q", got, want)
	}
	data, err := os.ReadFile(got)
	if err != nil || string(data) != string(payload) {
		t.Fatalf("extracted %q, %v; want %q", data, err, payload)
	}
}
