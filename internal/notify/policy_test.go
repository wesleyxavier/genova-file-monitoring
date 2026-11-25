package notify

import (
	"testing"
	"time"

	"github.com/wesleyxavier/genova-file-monitoring/internal/store"
)

func TestPolicyRateLimitDedupSeverityAndSilence(t *testing.T) {
	now := time.Date(2026, 11, 24, 18, 0, 0, 0, time.UTC)
	p := NewPolicy()
	p.nowFunc = func() time.Time { return now }

	cfg := store.DefaultConfig()
	cfg.MinSeverity = store.SeverityAlta
	cfg.SilencedRules = []string{"silenced-rule"}

	// Regra silenciada
	if ok, reason := p.Evaluate(store.Finding{FilePath: "a.go", RuleID: "silenced-rule", Severity: store.SeverityCritica}, cfg); ok || reason != "silenced_rule" {
		t.Fatalf("esperava silenced_rule, obteve ok=%v reason=%s", ok, reason)
	}

	// Abaixo do mínimo (media < alta) e baixa
	if ok, reason := p.Evaluate(store.Finding{FilePath: "a.go", RuleID: "r1", Severity: store.SeverityMedia}, cfg); ok || reason != "below_min_severity" {
		t.Fatalf("esperava below_min_severity, obteve ok=%v reason=%s", ok, reason)
	}

	// Deduplicação por 30m com mesmo snippetHash
	f1 := store.Finding{FilePath: "a.go", RuleID: "r-dedup", Severity: store.SeverityAlta, SnippetHash: "h1"}
	if ok, _ := p.Evaluate(f1, cfg); !ok {
		t.Fatal("primeira ocorrência deveria notificar")
	}
	now = now.Add(5 * time.Minute)
	if ok, reason := p.Evaluate(f1, cfg); ok || reason != "deduplicated" {
		t.Fatalf("esperava deduplicated, obteve ok=%v reason=%s", ok, reason)
	}
	// Trecho alterado libera notificação antes de 30m
	f1.SnippetHash = "h2"
	if ok, reason := p.Evaluate(f1, cfg); !ok {
		t.Fatalf("esperava notificação após mudança de trecho, obteve reason=%s", reason)
	}

	// Teto de 5 por minuto (já temos 2 na janela recente; avançar 2 min para limpar)
	now = now.Add(2 * time.Minute)
	for i := 0; i < 5; i++ {
		f := store.Finding{FilePath: "b.go", RuleID: string(rune('a' + i)), Severity: store.SeverityCritica, SnippetHash: "x"}
		if ok, reason := p.Evaluate(f, cfg); !ok {
			t.Fatalf("item %d deveria notificar, reason=%s", i, reason)
		}
	}
	extra := store.Finding{FilePath: "b.go", RuleID: "extra", Severity: store.SeverityCritica, SnippetHash: "x"}
	if ok, reason := p.Evaluate(extra, cfg); ok || reason != "rate_limit" {
		t.Fatalf("6º item no minuto deveria sofrer rate_limit, obteve ok=%v reason=%s", ok, reason)
	}
}
