package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wesleyxavier/genova-file-monitoring/internal/store"
)

type fakeStore struct {
	cfg      store.MonitoringConfig
	findings []store.Finding
}

func (f *fakeStore) Ping(ctx context.Context) error { return nil }
func (f *fakeStore) Close() error                   { return nil }
func (f *fakeStore) LoadConfig(ctx context.Context) (store.MonitoringConfig, error) {
	return f.cfg, nil
}
func (f *fakeStore) SaveConfig(ctx context.Context, cfg store.MonitoringConfig) error {
	f.cfg = cfg
	return nil
}
func (f *fakeStore) SaveFinding(ctx context.Context, item store.Finding) error {
	f.findings = append(f.findings, item)
	return nil
}
func (f *fakeStore) ListFindings(ctx context.Context, filter store.FindingFilter) ([]store.Finding, error) {
	return f.findings, nil
}

func TestServerBindAndEndpoints(t *testing.T) {
	fs := &fakeStore{
		cfg: store.DefaultConfig(),
		findings: []store.Finding{
			{ID: "1", FilePath: "main.go", RuleID: "no-debug-log", Severity: store.SeverityBaixa, Message: "debug"},
		},
	}

	hub := NewActivityHub()
	hub.Publish("clean", "main.go", "Análise concluída: nenhum problema relevante encontrado.")

	srv := NewServer(18080, fs, nil, hub, nil)
	if !strings.HasPrefix(srv.Addr(), "127.0.0.1:") {
		t.Fatalf("bind deve ser estritamente 127.0.0.1, obteve %s", srv.Addr())
	}

	// GET /api/config
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/config esperava 200, obteve %d", rec.Code)
	}

	// PUT /api/config válido
	cfg := store.DefaultConfig()
	cfg.WatchPath = t.TempDir()
	cfg.MinSeverity = store.SeverityMedia
	body, _ := json.Marshal(cfg)
	reqPut := httptest.NewRequest(http.MethodPut, "/api/config", bytes.NewReader(body))
	recPut := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recPut, reqPut)
	if recPut.Code != http.StatusOK || fs.cfg.MinSeverity != store.SeverityMedia {
		t.Fatalf("PUT /api/config falhou: code=%d body=%s", recPut.Code, recPut.Body.String())
	}

	// GET /api/findings
	reqFind := httptest.NewRequest(http.MethodGet, "/api/findings", nil)
	recFind := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recFind, reqFind)
	if recFind.Code != http.StatusOK || !strings.Contains(recFind.Body.String(), "no-debug-log") {
		t.Fatalf("GET /api/findings inesperado: code=%d body=%s", recFind.Code, recFind.Body.String())
	}

	// GET /api/activity
	reqAct := httptest.NewRequest(http.MethodGet, "/api/activity", nil)
	recAct := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recAct, reqAct)
	if recAct.Code != http.StatusOK || !strings.Contains(recAct.Body.String(), "main.go") {
		t.Fatalf("GET /api/activity inesperado: code=%d body=%s", recAct.Code, recAct.Body.String())
	}
}
