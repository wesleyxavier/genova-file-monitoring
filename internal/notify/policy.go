package notify

import (
	"sync"
	"time"

	"github.com/wesleyxavier/genova-file-monitoring/internal/store"
)

const (
	MaxPerMinute = 5
	DedupTTL     = 30 * time.Minute
)

type dedupEntry struct {
	snippetHash string
	expiresAt   time.Time
}

// Policy aplica controle de ruído: silenciamento, severidade mínima, deduplicação 30 min e teto 5/min.
type Policy struct {
	mu      sync.Mutex
	nowFunc func() time.Time
	recent  []time.Time
	dedup   map[string]dedupEntry
}

// NewPolicy cria um avaliador de ruído com relógio real.
func NewPolicy() *Policy {
	return &Policy{
		nowFunc: time.Now,
		dedup:   make(map[string]dedupEntry),
	}
}

// Evaluate decide se o achado deve notificar ou ser suprimido, retornando (notified, suppressionReason).
func (p *Policy) Evaluate(f store.Finding, cfg store.MonitoringConfig) (bool, string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, r := range cfg.SilencedRules {
		if r == f.RuleID {
			return false, "silenced_rule"
		}
	}

	// Baixa nunca notifica (apenas histórico); achados abaixo da severidade mínima também são suprimidos.
	if f.Severity == store.SeverityBaixa || f.Severity.Rank() < cfg.MinSeverity.Rank() {
		return false, "below_min_severity"
	}

	now := p.nowFunc()
	key := f.FilePath + "|" + f.RuleID
	if entry, ok := p.dedup[key]; ok && now.Before(entry.expiresAt) {
		if entry.snippetHash == f.SnippetHash {
			return false, "deduplicated"
		}
	}

	// Limpa timestamps fora da janela deslizante de 1 minuto.
	cutoff := now.Add(-1 * time.Minute)
	kept := p.recent[:0]
	for _, ts := range p.recent {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	p.recent = kept

	if len(p.recent) >= MaxPerMinute {
		return false, "rate_limit"
	}

	p.recent = append(p.recent, now)
	p.dedup[key] = dedupEntry{
		snippetHash: f.SnippetHash,
		expiresAt:   now.Add(DedupTTL),
	}
	return true, ""
}
