# Data Model: Monitor Local de Boas Práticas (v1)

**Feature**: `001-local-file-monitor` | **Storage**: FalkorDB (`GRAPH.QUERY` no grafo `genova_monitor`)

## Entidades Persistidas no FalkorDB

### 1. `:MonitoringConfig` (Singleton `id = "default"`)

Armazena as preferências ativas do usuário.

| Campo | Tipo | Padrão | Validação / Regras |
| --- | --- | --- | --- |
| `id` | string | `"default"` | Chave única do nó singleton |
| `watch_path` | string | `""` | Caminho absoluto de diretório existente no Windows (obrigatório para iniciar watcher) |
| `min_severity` | string | `"alta"` | Enum: `"critica"`, `"alta"`, `"media"`, `"baixa"` |
| `silenced_rules` | string (JSON array) | `[]` | Lista de IDs de regras silenciadas (ex.: `["no-debug-log"]`) |
| `extensions` | string (JSON array) | `[".go",".ts",".tsx",".js",".jsx",".py",".cs",".java",".rs",".sql"]` | Extensões iniciadas com `.` |
| `debounce_ms` | int | `500` | `>= 100` e `<= 10000` |
| `updated_at` | string (RFC3339) | timestamp atual | Atualizado a cada alteração via UI `127.0.0.1` |

### 2. `:Finding`

Registra cada achado validado (incluindo severidade `baixa` e achados suprimidos por limite de ruído/deduplicação).

| Campo | Tipo | Validação / Regras |
| --- | --- | --- |
| `id` | string (UUID/ULID) | Identificador único do achado |
| `file_path` | string | Caminho relativo à `watch_path` |
| `rule_id` | string | Identificador curto da regra (ex.: `"sql-injection"`, `"hardcoded-secret"`) |
| `severity` | string | Enum: `"critica"`, `"alta"`, `"media"`, `"baixa"` |
| `line_start` | int | `>= 0` (0 se aplicável ao arquivo todo) |
| `snippet_hash` | string | Hash SHA-256 do trecho relacionado (para deduplicação) |
| `message` | string | Descrição objetiva do problema (não vazia) |
| `source` | string | `"deterministic"` ou `"llm"` |
| `notified` | bool | `true` se gerou toast no Windows; `false` se suprimido/abaixo do mínimo |
| `suppression_reason` | string | `""`, `"below_min_severity"`, `"silenced_rule"`, `"rate_limit"`, `"deduplicated"` |
| `created_at` | string (RFC3339) | Data/hora do registro |

## Entidades em Memória (Execução)

### 3. `DebounceEntry` (em `internal/queue`)
- `file_path`: caminho absoluto do arquivo
- `timer`: `*time.Timer` resetado a cada evento de escrita dentro de `debounce_ms`

### 4. `NotificationWindow` (em `internal/notify`)
- `recent_timestamps`: slice de `time.Time` dos últimos 60 segundos (máximo 5 entradas)
- `dedup_cache`: mapa `file_path + "|" + rule_id` → `{ snippet_hash, expires_at }` (TTL 30 min)

## Regras de Transição (Notificação vs. Histórico)

1. Achado validado entra no avaliador de ruído.
2. **Sempre** é persistido no FalkorDB como `:Finding` (com `notified` e `suppression_reason` preenchidos).
3. Critérios de supressão (em ordem):
   - Regra está em `silenced_rules` → `notified = false`, `suppression_reason = "silenced_rule"`
   - Severidade é `baixa` ou menor que `min_severity` → `notified = false`, `suppression_reason = "below_min_severity"`
   - Deduplicação ativa (`arquivo + regra` dentro de 30 min com mesmo `snippet_hash`) → `notified = false`, `suppression_reason = "deduplicated"`
   - Teto de 5 notificações nos últimos 60s atingido → `notified = false`, `suppression_reason = "rate_limit"`
4. Caso contrário: `notified = true`, dispara toast Windows (`Alert` se `critica`, `Notify` se `alta`/`media`) e atualiza janela e cache de deduplicação.
