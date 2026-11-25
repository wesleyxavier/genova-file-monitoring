package analyzer

import "testing"

func TestValidateLlamaPort(t *testing.T) {
	if err := ValidateLlamaPort(11440); err != nil {
		t.Fatalf("11440 deveria ser válida: %v", err)
	}
	if err := ValidateLlamaPort(11450); err != nil {
		t.Fatalf("11450 deveria ser válida: %v", err)
	}
	if err := ValidateLlamaPort(11435); err == nil {
		t.Fatal("11435 deve ser rejeitada")
	}
	if err := ValidateLlamaPort(8080); err == nil {
		t.Fatal("8080 deve ser rejeitada por estar fora da faixa 11440-11450")
	}
}

func TestParseAndValidateLlamaPort(t *testing.T) {
	port, err := ParseAndValidateLlamaPort(`llama-server -m "C:\Models\gemma.gguf" --port 11442`)
	if err != nil || port != 11442 {
		t.Fatalf("esperava porta 11442, obteve %d (err=%v)", port, err)
	}

	if _, err := ParseAndValidateLlamaPort(`llama-server --port 11435`); err == nil {
		t.Fatal("esperava erro ao configurar porta proibida 11435")
	}
}
