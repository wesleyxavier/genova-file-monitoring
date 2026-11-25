# Quickstart & Validation Guide: Monitor Local de Boas Práticas (v1)

**Feature**: `001-local-file-monitor`

## Pré-requisitos

1. **Go** conforme `go.mod` (`1.27.4`).
2. **Docker** com o serviço FalkorDB ativo:
   - Copie `.env.example` para `.env` e preencha `FALKORDB_USER`, `FALKORDB_PASSWORD` e a variável de comando do `llama-server` (ex.: `LLAMA_SERVER_CMD`).
   - Suba o FalkorDB:

     ```powershell
     docker compose --env-file .env up -d
     ```

3. **llama-server** (llama.cpp) e modelo `gemma-4-12B-it-qat-q4_0.gguf` disponíveis localmente (iniciados automaticamente pelo app Go na faixa de porta `11440–11450`).

## Verificação Automatizada (Build & Testes)

```powershell
go build ./...
go test ./...
```

## Cenários de Validação Ponta a Ponta

1. **Inicialização e Interface Local (`127.0.0.1`)**:
   - Execute `go run ./cmd/genova-file-monitoring`.
   - Confirme no log que o `llama-server` subiu em `127.0.0.1:11440` (ou outra porta `11440–11450`, nunca `11435`) e o FalkorDB conectou.
   - Acesse a UI local em `http://127.0.0.1:<porta_web>` e defina uma pasta de teste em `watch_path`.

2. **Detecção no Save & Somente Leitura**:
   - Salve um arquivo `.go` ou `.ts` na pasta observada contendo um problema Crítico (ex.: marcador de conflito de merge ou segredo hardcoded).
   - Verifique que o toast nativo do Windows aparece após a janela de debounce (`500ms`) + análise, e que o arquivo observado não teve nenhuma modificação de timestamp/conteúdo pelo monitor.

3. **Controle de Ruído e Histórico no FalkorDB**:
   - Salve um arquivo com achado de severidade `baixa` (ou dispare mais de 5 achados em 1 minuto).
   - Confirme que não houve excesso de notificações (>5/min) nem popup para severidade `baixa`.
   - Consulte a aba de histórico na UI `127.0.0.1` (ou `GET /api/findings`) e confirme que todos os achados foram persistidos no FalkorDB.

4. **Encerramento Limpo**:
   - Encerre o aplicativo (`Ctrl+C`) e confirme que o processo filho `llama-server` foi encerrado.
