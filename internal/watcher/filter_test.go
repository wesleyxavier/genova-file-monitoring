package watcher

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilterExtensionsAndIgnores(t *testing.T) {
	dir := t.TempDir()
	gi := "ignored.go\ncustom_out/\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gi), 0644); err != nil {
		t.Fatal(err)
	}

	f := NewFilter(dir, []string{".go", ".ts"})

	if !f.ShouldAnalyzePath(filepath.Join(dir, "main.go")) {
		t.Fatal("main.go deveria ser analisado")
	}
	if f.ShouldAnalyzePath(filepath.Join(dir, "README.md")) {
		t.Fatal("README.md não está nas extensões permitidas")
	}
	if f.ShouldAnalyzePath(filepath.Join(dir, "ignored.go")) {
		t.Fatal("ignored.go deveria ser ignorado pelo .gitignore")
	}
	if f.ShouldAnalyzePath(filepath.Join(dir, "node_modules", "pkg", "index.ts")) {
		t.Fatal("node_modules deveria ser ignorado por padrão")
	}
	if !f.ShouldIgnoreDir(filepath.Join(dir, "custom_out")) {
		t.Fatal("custom_out deveria ser ignorado pelo .gitignore")
	}
}

func TestIsEligibleContent(t *testing.T) {
	if IsEligibleContent(nil) {
		t.Fatal("arquivo vazio não deve ser elegível")
	}
	if IsEligibleContent([]byte("abc\x00def")) {
		t.Fatal("arquivo binário não deve ser elegível")
	}
	if !IsEligibleContent([]byte("package main\nfunc main() {}\n")) {
		t.Fatal("arquivo Go válido deve ser elegível")
	}
}
