package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultGraphName = "genova_monitor"

// GraphClient abstrai a execução de comandos Redis/FalkorDB para facilitar testes.
type GraphClient interface {
	Ping(ctx context.Context) error
	Query(ctx context.Context, graph, cypher string) ([][]any, error)
	Close() error
}

type redisGraphClient struct {
	rdb *redis.Client
}

func (c *redisGraphClient) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *redisGraphClient) Query(ctx context.Context, graph, cypher string) ([][]any, error) {
	res, err := c.rdb.Do(ctx, "GRAPH.QUERY", graph, cypher).Result()
	if err != nil {
		return nil, err
	}
	top, ok := res.([]any)
	if !ok || len(top) < 2 {
		return nil, nil
	}
	rows, ok := top[1].([]any)
	if !ok {
		return nil, nil
	}
	out := make([][]any, 0, len(rows))
	for _, r := range rows {
		if cols, ok := r.([]any); ok {
			out = append(out, cols)
		}
	}
	return out, nil
}

func (c *redisGraphClient) Close() error {
	return c.rdb.Close()
}

// FalkorStore implementa Store usando FalkorDB via protocolo RESP.
type FalkorStore struct {
	client    GraphClient
	graphName string
}

// NewFalkorStore cria um cliente FalkorDB conectado ao endereço informado.
func NewFalkorStore(addr, username, password string) *FalkorStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
	})
	return NewFalkorStoreWithClient(&redisGraphClient{rdb: rdb}, defaultGraphName)
}

// NewFalkorStoreWithClient permite injetar um GraphClient customizado (útil em testes).
func NewFalkorStoreWithClient(client GraphClient, graphName string) *FalkorStore {
	if graphName == "" {
		graphName = defaultGraphName
	}
	return &FalkorStore{client: client, graphName: graphName}
}

func (s *FalkorStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx)
}

func (s *FalkorStore) Close() error {
	return s.client.Close()
}

// LoadConfig carrega a configuração singleton; se não existir, retorna DefaultConfig.
func (s *FalkorStore) LoadConfig(ctx context.Context) (MonitoringConfig, error) {
	cypher := `MATCH (c:MonitoringConfig {id: 'default'}) RETURN c.watch_path, c.min_severity, c.silenced_rules, c.extensions, c.debounce_ms, c.updated_at`
	rows, err := s.client.Query(ctx, s.graphName, cypher)
	if err != nil {
		return MonitoringConfig{}, err
	}
	if len(rows) == 0 || len(rows[0]) < 6 {
		return DefaultConfig(), nil
	}

	row := rows[0]
	cfg := DefaultConfig()
	cfg.WatchPath = asString(row[0])
	if sev := Severity(asString(row[1])); sev.IsValid() {
		cfg.MinSeverity = sev
	}
	if raw := asString(row[2]); raw != "" {
		var rules []string
		if err := json.Unmarshal([]byte(raw), &rules); err == nil {
			cfg.SilencedRules = rules
		}
	}
	if raw := asString(row[3]); raw != "" {
		var exts []string
		if err := json.Unmarshal([]byte(raw), &exts); err == nil && len(exts) > 0 {
			cfg.Extensions = exts
		}
	}
	if d := asInt(row[4]); d >= 100 && d <= 10000 {
		cfg.DebounceMS = d
	}
	cfg.UpdatedAt = asString(row[5])
	return cfg, nil
}

// SaveConfig valida e persiste a configuração singleton no FalkorDB.
func (s *FalkorStore) SaveConfig(ctx context.Context, cfg MonitoringConfig) error {
	if cfg.ID == "" {
		cfg.ID = "default"
	}
	if cfg.SilencedRules == nil {
		cfg.SilencedRules = []string{}
	}
	if err := ValidateConfig(cfg); err != nil {
		return err
	}
	cfg.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	silencedJSON, err := json.Marshal(cfg.SilencedRules)
	if err != nil {
		return err
	}
	extsJSON, err := json.Marshal(cfg.Extensions)
	if err != nil {
		return err
	}

	cypher := fmt.Sprintf(
		`MERGE (c:MonitoringConfig {id: 'default'}) SET c.watch_path = %s, c.min_severity = %s, c.silenced_rules = %s, c.extensions = %s, c.debounce_ms = %d, c.updated_at = %s`,
		cypherQuote(cfg.WatchPath),
		cypherQuote(string(cfg.MinSeverity)),
		cypherQuote(string(silencedJSON)),
		cypherQuote(string(extsJSON)),
		cfg.DebounceMS,
		cypherQuote(cfg.UpdatedAt),
	)
	_, err = s.client.Query(ctx, s.graphName, cypher)
	return err
}

// SaveFinding persiste um achado no grafo do FalkorDB.
func (s *FalkorStore) SaveFinding(ctx context.Context, f Finding) error {
	if f.CreatedAt == "" {
		f.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	cypher := fmt.Sprintf(
		`CREATE (:Finding {id: %s, file_path: %s, rule_id: %s, severity: %s, line_start: %d, snippet_hash: %s, message: %s, source: %s, notified: %t, suppression_reason: %s, created_at: %s})`,
		cypherQuote(f.ID),
		cypherQuote(f.FilePath),
		cypherQuote(f.RuleID),
		cypherQuote(string(f.Severity)),
		f.LineStart,
		cypherQuote(f.SnippetHash),
		cypherQuote(f.Message),
		cypherQuote(f.Source),
		f.Notified,
		cypherQuote(f.SuppressionReason),
		cypherQuote(f.CreatedAt),
	)
	_, err := s.client.Query(ctx, s.graphName, cypher)
	return err
}

// ListFindings consulta o histórico de achados aplicando filtros opcionais.
func (s *FalkorStore) ListFindings(ctx context.Context, filter FindingFilter) ([]Finding, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	var where []string
	if filter.Severity != "" && filter.Severity.IsValid() {
		where = append(where, fmt.Sprintf("f.severity = %s", cypherQuote(string(filter.Severity))))
	}
	if filter.File != "" {
		where = append(where, fmt.Sprintf("f.file_path CONTAINS %s", cypherQuote(filter.File)))
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ") + " "
	}

	cypher := fmt.Sprintf(
		`MATCH (f:Finding) %sRETURN f.id, f.file_path, f.rule_id, f.severity, f.line_start, f.snippet_hash, f.message, f.source, f.notified, f.suppression_reason, f.created_at ORDER BY f.created_at DESC LIMIT %d`,
		whereClause,
		limit,
	)

	rows, err := s.client.Query(ctx, s.graphName, cypher)
	if err != nil {
		return nil, err
	}

	findings := make([]Finding, 0, len(rows))
	for _, row := range rows {
		if len(row) < 11 {
			continue
		}
		findings = append(findings, Finding{
			ID:                asString(row[0]),
			FilePath:          asString(row[1]),
			RuleID:            asString(row[2]),
			Severity:          Severity(asString(row[3])),
			LineStart:         asInt(row[4]),
			SnippetHash:       asString(row[5]),
			Message:           asString(row[6]),
			Source:            asString(row[7]),
			Notified:          asBool(row[8]),
			SuppressionReason: asString(row[9]),
			CreatedAt:         asString(row[10]),
		})
	}
	return findings, nil
}

func cypherQuote(s string) string {
	escaped := strings.ReplaceAll(s, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	escaped = strings.ReplaceAll(escaped, "\r", `\r`)
	escaped = strings.ReplaceAll(escaped, "\n", `\n`)
	return "'" + escaped + "'"
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

func asInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case string:
		i, _ := strconv.Atoi(n)
		return i
	default:
		return 0
	}
}

func asBool(v any) bool {
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return strings.EqualFold(b, "true")
	default:
		return false
	}
}
