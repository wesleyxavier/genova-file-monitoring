package analyzer

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"time"

	"github.com/wesleyxavier/genova-file-monitoring/internal/store"
)

// BestPracticesGuide é o guia fixo de boas práticas embutido na v1 (sem RAG).
const BestPracticesGuide = `Você é um revisor de código estático local. Analise o arquivo fornecido e identifique apenas problemas reais e objetivos segundo estas diretrizes:
1. Segurança (critica): credenciais hardcoded, SQL injection, command injection, path traversal, exposição de segredos.
2. Bugs prováveis (critica/alta): nil/null dereference evidente, race condition clara, recursos não fechados, tratamento de erro ignorado.
3. Arquitetura e clareza (media): funções excessivamente longas/complexas, responsabilidades misturadas, duplicação evidente.
4. Estilo e higiene (baixa): logs de debug esquecidos, código morto evidente.
Não invente problemas se o código estiver limpo; retorne findings vazio.`

var (
	secretPattern = regexp.MustCompile(`(?i)(api[_-]?key|secret|password|passwd|token)\s*[:=]\s*["'][^"']{8,}["']`)
	awsKeyPattern = regexp.MustCompile(`AKIA[0-9A-Z]{16}`)
)

// RunDeterministic executa regras determinísticas antes do LLM (Princípio IX).
func RunDeterministic(relPath, content string) []store.Finding {
	var findings []store.Finding
	lines := strings.Split(content, "\n")
	now := time.Now().UTC().Format(time.RFC3339)

	for i, rawLine := range lines {
		lineNum := i + 1
		line := strings.TrimSpace(rawLine)

		if strings.HasPrefix(line, "<<<<<<<") || strings.HasPrefix(line, ">>>>>>>") {
			findings = append(findings, store.Finding{
				ID:          newFindingID(relPath, "merge-conflict-marker", lineNum),
				FilePath:    relPath,
				RuleID:      "merge-conflict-marker",
				Severity:    store.SeverityCritica,
				LineStart:   lineNum,
				SnippetHash: HashSnippet(line),
				Message:     "Marcador de conflito de merge encontrado no código.",
				Source:      "deterministic",
				CreatedAt:   now,
			})
			continue
		}

		if awsKeyPattern.MatchString(line) || secretPattern.MatchString(line) {
			findings = append(findings, store.Finding{
				ID:          newFindingID(relPath, "hardcoded-secret", lineNum),
				FilePath:    relPath,
				RuleID:      "hardcoded-secret",
				Severity:    store.SeverityCritica,
				LineStart:   lineNum,
				SnippetHash: HashSnippet(line),
				Message:     "Possível credencial ou segredo fixo detectado no código-fonte.",
				Source:      "deterministic",
				CreatedAt:   now,
			})
			continue
		}

		if strings.Contains(line, "console.log(") || strings.Contains(line, "fmt.Println(") {
			findings = append(findings, store.Finding{
				ID:          newFindingID(relPath, "no-debug-log", lineNum),
				FilePath:    relPath,
				RuleID:      "no-debug-log",
				Severity:    store.SeverityBaixa,
				LineStart:   lineNum,
				SnippetHash: HashSnippet(line),
				Message:     "Log de depuração encontrado; prefira logger estruturado ou remova antes do commit.",
				Source:      "deterministic",
				CreatedAt:   now,
			})
		}
	}

	return findings
}

// HashSnippet gera um hash SHA-256 curto do trecho para deduplicação de 30 minutos.
func HashSnippet(snippet string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(snippet)))
	return hex.EncodeToString(sum[:8])
}

func newFindingID(filePath, ruleID string, line int) string {
	raw := filePath + "|" + ruleID + "|" + time.Now().UTC().Format(time.RFC3339Nano) + "|" + string(rune(line))
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:12])
}
