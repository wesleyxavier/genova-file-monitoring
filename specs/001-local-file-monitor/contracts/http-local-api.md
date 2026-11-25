# Contract: Local HTTP API & Web UI (`127.0.0.1`)

**Feature**: `001-local-file-monitor` | **Bind Address**: Obrigatoriamente `127.0.0.1:<port>` (nunca `0.0.0.0`)

## Rotas HTML

- `GET /` — Interface web local mínima (HTML/JS embutido via `embed.FS`) para:
  - Visualizar status do watcher, do `llama-server` e do FalkorDB.
  - Configurar `watch_path`, `min_severity`, `silenced_rules`, `extensions` e `debounce_ms`.
  - Consultar e filtrar o histórico de achados (`severity`, `file_path`, `rule_id`, `notified`).

## Endpoints JSON

### `GET /api/health`
Retorna estado atual dos subsistemas locais.
```json
{
  "status": "ok",
  "watcher_active": true,
  "watch_path": "C:\\Projects\\Example",
  "llm_ready": true,
  "llm_port": 11440,
  "falkordb_ready": true
}
```

### `GET /api/config`
Retorna a configuração ativa persistida no FalkorDB.
```json
{
  "watch_path": "C:\\Projects\\Example",
  "min_severity": "alta",
  "silenced_rules": ["no-debug-log"],
  "extensions": [".go", ".ts", ".tsx", ".js", ".py"],
  "debounce_ms": 500
}
```

### `PUT /api/config`
Atualiza a configuração no FalkorDB e recarrega o watcher/notificador sem reiniciar o processo.
- **Validation**:
  - `watch_path`: se não vazio, deve existir e ser diretório.
  - `min_severity`: `"critica" | "alta" | "media" | "baixa"`.
  - `debounce_ms`: `100..10000`.
- **Responses**: `200 OK` com config atualizada, `400 Bad Request` se inválido, `503 Service Unavailable` se FalkorDB indisponível.

### `GET /api/findings`
Lista achados persistidos no FalkorDB (ordenados por `created_at` descendente).
- **Query params opcionais**:
  - `severity` (`critica|alta|media|baixa`)
  - `file` (substring do `file_path`)
  - `limit` (padrão `50`, máximo `200`)
- **Response**: `200 OK`
```json
{
  "items": [
    {
      "id": "01J...",
      "file_path": "internal/auth/login.go",
      "rule_id": "hardcoded-secret",
      "severity": "critica",
      "line_start": 42,
      "message": "Possível credencial fixa no código-fonte.",
      "source": "deterministic",
      "notified": true,
      "suppression_reason": "",
      "created_at": "2026-11-24T18:00:00-03:00"
    }
  ]
}
```
