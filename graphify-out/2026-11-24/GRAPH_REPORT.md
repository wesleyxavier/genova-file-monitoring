# Graph Report - 001-local-file-monitor  (2026-11-24)

## Corpus Check
- 82 files · ~65,270 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 631 nodes · 771 edges · 58 communities (56 shown, 2 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 43 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `49022261`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Tasks: [FEATURE NAME]
- Execution Steps
- common.ps1
- genova-file-monitoring — Instruções para Agentes de IA
- SKILL.md
- 4. Detection Passes (Token-Efficient Analysis)
- Feature Specification: [FEATURE NAME]
- SKILL.md
- SKILL.md
- SKILL.md
- Core Principles
- Core Principles
- Implementation Plan: [FEATURE]
- SKILL.md
- create-new-feature.ps1
- SKILL.md
- SKILL.md
- SKILL.md
- SKILL.md
- [CHECKLIST TYPE] Checklist: [FEATURE NAME]
- main.go
- github.com/wesleyxavier/genova-file-monitoring
- Execution Steps
- SKILL.md
- SKILL.md
- SKILL.md
- SKILL.md
- SKILL.md
- SKILL.md
- SKILL.md
- SKILL.md
- Server
- Watcher
- FalkorStore
- llm_client.go
- Tasks: Monitor Local de Boas Práticas (v1)
- Queue
- User Scenarios & Testing *(mandatory)*
- LlamaProcess
- Data Model: Monitor Local de Boas Práticas (v1)
- Implementation Plan: Monitor Local de Boas Práticas (v1)
- Endpoints JSON
- Research: Monitor Local de Boas Práticas (v1)
- Specification Quality Checklist: Monitor Local de Boas Práticas (v1)
- fakeGraphClient
- Contract: LLM Analysis Input & Output Schema
- Quickstart & Validation Guide: Monitor Local de Boas Práticas (v1)

## God Nodes (most connected - your core abstractions)
1. `Server` - 18 edges
2. `MonitoringConfig` - 17 edges
3. `appRuntime` - 14 edges
4. `Finding` - 13 edges
5. `Tasks: [FEATURE NAME]` - 13 edges
6. `Core Principles` - 12 edges
7. `Queue` - 11 edges
8. `Severity` - 11 edges
9. `Tasks: Monitor Local de Boas Práticas (v1)` - 11 edges
10. `main()` - 10 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `LoadEnvConfig()`  [INFERRED]
  cmd/genova-file-monitoring/main.go → internal/analyzer/env.go
- `main()` --calls--> `NewLLMClient()`  [INFERRED]
  cmd/genova-file-monitoring/main.go → internal/analyzer/llm_client.go
- `main()` --calls--> `NewLlamaProcess()`  [INFERRED]
  cmd/genova-file-monitoring/main.go → internal/analyzer/process.go
- `main()` --calls--> `NewFalkorStore()`  [INFERRED]
  cmd/genova-file-monitoring/main.go → internal/store/falkordb.go
- `appRuntime` --references--> `LlamaProcess`  [EXTRACTED]
  cmd/genova-file-monitoring/main.go → internal/analyzer/process.go

## Import Cycles
- None detected.

## Communities (58 total, 2 thin omitted)

### Community 0 - "Tasks: [FEATURE NAME]"
Cohesion: 0.07
Nodes (26): Dependencies & Execution Order, Format: `[ID] [P?] [Story] Description`, Implementation for User Story 1, Implementation for User Story 2, Implementation for User Story 3, Implementation Strategy, Incremental Delivery, MVP First (User Story 1 Only) (+18 more)

### Community 1 - "Execution Steps"
Cohesion: 0.12
Nodes (15): 1. Initialize Convergence Context, 2. Load Artifacts (Progressive Disclosure), 3. Build the Intent Inventory, 4. Assess the Codebase and Classify Findings, 5. Assign Severity, 6. Present the In-Session Findings Summary, 7. Append Convergence Tasks (or report converged), 8. Provide Next Actions (Handoff) (+7 more)

### Community 2 - "common.ps1"
Cohesion: 0.23
Nodes (13): Find-SpecifyRoot(), Format-SpecKitCommand(), Get-CurrentBranch(), Get-FeaturePathsEnv(), Get-InvokeSeparator(), Get-NormalizedPriority(), Get-Python3Command(), Get-RepoRoot() (+5 more)

### Community 3 - "genova-file-monitoring — Instruções para Agentes de IA"
Cohesion: 0.13
Nodes (13): Comandos, Convenções, Fluxo de especificação, genova-file-monitoring — Instruções para Agentes de IA, Propósito, Regras de operação, Restrições, Stack (+5 more)

### Community 4 - "SKILL.md"
Cohesion: 0.08
Nodes (25): 1. Initialize Analysis Context, 2. Load Artifacts (Progressive Disclosure), 3. Build Semantic Models, 4. Detection Passes (Token-Efficient Analysis), 5. Severity Assignment, 6. Produce Compact Analysis Report, 7. Provide Next Actions, 8. Offer Remediation (+17 more)

### Community 5 - "4. Detection Passes (Token-Efficient Analysis)"
Cohesion: 0.08
Nodes (25): 1. Initialize Analysis Context, 2. Load Artifacts (Progressive Disclosure), 3. Build Semantic Models, 4. Detection Passes (Token-Efficient Analysis), 5. Severity Assignment, 6. Produce Compact Analysis Report, 7. Provide Next Actions, 8. Offer Remediation (+17 more)

### Community 6 - "Feature Specification: [FEATURE NAME]"
Cohesion: 0.15
Nodes (12): Assumptions, Edge Cases, Feature Specification: [FEATURE NAME], Functional Requirements, Key Entities *(include if feature involves data)*, Measurable Outcomes, Requirements *(mandatory)*, Success Criteria *(mandatory)* (+4 more)

### Community 7 - "SKILL.md"
Cohesion: 0.18
Nodes (10): Completion Report, Done When, Key rules, Mandatory Post-Execution Hooks, Outline, Phase 0: Outline & Research, Phase 1: Design & Contracts, Phases (+2 more)

### Community 8 - "SKILL.md"
Cohesion: 0.18
Nodes (10): Completion Report, Done When, For AI Generation, Mandatory Post-Execution Hooks, Outline, Pre-Execution Checks, Quick Guidelines, Section Requirements (+2 more)

### Community 9 - "SKILL.md"
Cohesion: 0.18
Nodes (10): Checklist Format (REQUIRED), Completion Report, Done When, Mandatory Post-Execution Hooks, Outline, Phase Structure, Pre-Execution Checks, Task Generation Rules (+2 more)

### Community 10 - "Core Principles"
Cohesion: 0.12
Nodes (16): Core Principles, Fluxo de Desenvolvimento, genova-file-monitoring Constitution, Governance, I. Local-First e Privacidade, II. Somente Leitura, III. Interface Local Mínima, IV. O Modelo Sugere, o App Decide (+8 more)

### Community 11 - "Core Principles"
Cohesion: 0.18
Nodes (10): Core Principles, Governance, [PRINCIPLE_1_NAME], [PRINCIPLE_2_NAME], [PRINCIPLE_3_NAME], [PRINCIPLE_4_NAME], [PRINCIPLE_5_NAME], [PROJECT_NAME] Constitution (+2 more)

### Community 12 - "Implementation Plan: [FEATURE]"
Cohesion: 0.22
Nodes (8): Complexity Tracking, Constitution Check, Documentation (this feature), Implementation Plan: [FEATURE], Project Structure, Source Code (repository root), Summary, Technical Context

### Community 13 - "SKILL.md"
Cohesion: 0.25
Nodes (7): Anti-Examples: What NOT To Do, Checklist Purpose: "Unit Tests for English", Example Checklist Types & Sample Items, Execution Steps, Post-Execution Checks, Pre-Execution Checks, User Input

### Community 14 - "create-new-feature.ps1"
Cohesion: 0.39
Nodes (4): ConvertTo-AsciiLower(), ConvertTo-CleanBranchName(), ConvertTo-UnicodeWords(), Get-BranchName()

### Community 15 - "SKILL.md"
Cohesion: 0.29
Nodes (6): Completion Report, Done When, Mandatory Post-Execution Hooks, Outline, Pre-Execution Checks, User Input

### Community 16 - "SKILL.md"
Cohesion: 0.29
Nodes (6): Completion Report, Done When, Mandatory Post-Execution Hooks, Outline, Pre-Execution Checks, User Input

### Community 17 - "SKILL.md"
Cohesion: 0.33
Nodes (5): Outline, Post-Execution Checks, Pre-Execution Checks, Scope Guard, User Input

### Community 18 - "SKILL.md"
Cohesion: 0.33
Nodes (5): Deprecation Notice, Outline, Post-Execution Checks, Pre-Execution Checks, User Input

### Community 19 - "[CHECKLIST TYPE] Checklist: [FEATURE NAME]"
Cohesion: 0.40
Nodes (4): [Category 1], [Category 2], [CHECKLIST TYPE] Checklist: [FEATURE NAME], Notes

### Community 20 - "main.go"
Cohesion: 0.05
Nodes (40): CancelFunc, Context, Mutex, main(), appRuntime, Mutex, NewPolicy(), T (+32 more)

### Community 33 - "Execution Steps"
Cohesion: 0.12
Nodes (15): 1. Initialize Convergence Context, 2. Load Artifacts (Progressive Disclosure), 3. Build the Intent Inventory, 4. Assess the Codebase and Classify Findings, 5. Assign Severity, 6. Present the In-Session Findings Summary, 7. Append Convergence Tasks (or report converged), 8. Provide Next Actions (Handoff) (+7 more)

### Community 34 - "SKILL.md"
Cohesion: 0.18
Nodes (10): Completion Report, Done When, Key rules, Mandatory Post-Execution Hooks, Outline, Phase 0: Outline & Research, Phase 1: Design & Contracts, Phases (+2 more)

### Community 35 - "SKILL.md"
Cohesion: 0.18
Nodes (10): Completion Report, Done When, For AI Generation, Mandatory Post-Execution Hooks, Outline, Pre-Execution Checks, Quick Guidelines, Section Requirements (+2 more)

### Community 36 - "SKILL.md"
Cohesion: 0.18
Nodes (10): Checklist Format (REQUIRED), Completion Report, Done When, Mandatory Post-Execution Hooks, Outline, Phase Structure, Pre-Execution Checks, Task Generation Rules (+2 more)

### Community 37 - "SKILL.md"
Cohesion: 0.25
Nodes (7): Anti-Examples: What NOT To Do, Checklist Purpose: "Unit Tests for English", Example Checklist Types & Sample Items, Execution Steps, Post-Execution Checks, Pre-Execution Checks, User Input

### Community 38 - "SKILL.md"
Cohesion: 0.29
Nodes (6): Completion Report, Done When, Mandatory Post-Execution Hooks, Outline, Pre-Execution Checks, User Input

### Community 39 - "SKILL.md"
Cohesion: 0.29
Nodes (6): Completion Report, Done When, Mandatory Post-Execution Hooks, Outline, Pre-Execution Checks, User Input

### Community 40 - "SKILL.md"
Cohesion: 0.33
Nodes (5): Outline, Post-Execution Checks, Pre-Execution Checks, Scope Guard, User Input

### Community 41 - "SKILL.md"
Cohesion: 0.33
Nodes (5): Deprecation Notice, Outline, Post-Execution Checks, Pre-Execution Checks, User Input

### Community 42 - "Server"
Cohesion: 0.20
Nodes (9): Handler, ResolveFolderByName(), searchDirDepth(), Context, writeJSON(), Request, ResponseWriter, HealthProvider (+1 more)

### Community 43 - "Watcher"
Cohesion: 0.14
Nodes (11): Event, NewFilter(), T, TestFilterExtensionsAndIgnores(), TestIsEligibleContent(), Context, RWMutex, New() (+3 more)

### Community 44 - "FalkorStore"
Cohesion: 0.21
Nodes (11): asBool(), asInt(), asString(), cypherQuote(), Client, Context, NewFalkorStore(), NewFalkorStoreWithClient() (+3 more)

### Community 45 - "llm_client.go"
Cohesion: 0.12
Nodes (18): chatCompletionRequest, chatCompletionResponse, chatMessage, LLMClient, rawLLMFinding, rawLLMOutput, HashSnippet(), newFindingID() (+10 more)

### Community 46 - "Tasks: Monitor Local de Boas Práticas (v1)"
Cohesion: 0.11
Nodes (17): 0. Branch, 1.1 Commit Group 1, 2.1 Commit Group 2, 3.1 Commit Group 3, 4.1 Commit Group 4, 5.1 Commit Group 5, 6.1 Commit Group 6, Dependencies & Execution Order (+9 more)

### Community 47 - "Queue"
Cohesion: 0.21
Nodes (8): Context, Duration, Mutex, New(), T, TestQueueDebounceAndSequentialExecution(), Queue, Timer

### Community 48 - "User Scenarios & Testing *(mandatory)*"
Cohesion: 0.14
Nodes (13): Assumptions, Edge Cases, Feature Specification: Monitor Local de Boas Práticas (v1), Functional Requirements, Key Entities *(include if feature involves data)*, Measurable Outcomes, Requirements *(mandatory)*, Success Criteria *(mandatory)* (+5 more)

### Community 49 - "LlamaProcess"
Cohesion: 0.12
Nodes (16): EnvConfig, LlamaProcess, Bool, Cmd, LoadEnvConfig(), LoadEnvFile(), ParseAndValidateLlamaPort(), ValidateLlamaPort() (+8 more)

### Community 50 - "Data Model: Monitor Local de Boas Práticas (v1)"
Cohesion: 0.22
Nodes (8): 1. `:MonitoringConfig` (Singleton `id = "default"`), 2. `:Finding`, 3. `DebounceEntry` (em `internal/queue`), 4. `NotificationWindow` (em `internal/notify`), Data Model: Monitor Local de Boas Práticas (v1), Entidades em Memória (Execução), Entidades Persistidas no FalkorDB, Regras de Transição (Notificação vs. Histórico)

### Community 51 - "Implementation Plan: Monitor Local de Boas Práticas (v1)"
Cohesion: 0.22
Nodes (8): Complexity Tracking, Constitution Check, Documentation (this feature), Implementation Plan: Monitor Local de Boas Práticas (v1), Project Structure, Source Code (repository root), Summary, Technical Context

### Community 52 - "Endpoints JSON"
Cohesion: 0.25
Nodes (7): Contract: Local HTTP API & Web UI (`127.0.0.1`), Endpoints JSON, `GET /api/config`, `GET /api/findings`, `GET /api/health`, `PUT /api/config`, Rotas HTML

### Community 53 - "Research: Monitor Local de Boas Práticas (v1)"
Cohesion: 0.25
Nodes (7): 1. Cliente FalkorDB em Go (Configurações e Histórico), 2. Gerenciamento do `llama-server` e Comunicação LLM, 3. Watcher Recursivo, Debounce e Fila Sequencial, 4. Regras Determinísticas Pré-LLM, 5. Notificações Windows e Controle de Ruído, 6. Interface Web Local (`127.0.0.1`), Research: Monitor Local de Boas Práticas (v1)

### Community 54 - "Specification Quality Checklist: Monitor Local de Boas Práticas (v1)"
Cohesion: 0.33
Nodes (5): Content Quality, Feature Readiness, Notes, Requirement Completeness, Specification Quality Checklist: Monitor Local de Boas Práticas (v1)

### Community 56 - "Contract: LLM Analysis Input & Output Schema"
Cohesion: 0.40
Nodes (4): Contract: LLM Analysis Input & Output Schema, Isolamento de Prompt (Mitigação de Prompt Injection), Regras de Validação no App Go, Schema de Resposta Esperado (JSON Estrito)

### Community 57 - "Quickstart & Validation Guide: Monitor Local de Boas Práticas (v1)"
Cohesion: 0.40
Nodes (4): Cenários de Validação Ponta a Ponta, Pré-requisitos, Quickstart & Validation Guide: Monitor Local de Boas Práticas (v1), Verificação Automatizada (Build & Testes)

## Knowledge Gaps
- **303 isolated node(s):** `github.com/wesleyxavier/genova-file-monitoring`, `chatCompletionResponse`, `User Input`, `Pre-Execution Checks`, `Goal` (+298 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `appRuntime` connect `main.go` to `LlamaProcess`, `Watcher`, `llm_client.go`, `Queue`?**
  _High betweenness centrality (0.044) - this node is a cross-community bridge._
- **Why does `MonitoringConfig` connect `main.go` to `Server`, `FalkorStore`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Why does `Server` connect `Server` to `main.go`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **What connects `github.com/wesleyxavier/genova-file-monitoring`, `chatCompletionResponse`, `User Input` to the rest of the system?**
  _303 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Tasks: [FEATURE NAME]` be split into smaller, more focused modules?**
  _Cohesion score 0.07407407407407407 - nodes in this community are weakly interconnected._
- **Should `Execution Steps` be split into smaller, more focused modules?**
  _Cohesion score 0.125 - nodes in this community are weakly interconnected._
- **Should `genova-file-monitoring — Instruções para Agentes de IA` be split into smaller, more focused modules?**
  _Cohesion score 0.13333333333333333 - nodes in this community are weakly interconnected._