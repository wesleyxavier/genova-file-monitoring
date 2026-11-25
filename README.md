# genova-file-monitoring

Monitor de Boas Práticas: app local para Windows, escrito em Go, que observa uma pasta de projeto em modo somente leitura, analisa cada arquivo salvo (regras determinísticas + modelo local `llama-server`) e avisa no Windows quando encontra problemas.

## Objetivo

Dar feedback de qualidade no momento do save, quando a correção custa menos. Na v1, a análise usa regras determinísticas pré-LLM e o LLM local com um guia de boas práticas embutido como texto fixo (sem RAG, sem busca web). O FalkorDB é usado na v1 exclusivamente para persistência de configurações e histórico de achados; RAG e ingestão de documentação ficam para fase futura. Uso pessoal, sem publicação.

## Stack

- Go (`go.mod`: 1.27.4, módulo `github.com/wesleyxavier/genova-file-monitoring`)
- `fsnotify` (watcher recursivo), `beeep` (notificações Windows)
- `go-redis/v9` → FalkorDB (Docker) para configurações (`:MonitoringConfig`) e histórico de achados (`:Finding`)
- LLM local: `llama-server` (llama.cpp) gerenciado como processo filho pelo app Go na faixa `11440–11450` (nunca `11435`)
- Interface local: servidor HTTP + UI embarcada escutando apenas em `127.0.0.1` (`WEB_PORT`, padrão `18080`)
- Embedding / RAG: fora da v1

## Como executar

1. Copie `.env.example` para `.env` e preencha:
   - `FALKORDB_USER`, `FALKORDB_PASSWORD`, `FALKORDB_ADDR` (padrão `127.0.0.1:6379`)
   - `LLAMA_SERVER_CMD` (ex.: `llama-server -m C:\LLamaModels\gemma-4-12B-it-qat-q4_0.gguf -ngl 99 -fa on --cache-type-k q8_0 --cache-type-v q8_0 --port 11440`)
   - `WEB_PORT` (padrão `8080`)
2. Suba o FalkorDB:

   ```powershell
   docker compose --env-file .env up -d
   ```

3. Execute o app (ele inicia o `llama-server` automaticamente como processo filho e abre a UI local em `http://127.0.0.1:8080`):

   ```powershell
   go run ./cmd/genova-file-monitoring
   ```

## Regras de domínio

| Severidade | Critério | Notifica? |
| --- | --- | --- |
| Crítica | Bug provável ou risco de segurança | Sim, com som |
| Alta | Violação clara de regra ativa | Sim |
| Média | Nomes pouco claros, responsabilidades misturadas | Se o mínimo configurado for média |
| Baixa | Estilo e preferências | Não; só no histórico |

Controle de ruído: teto de 5 avisos por minuto; mesmo arquivo e regra não avisam de novo por 30 minutos (salvo trecho alterado); regras podem ser silenciadas via interface local `127.0.0.1`. O app nunca altera os arquivos observados.
