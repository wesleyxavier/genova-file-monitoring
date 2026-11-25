# Implementation Plan: Monitor Local de Boas Práticas (v1)

**Branch**: `001-local-file-monitor` | **Date**: 2026-11-24 | **Spec**: [spec.md](file:///c:/Users/User/Documents/TI/Lavi/Web/Genova/src/PROJETOS/GenovaFileMonitoring/specs/001-local-file-monitor/spec.md)

**Input**: Feature specification from `/specs/001-local-file-monitor/spec.md`

## Summary

Implementar a v1 do `genova-file-monitoring` em Go para Windows 10/11: um monitor de arquivos somente leitura com debounce de 500 ms e fila sequencial em segundo plano, pipeline de análise em duas etapas (regras determinísticas antes do LLM local `llama-server` gerenciado como processo filho na porta `11440–11450` com guia de boas práticas embutido e validação estrita de schema JSON), notificações nativas Windows via `beeep` com controle de ruído (teto 5/min, deduplicação 30 min por arquivo+regra+hash do trecho, severidade mínima e regras silenciadas), persistência de configurações e histórico de achados no FalkorDB local e interface web mínima escutando exclusivamente em `127.0.0.1`.

## Technical Context

**Language/Version**: Go `1.27.4` (conforme `go.mod`)
**Primary Dependencies**:

- `github.com/fsnotify/fsnotify` (watcher de arquivos Windows)
- `github.com/gen2brain/beeep` (notificações nativas Windows com/sem som)
- `github.com/redis/go-redis/v9` (cliente RESP para comandos `GRAPH.QUERY` no FalkorDB)
- Stdlib Go (`net/http`, `os/exec`, `encoding/json`, `embed`, `crypto/sha256`, `sync`)

**Storage**: FalkorDB local via Docker (`docker-compose.yml`, porta `6379`) apenas para `:MonitoringConfig` e `:Finding` (sem RAG/embeddings na v1).
**Testing**: `go test ./...` (testes unitários e de integração com mocks/fakes para LLM e FalkorDB, e testes reais de watcher/debounce/rate-limiter).
**Target Platform**: Windows 10 / Windows 11 (`127.0.0.1` loopback apenas).
**Project Type**: CLI / Background desktop service com UI web local embarcada (`127.0.0.1`).
**Performance Goals**: Debounce padrão de `500 ms`; fila sequencial (1 job por vez) sem bloquear a IDE; resposta de notificações logo após inferência local.
**Constraints**:

- Somente leitura na pasta observada (zero escritas/movimentações/exclusões).
- Zero chamadas de rede externas (apenas `127.0.0.1`).
- `llama-server` obrigatoriamente na faixa `11440–11450` (nunca `11435`).
- Teto de 5 notificações/minuto e deduplicação de 30 minutos.
- Segredos apenas em `.env*` ignorado pelo git.

**Scale/Scope**: Uso pessoal local, 1 pasta de projeto monitorada por vez, repositório único Go.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Princípio | Status | Verificação |
| --- | --- | --- |
| I. Local-First e Privacidade | PASS | Código observado não sai da máquina; apenas loopback (`127.0.0.1` / `localhost:6379`). |
| II. Somente Leitura | PASS | `internal/watcher` e `internal/analyzer` usam apenas leitura (`os.ReadFile`, `os.Stat`). |
| III. Interface Local Mínima | PASS | `internal/web` faz bind estrito em `127.0.0.1:<port>`, rejeitando `0.0.0.0`. |
| IV. O Modelo Sugere, o App Decide | PASS | Schema JSON validado em `internal/analyzer`; prompt separa instruções do sistema e dados do arquivo. |
| V. Análise Direta com LLM, sem RAG na v1 | PASS | Emenda `1.1.0` autoriza FalkorDB na v1 apenas para configurações e histórico; RAG/embeddings/ingestão fora da v1 (`internal/ingest` intocado). |
| VI. LLM Gerenciado pelo App | PASS | App inicia `llama-server` como processo filho via `.env`, porta `11440–11450` (nunca `11435`), e encerra ao sair. |
| VII. Ruído Controlado | PASS | Teto 5/min, deduplicação 30 min por arquivo+regra+trecho, severidade mínima e silenciamento em `internal/notify`. |
| VIII. Não Bloqueia a IDE | PASS | Debounce 500 ms + fila sequencial 1 job por vez em `internal/queue`. |
| IX. Determinismo antes do Modelo | PASS | `internal/analyzer` executa regras determinísticas antes de chamar `llama-server`. |
| X. Dependências e Segredos | PASS | Stdlib preferida; deps externas fixadas no `go.mod`; `.env` fora do git. |
| XI. Fluxo de Trabalho | PASS | Seguindo Spec Kit (`/speckit.specify` → `/speckit.plan` → `/speckit.tasks` → worktree). |

*Re-check pós-design (Phase 1)*: Todos os artefatos (`research.md`, `data-model.md`, `contracts/*`, `quickstart.md`) permanecem 100% conformes com a Constituição `1.1.0`. Nenhuma violação detectada.

## Project Structure

### Documentation (this feature)

```text
specs/001-local-file-monitor/
├── plan.md              # Este arquivo
├── research.md          # Decisões técnicas (Phase 0)
├── data-model.md        # Entidades FalkorDB e memória (Phase 1)
├── quickstart.md        # Guia de execução e validação (Phase 1)
├── contracts/
│   ├── http-local-api.md
│   └── llm-analysis-schema.md
└── tasks.md             # Será gerado por /speckit.tasks
```

### Source Code (repository root)

```text
cmd/
└── genova-file-monitoring/
    └── main.go              # Bootstrap, carregamento .env, wiring e graceful shutdown

internal/
├── analyzer/                # Processo filho llama-server, guia fixo, regras determinísticas, cliente LLM e validação JSON
├── ingest/                  # Reservado para fase futura (não usado na v1)
├── notify/                  # beeep Windows toast, rate limit 5/min, deduplicação 30m e filtro de severidade
├── queue/                   # Debounce por arquivo (500ms) e fila sequencial (1 job/vez)
├── store/                   # Cliente FalkorDB (configurações + histórico de achados)
├── watcher/                 # fsnotify recursivo, filtros de extensão, ignore padrão e .gitignore
└── web/                     # Servidor HTTP 127.0.0.1 + UI embarcada (embed.FS)
```

**Structure Decision**: Manter a estrutura modular Go já iniciada em `cmd/genova-file-monitoring` e `internal/*`, mantendo `internal/ingest` inativo na v1 conforme Constituição V.

## Complexity Tracking

Nenhuma violação à Constituição; tabela não aplicável.
