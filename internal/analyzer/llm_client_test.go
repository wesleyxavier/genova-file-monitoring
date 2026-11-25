package analyzer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestParseAndValidateLLMJSON(t *testing.T) {
	// JSON inválido deve ser descartado (retorna nil)
	if res := ParseAndValidateLLMJSON("a.go", "{invalid"); res != nil {
		t.Fatalf("esperava nil para JSON inválido, obteve %+v", res)
	}

	// Campo findings ausente deve ser descartado
	if res := ParseAndValidateLLMJSON("a.go", `{"other":[]}`); res != nil {
		t.Fatalf("esperava nil para findings ausente, obteve %+v", res)
	}

	// Item com severidade inválida é descartado, válido é mantido
	raw := `{"findings":[
		{"rule_id":"bad-sev","severity":"urgente","line_start":1,"message":"x"},
		{"rule_id":"nil-check","severity":"alta","line_start":12,"snippet":"x.Foo()","message":"Possível nil dereference."}
	]}`
	res := ParseAndValidateLLMJSON("a.go", raw)
	if len(res) != 1 || res[0].RuleID != "nil-check" || res[0].Severity != "alta" {
		t.Fatalf("esperava apenas 1 achado válido, obteve %+v", res)
	}
}

func TestLLMClientAnalyzeFileWithMockServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"findings\":[{\"rule_id\":\"err-ignored\",\"severity\":\"alta\",\"line_start\":5,\"message\":\"Erro ignorado.\"}]}"}}]}`))
	}))
	defer srv.Close()

	// Extrai porta do mock apenas para testar o fluxo HTTP via baseURL
	parts := strings.Split(srv.URL, ":")
	port, _ := strconv.Atoi(parts[len(parts)-1])
	client := &LLMClient{
		baseURL:    "http://127.0.0.1:" + strconv.Itoa(port),
		httpClient: srv.Client(),
	}

	findings, err := client.AnalyzeFile(context.Background(), "main.go", "package main")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(findings) != 1 || findings[0].RuleID != "err-ignored" {
		t.Fatalf("achados inesperados: %+v", findings)
	}
}
