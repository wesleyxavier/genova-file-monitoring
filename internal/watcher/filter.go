package watcher

import (
	"bufio"
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const MaxAnalyzableBytes = 256 * 1024

var defaultIgnoredDirs = map[string]struct{}{
	".git":         {},
	"node_modules": {},
	"bin":          {},
	"obj":          {},
	"dist":         {},
	"build":        {},
	"vendor":       {},
	"worktrees":    {},
}

// Filter decide quais diretórios e arquivos devem ser observados e analisados.
type Filter struct {
	root       string
	extensions map[string]struct{}
	gitRules   []gitIgnoreRule
}

type gitIgnoreRule struct {
	pattern string
	dirOnly bool
	hasPath bool
}

// NewFilter cria um filtro carregando extensões permitidas e o .gitignore da raiz (se existir).
func NewFilter(root string, extensions []string) *Filter {
	extSet := make(map[string]struct{}, len(extensions))
	for _, ext := range extensions {
		e := strings.ToLower(strings.TrimSpace(ext))
		if e != "" {
			extSet[e] = struct{}{}
		}
	}
	f := &Filter{
		root:       filepath.Clean(root),
		extensions: extSet,
	}
	f.loadGitIgnore()
	return f
}

func (f *Filter) loadGitIgnore() {
	giPath := filepath.Join(f.root, ".gitignore")
	file, err := os.Open(giPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		dirOnly := strings.HasSuffix(line, "/")
		line = strings.TrimSuffix(line, "/")
		line = strings.TrimPrefix(line, "/")
		if line == "" {
			continue
		}
		f.gitRules = append(f.gitRules, gitIgnoreRule{
			pattern: line,
			dirOnly: dirOnly,
			hasPath: strings.Contains(line, "/"),
		})
	}
}

// ShouldIgnoreDir informa se o diretório deve ser ignorado pelo watcher.
func (f *Filter) ShouldIgnoreDir(fullPath string) bool {
	base := filepath.Base(fullPath)
	if _, ok := defaultIgnoredDirs[strings.ToLower(base)]; ok {
		return true
	}
	rel, err := filepath.Rel(f.root, fullPath)
	if err != nil || rel == "." {
		return false
	}
	relSlash := filepath.ToSlash(rel)
	return f.matchesGitIgnore(relSlash, base, true)
}

// ShouldAnalyzePath verifica extensão e regras de ignore de caminho sem ler o conteúdo.
func (f *Filter) ShouldAnalyzePath(fullPath string) bool {
	ext := strings.ToLower(filepath.Ext(fullPath))
	if _, ok := f.extensions[ext]; !ok {
		return false
	}
	rel, err := filepath.Rel(f.root, fullPath)
	if err != nil {
		return false
	}
	relSlash := filepath.ToSlash(rel)
	parts := strings.Split(relSlash, "/")
	for i := 0; i < len(parts)-1; i++ {
		if _, ok := defaultIgnoredDirs[strings.ToLower(parts[i])]; ok {
			return false
		}
	}
	return !f.matchesGitIgnore(relSlash, filepath.Base(fullPath), false)
}

// IsEligibleContent verifica se o conteúdo não é vazio, não excede o limite e não é binário.
func IsEligibleContent(data []byte) bool {
	if len(data) == 0 || len(data) > MaxAnalyzableBytes {
		return false
	}
	sample := data
	if len(sample) > 8000 {
		sample = sample[:8000]
	}
	return !bytes.Contains(sample, []byte{0})
}

func (f *Filter) matchesGitIgnore(relSlash, base string, isDir bool) bool {
	for _, r := range f.gitRules {
		if r.dirOnly && !isDir {
			continue
		}
		target := base
		if r.hasPath {
			target = relSlash
		}
		if ok, _ := path.Match(r.pattern, target); ok {
			return true
		}
		// Se uma pasta pai bate com regra dirOnly, ignora arquivos internos também.
		if !isDir && strings.HasPrefix(relSlash, r.pattern+"/") {
			return true
		}
	}
	return false
}
