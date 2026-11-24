# genova-file-monitoring

Monitor de Boas Práticas: app local para Windows, escrito em Go, que observa uma pasta de projeto, analisa cada arquivo salvo com um modelo local e avisa no Windows quando encontra problemas.

## Objetivo
Dar feedback de qualidade no momento do save, quando a correção custa menos. A análise usa um guia de boas práticas e a documentação cadastrada pelo usuário, armazenados no FalkorDB (índice vetorial) e recuperados a cada análise. Uso pessoal, sem publicação.

## Stack
- Go (módulo `github.com/wesleyxavier/genova-file-monitoring`)
- fsnotify, beeep
- FalkorDB (Docker)
- LLM local e modelo de embedding: A definir

## Como executar
1. Copie `.env.example` para `.env.development` e preencha `FALKORDB_USER` e `FALKORDB_PASSWORD`.
2. Suba o FalkorDB: `docker compose --env-file .env.development up -d` (Redis em `6379`, Browser em `http://localhost:3900`).
3. Execute o app: `go run ./cmd/genova-file-monitoring`.

Mais detalhes no PRD (fora do repositório). Fluxo de especificação: Spec Kit (`.specify/`).

## Regras de domínio
| Severidade | Critério | Notifica? |
| --- | --- | --- |
| Crítica | Bug provável ou risco de segurança | Sim, com som |
| Alta | Violação clara de regra ativa | Sim |
| Média | Nomes pouco claros, responsabilidades misturadas | Se o mínimo configurado for média |
| Baixa | Estilo e preferências | Não; só no histórico |

Controle de ruído: teto de 5 avisos por minuto; mesmo arquivo e regra não avisam de novo por 30 minutos (salvo trecho alterado); regras podem ser silenciadas. O app nunca altera os arquivos observados.
