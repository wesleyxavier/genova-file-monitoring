---
description: "Lista de tarefas ordenada por dependências para 001-local-file-monitor"
---

# Tasks: Monitor Local de Boas Práticas (v1)

**Input**: Design documents from `/specs/001-local-file-monitor/`
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/http-local-api.md`, `contracts/llm-analysis-schema.md`, `quickstart.md`

**Organization**: Tarefas organizadas por fase e User Story (`US1`, `US2`, `US3`, `US4`) para entrega incremental e testável de forma independente, incluindo as seções obrigatórias do fluxo Spec Kit (`## 0. Branch`, commits por grupo com autorização explícita, atualização de `README.md`, `graphify update .` e PR).

## 0. Branch

- [X] T001 Criar worktree para a feature `001-local-file-monitor` a partir de `develop` em `worktrees/001-local-file-monitor` (skill `openspec-worktree`)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Inicialização de dependências Go fixadas no `go.mod` e configuração de ambiente local.

- [X] T002 Adicionar e fixar versões no `go.mod` para `github.com/fsnotify/fsnotify`, `github.com/gen2brain/beeep` e `github.com/redis/go-redis/v9`
- [X] T003 [P] Atualizar `.env.example` com as variáveis `FALKORDB_USER`, `FALKORDB_PASSWORD`, `FALKORDB_ADDR` (padrão `127.0.0.1:6379`), `LLAMA_SERVER_CMD` e `WEB_PORT` (sem segredos reais)

### 1.1 Commit Group 1

- [ ] T004 Solicitar autorização explícita e executar `git commit` + `git push` da base do Setup

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Modelos compartilhados, carregamento de `.env` e cliente FalkorDB usados por todas as histórias.

- [X] T005 [P] Criar tipos compartilhados de domínio (`Severity`: `"critica"|"alta"|"media"|"baixa"`, `Finding`, `MonitoringConfig`) e validação de campos em `internal/store/models.go`
- [X] T006 [P] Implementar carregador de variáveis `.env` e validação de configuração de boot (incluindo faixa de porta `11440–11450`, nunca `11435`) em `internal/analyzer/env.go`
- [X] T007 Implementar cliente FalkorDB (`GRAPH.QUERY` no grafo `genova_monitor`) com `LoadConfig`, `SaveConfig`, `SaveFinding` e `ListFindings` em `internal/store/falkordb.go`
- [X] T008 [P] Escrever testes unitários de validação de modelos e queries/filtros com mock/fake em `internal/store/falkordb_test.go`

### 2.1 Commit Group 2

- [ ] T009 Solicitar autorização explícita e executar `git commit` + `git push` da camada Foundational

---

## Phase 3: User Story 1 - Monitorar pasta e notificar problemas críticos/altos ao salvar (Priority: P1) 🎯 MVP

**Goal**: Observar pasta configurada em modo somente leitura, aplicar debounce de 500 ms, rodar fila sequencial, analisar com regras determinísticas + LLM local e notificar no Windows.

**Independent Test**: Iniciar o monitor apontando para uma pasta de teste, salvar um arquivo elegível com problema Crítico/Alto e verificar que o toast nativo do Windows é exibido sem alterar nenhum arquivo observado.

- [X] T010 [P] [US1] Implementar matcher de `.gitignore`, ignore padrão (`.git`, `node_modules`, `bin`, `obj`, `dist`, `build`, `vendor`), detecção de binários/vazios e filtro por extensões configuráveis em `internal/watcher/filter.go`
- [X] T011 [US1] Implementar watcher recursivo somente leitura com `fsnotify` (incluindo novos subdiretórios) em `internal/watcher/watcher.go`
- [X] T012 [P] [US1] Implementar debounce por arquivo (padrão `500ms`, limites `100..10000ms`) e fila sequencial de 1 job por vez em `internal/queue/queue.go`
- [X] T013 [P] [US1] Embutir o guia fixo de boas práticas e implementar regras determinísticas pré-LLM (conflitos de merge, segredos óbvios, logs de debug) em `internal/analyzer/deterministic.go`
- [X] T014 [US1] Implementar cliente HTTP loopback para `/v1/chat/completions` com isolamento de prompt (`<<<BEGIN_UNTRUSTED_FILE_CONTENT>>>`) e validação estrita do schema JSON em `internal/analyzer/llm_client.go`
- [X] T015 [P] [US1] Implementar despachante de notificações nativas Windows via `beeep` (`Alert` para `critica`, `Notify` para `alta`/`media`) em `internal/notify/windows.go`
- [X] T016 [P] [US1] Escrever testes unitários para filtro de watcher, debounce/fila sequencial e regras determinísticas em `internal/watcher/filter_test.go`, `internal/queue/queue_test.go` e `internal/analyzer/deterministic_test.go`

### 3.1 Commit Group 3

- [ ] T017 Solicitar autorização explícita e executar `git commit` + `git push` da User Story 1 (MVP)

---

## Phase 4: User Story 2 - Controlar ruído e evitar fadiga de notificações (Priority: P2)

**Goal**: Aplicar teto de 5 avisos/minuto, deduplicação de 30 minutos por `arquivo + regra + snippet_hash`, severidade mínima configurável e regras silenciadas, persistindo tudo no FalkorDB.

**Independent Test**: Disparar mais de 5 achados em 1 minuto e repetir o mesmo save sem alterar o trecho; confirmar teto de 5 toasts, supressão por 30 min (liberando se trecho mudar) e registro completo no FalkorDB.

- [X] T018 [US2] Implementar avaliador de ruído (`silenced_rules`, `min_severity`, deduplicação 30 min por `file_path + rule_id + snippet_hash`, teto 5/min) em `internal/notify/policy.go`
- [X] T019 [US2] Integrar avaliador de ruído com persistência de `:Finding` (`notified` e `suppression_reason`) no FalkorDB em `internal/notify/service.go`
- [X] T020 [P] [US2] Escrever testes unitários cobrindo teto 5/min, janela 30 min, mudança de trecho, severidade mínima e silenciamento em `internal/notify/policy_test.go`

### 4.1 Commit Group 4

- [ ] T021 Solicitar autorização explícita e executar `git commit` + `git push` da User Story 2

---

## Phase 5: User Story 3 - Ciclo de vida automático do analisador local e validação estrita (Priority: P2)

**Goal**: Iniciar `llama-server` como processo filho a partir de `.env`, aguardar `/health`, garantir faixa `11440–11450` (nunca `11435`), encerrar no shutdown e descartar JSON inválido.

**Independent Test**: Subir e encerrar o app verificando processo filho `llama-server` na porta configurada, e validar via mock/teste que JSON malformado é descartado e logado sem notificar.

- [X] T022 [US3] Implementar gerenciador de processo filho `llama-server` (start com `os/exec`, validação de porta `11440–11450` != `11435`, polling `/health` e shutdown limpo) em `internal/analyzer/process.go`
- [X] T023 [US3] Orquestrar ciclo de vida completo (env, store, `llama-server`, queue, watcher, notify, graceful shutdown por sinal) em `cmd/genova-file-monitoring/main.go`
- [X] T024 [P] [US3] Escrever testes unitários para validação de porta, descarte de JSON inválido/campos ausentes e tratamento de falha do LLM em `internal/analyzer/llm_client_test.go` e `internal/analyzer/process_test.go`

### 5.1 Commit Group 5

- [ ] T025 Solicitar autorização explícita e executar `git commit` + `git push` da User Story 3

---

## Phase 6: User Story 4 - Configurar monitoramento e consultar histórico via interface local (Priority: P3)

**Goal**: Expor servidor HTTP e UI web embarcada exclusivamente em `127.0.0.1` para configurar pasta observada, severidade mínima, regras silenciadas, extensões e consultar/filtrar o histórico no FalkorDB.

**Independent Test**: Abrir `http://127.0.0.1:<WEB_PORT>`, alterar configurações, salvar arquivo com achado Baixo, recarregar a página e confirmar persistência no FalkorDB.

- [X] T026 [US4] Implementar servidor HTTP com bind estrito em `127.0.0.1` e handlers `/api/health`, `/api/config` (`GET`/`PUT` com hot-reload do watcher/notify) e `/api/findings` (`GET` com filtros) em `internal/web/server.go`
- [X] T027 [P] [US4] Criar interface HTML/JS embarcada (`embed.FS`) para status, formulário de configuração e tabela filtrável de histórico de achados em `internal/web/static/index.html`
- [X] T028 [US4] Conectar servidor web `127.0.0.1` ao bootstrap e graceful shutdown em `cmd/genova-file-monitoring/main.go`
- [X] T029 [P] [US4] Escrever testes de contrato HTTP (`httptest`) para `/api/health`, `/api/config` e `/api/findings` (incluindo rejeição de bind fora de `127.0.0.1`) em `internal/web/server_test.go`

### 6.1 Commit Group 6

- [ ] T030 Solicitar autorização explícita e executar `git commit` + `git push` da User Story 4

---

## Phase 7: Execute Tests & Pre-PR Checklist

**Purpose**: Validação completa, documentação e sincronização do grafo de conhecimento.

- [X] T031 Executar `go build ./...` e `go test ./...` garantindo 100% de aprovação
- [ ] T032 Validar cenários ponta a ponta do `specs/001-local-file-monitor/quickstart.md`
- [X] T033 Atualizar `README.md` documentando variáveis de `.env`, interface local `127.0.0.1`, uso do FalkorDB para configurações/histórico e comando de execução
- [X] T034 Executar `graphify update .` para sincronizar `graphify-out/`

---

## Phase 8: Create PR

- [ ] T035 Solicitar autorização explícita do usuário para commit final + push na branch `001-local-file-monitor` e criação do PR para `develop` via MCP

---

## Dependencies & Execution Order

- **Phase 1 (Setup)** → **Phase 2 (Foundational)** → **Phase 3 (US1 - MVP)** → **Phase 4 (US2)** & **Phase 5 (US3)** & **Phase 6 (US4)** → **Phase 7 (Testes & Pre-PR)** → **Phase 8 (Create PR)**.
- Tarefas marcadas com `[P]` atuam em arquivos distintos e podem ser executadas em paralelo dentro da mesma fase.
- **MVP Sugerido**: Concluir Phases 1, 2 e 3 (`T001`–`T017`) já entrega detecção no save em modo somente leitura com notificação no Windows.
