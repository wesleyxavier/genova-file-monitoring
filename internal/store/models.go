package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Severity representa os quatro níveis de severidade de um achado.
type Severity string

const (
	SeverityCritica Severity = "critica"
	SeverityAlta    Severity = "alta"
	SeverityMedia   Severity = "media"
	SeverityBaixa   Severity = "baixa"
)

// IsValid informa se a severidade pertence ao conjunto permitido.
func (s Severity) IsValid() bool {
	switch s {
	case SeverityCritica, SeverityAlta, SeverityMedia, SeverityBaixa:
		return true
	default:
		return false
	}
}

// Rank retorna a precedência numérica da severidade (maior = mais grave).
func (s Severity) Rank() int {
	switch s {
	case SeverityCritica:
		return 4
	case SeverityAlta:
		return 3
	case SeverityMedia:
		return 2
	case SeverityBaixa:
		return 1
	default:
		return 0
	}
}

// MonitoringConfig armazena as preferências ativas de monitoramento.
type MonitoringConfig struct {
	ID            string   `json:"id,omitempty"`
	WatchPath     string   `json:"watch_path"`
	MinSeverity   Severity `json:"min_severity"`
	SilencedRules []string `json:"silenced_rules"`
	Extensions    []string `json:"extensions"`
	DebounceMS    int      `json:"debounce_ms"`
	UpdatedAt     string   `json:"updated_at,omitempty"`
}

// DefaultConfig retorna a configuração inicial padrão para a v1.
func DefaultConfig() MonitoringConfig {
	return MonitoringConfig{
		ID:            "default",
		WatchPath:     "",
		MinSeverity:   SeverityAlta,
		SilencedRules: []string{},
		Extensions:    []string{".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".cs", ".java", ".rs", ".sql"},
		DebounceMS:    500,
		UpdatedAt:     time.Now().UTC().Format(time.RFC3339),
	}
}

// ValidateConfig valida os campos de MonitoringConfig conforme o data-model.
func ValidateConfig(cfg MonitoringConfig) error {
	if !cfg.MinSeverity.IsValid() {
		return fmt.Errorf("severidade mínima inválida: %q", cfg.MinSeverity)
	}
	if cfg.DebounceMS < 100 || cfg.DebounceMS > 10000 {
		return fmt.Errorf("debounce_ms deve estar entre 100 e 10000 (recebido: %d)", cfg.DebounceMS)
	}
	if cfg.WatchPath != "" {
		info, err := os.Stat(cfg.WatchPath)
		if err != nil {
			return fmt.Errorf("watch_path inacessível: %w", err)
		}
		if !info.IsDir() {
			return errors.New("watch_path deve ser um diretório")
		}
	}
	for _, ext := range cfg.Extensions {
		if !strings.HasPrefix(ext, ".") || len(ext) < 2 {
			return fmt.Errorf("extensão inválida %q: deve iniciar com '.'", ext)
		}
	}
	return nil
}

// Finding representa um achado validado persistido no histórico.
type Finding struct {
	ID                string   `json:"id"`
	FilePath          string   `json:"file_path"`
	RuleID            string   `json:"rule_id"`
	Severity          Severity `json:"severity"`
	LineStart         int      `json:"line_start"`
	SnippetHash       string   `json:"snippet_hash,omitempty"`
	Message           string   `json:"message"`
	Source            string   `json:"source"`
	Notified          bool     `json:"notified"`
	SuppressionReason string   `json:"suppression_reason"`
	CreatedAt         string   `json:"created_at"`
}

// FindingFilter define filtros para consulta ao histórico de achados.
type FindingFilter struct {
	Severity Severity
	File     string
	Limit    int
}

// Store define o contrato de persistência de configurações e histórico.
type Store interface {
	Ping(ctx context.Context) error
	LoadConfig(ctx context.Context) (MonitoringConfig, error)
	SaveConfig(ctx context.Context, cfg MonitoringConfig) error
	SaveFinding(ctx context.Context, f Finding) error
	ListFindings(ctx context.Context, filter FindingFilter) ([]Finding, error)
	Close() error
}
