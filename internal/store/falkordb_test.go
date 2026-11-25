package store

import (
	"context"
	"strings"
	"testing"
)

type fakeGraphClient struct {
	queries []string
	rows    [][]any
}

func (f *fakeGraphClient) Ping(ctx context.Context) error { return nil }
func (f *fakeGraphClient) Close() error                   { return nil }
func (f *fakeGraphClient) Query(ctx context.Context, graph, cypher string) ([][]any, error) {
	f.queries = append(f.queries, cypher)
	return f.rows, nil
}

func TestValidateConfig(t *testing.T) {
	cfg := DefaultConfig()
	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("config padrão deveria ser válida: %v", err)
	}

	invalidSev := cfg
	invalidSev.MinSeverity = Severity("invalida")
	if err := ValidateConfig(invalidSev); err == nil {
		t.Fatal("esperava erro para severidade inválida")
	}

	invalidDebounce := cfg
	invalidDebounce.DebounceMS = 50
	if err := ValidateConfig(invalidDebounce); err == nil {
		t.Fatal("esperava erro para debounce_ms < 100")
	}
}

func TestFalkorStoreSaveAndLoadConfig(t *testing.T) {
	fake := &fakeGraphClient{}
	s := NewFalkorStoreWithClient(fake, "genova_monitor")

	cfg := DefaultConfig()
	cfg.WatchPath = t.TempDir()
	cfg.SilencedRules = []string{"no-debug-log"}

	if err := s.SaveConfig(context.Background(), cfg); err != nil {
		t.Fatalf("erro ao salvar config: %v", err)
	}
	if len(fake.queries) != 1 || !strings.Contains(fake.queries[0], "MERGE (c:MonitoringConfig") {
		t.Fatalf("query inesperada: %v", fake.queries)
	}
	if !strings.Contains(fake.queries[0], `\\`) {
		t.Fatalf("esperava escape de barra invertida do Windows no Cypher: %s", fake.queries[0])
	}

	fake.rows = [][]any{{
		cfg.WatchPath,
		"media",
		`["no-debug-log"]`,
		`[".go"]`,
		int64(750),
		"2026-11-24T18:00:00Z",
	}}

	loaded, err := s.LoadConfig(context.Background())
	if err != nil {
		t.Fatalf("erro ao carregar config: %v", err)
	}
	if loaded.MinSeverity != SeverityMedia || loaded.DebounceMS != 750 || len(loaded.Extensions) != 1 {
		t.Fatalf("config carregada incorreta: %+v", loaded)
	}
}

func TestFalkorStoreListFindingsFilters(t *testing.T) {
	fake := &fakeGraphClient{
		rows: [][]any{{
			"id-1", "main.go", "hardcoded-secret", "critica", int64(10), "hash", "msg", "deterministic", true, "", "2026-11-24T18:00:00Z",
		}},
	}
	s := NewFalkorStoreWithClient(fake, "genova_monitor")

	items, err := s.ListFindings(context.Background(), FindingFilter{
		Severity: SeverityCritica,
		File:     "main.go",
		Limit:    500,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(items) != 1 || items[0].RuleID != "hardcoded-secret" || !items[0].Notified {
		t.Fatalf("resultado inesperado: %+v", items)
	}
	q := fake.queries[0]
	if !strings.Contains(q, "LIMIT 200") || !strings.Contains(q, "f.severity = 'critica'") {
		t.Fatalf("query com filtros/limite incorretos: %s", q)
	}
}
