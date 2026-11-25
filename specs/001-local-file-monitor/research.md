# Research: Monitor Local de Boas Práticas (v1)

**Feature**: `001-local-file-monitor` | **Date**: 2026-11-24

## 1. Cliente FalkorDB em Go (Configurações e Histórico)

- **Decision**: Utilizar cliente Redis compatível com protocolo RESP (`github.com/redis/go-redis/v9`) executando comandos `GRAPH.QUERY` no FalkorDB, ou o SDK oficial `github.com/FalkorDB/falkordb-go/v2` (que encapsula `go-redis/v9`), com versão fixada no `go.mod`.
- **Rationale**: O `docker-compose.yml` do projeto já sobe `falkordb/falkordb:latest` na porta `6379` com autenticação `FALKORDB_USER` / `FALKORDB_PASSWORD`. Usar Cypher via `GRAPH.QUERY` mantém um único grafo (`genova_monitor`) para nós `:Config` e `:Finding`, já preparando o terreno para a fase futura sem adicionar outro banco (como SQLite) na v1.
- **Alternatives considered**:
  - `SQLite` local: rejeitado porque o usuário definiu explicitamente o uso do FalkorDB para configurações e histórico (Constituição v1.1.0, Princípio V).
  - Driver RESP manual via `net.Conn`: rejeitado por reinventar parsing RESP e pool de conexões sem ganho (Princípio X permite dependências com versão fixada no `go.mod`).

## 2. Gerenciamento do `llama-server` e Comunicação LLM

- **Decision**:
  - O app Go lê a variável de ambiente `LLAMA_SERVER_CMD` (ou equivalente configurada em `.env`, mantendo compatibilidade caso o usuário defina o nome exato) e inicia o `llama-server` via `os/exec` (`exec.CommandContext`) com porta validada na faixa `11440–11450` (rejeitando explicitamente `11435`).
  - Comunicação via `net/http` da stdlib apontando exclusivamente para `http://127.0.0.1:<porta>/v1/chat/completions` (e polling em `/health` na inicialização).
  - Resposta estruturada forçada via `response_format` (`json_object` / `json_schema` suportado pelo `llama-server`) e validada estritamente em Go (`encoding/json` + validação de campos obrigatórios e enum de severidade).
  - Encerramento limpo via `ctx` cancelado + `cmd.Process.Kill()`/`Wait()` no shutdown (Ctrl+C / `SIGTERM` / encerramento do app).
- **Rationale**: Atende aos Princípios I (loopback), IV (schema estrito, prompt injection mitigado separando `system` prompt com o guia fixo e `user` message com delimitadores claros para código como dado) e VI (processo filho gerenciado).
- **Alternatives considered**:
  - Subir `llama-server` externamente: rejeitado pelo Princípio VI ("O app Go inicia o `llama-server` como processo filho e o encerra ao sair").

## 3. Watcher Recursivo, Debounce e Fila Sequencial

- **Decision**:
  - `github.com/fsnotify/fsnotify` (versão fixada no `go.mod`) com registro recursivo de subdiretórios (adicionando novos diretórios criados dinamicamente, ignorando `.git`, `node_modules`, `bin`, `obj`, `dist`, `build`, `vendor` e regras do `.gitignore`).
  - Debounce por arquivo (`map[string]*time.Timer` protegido por `sync.Mutex`) com janela padrão de `500ms`.
  - Fila sequencial (`chan string` com buffer controlado + 1 goroutine worker) garantindo 1 job de análise por vez (Princípio VIII).
  - Leitura somente leitura (`os.ReadFile`), detecção de arquivo binário (verificação de byte nulo nos primeiros bytes / `http.DetectContentType`) e limite de tamanho máximo por arquivo (ex.: 256 KB padrão) para evitar travar o contexto do LLM.
- **Rationale**: `fsnotify` no Windows usa `ReadDirectoryChangesW`, que não é recursivo por padrão para subpastas no wrapper Go; portanto, percorrer diretórios elegíveis com `filepath.WalkDir` na inicialização e observar eventos `Create` de diretórios garante cobertura completa sem observar pastas ignoradas.
- **Alternatives considered**:
  - Polling periódico de timestamps: rejeitado por latência maior e alto uso de I/O em repositórios grandes.

## 4. Regras Determinísticas Pré-LLM

- **Decision**: Criar pipeline de regras determinísticas em `internal/analyzer/deterministic.go` que executa antes do LLM (Princípio IX). Exemplos iniciais agnósticos/por extensão: detecção de possíveis segredos hardcoded (chaves/tokens óbvios), `fmt.Println`/`console.log` de debug excessivo (severidade Baixa), ou arquivos com conflitos de merge (`<<<<<<<`, `=======`, `>>>>>>>` como Crítica).
- **Rationale**: Garante resposta instantânea sem custo de inferência para padrões triviais e cumpre o Princípio IX.
- **Alternatives considered**: Enviar tudo direto ao LLM: rejeitado por violar o Princípio IX.

## 5. Notificações Windows e Controle de Ruído

- **Decision**:
  - `github.com/gen2brain/beeep` (versão fixada no `go.mod`) para notificações nativas no Windows 10/11 (`beeep.Alert` para severidade Crítica com som e `beeep.Notify` para Alta/Média).
  - Rate limiter em janela deslizante de 1 minuto com teto de `5` notificações.
  - Deduplicação em memória (e/ou consultando estado recente) por chave `arquivo + regra` com TTL de `30 minutos` e hash SHA-256 do trecho (`snippet_hash`) para liberar novo alerta caso o trecho seja alterado.
- **Rationale**: Atende exatamente às regras de domínio do `README.md` e Princípio VII.

## 6. Interface Web Local (`127.0.0.1`)

- **Decision**: Servidor HTTP em `internal/web` usando `net/http` da stdlib + `embed.FS` para HTML/CSS/JS estático (zero build frontend externo na v1), escutando obrigatoriamente em `127.0.0.1:<WEB_PORT>` (ex.: `127.0.0.1:8080` ou configurável, validando que o host seja sempre `127.0.0.1`).
- **Rationale**: Mantém o binário Go autocontido, sem Node.js/npm necessário para rodar a v1, e cumpre rigorosamente o Princípio III e X (preferência pela stdlib).
