# genova-file-monitoring — Instruções para Agentes de IA

> Documentação para humanos: [README.md](README.md).
> Campos "A definir" devem ser preenchidos pelo dono do projeto. Não invente valores.

## Propósito

App local para Windows, escrito em Go. Observa uma pasta de projeto, analisa cada arquivo salvo com um modelo local (LLM) usando um guia de boas práticas e documentação cadastrada (RAG no FalkorDB, com índice vetorial) e avisa no Windows quando encontra problemas. Uso pessoal, sem publicação.

## Stack

- Go (`go.mod`: 1.27.4), módulo `github.com/wesleyxavier/genova-file-monitoring`
- fsnotify (watcher), beeep (notificações Windows)
- FalkorDB via Docker (`docker-compose.yml`)
- LLM local: llama-server (llama.cpp), modelo `gemma-4-12B-it-qat-q4_0.gguf`, porta 11440 (nunca 11435; faixa permitida 11440-11450)
- Modelo de embedding: A definir
- Modelo de embedding: A definir

## Comandos

| Ação | Comando |
| ---- | ------- |
| Build | `go build ./...` |
| Testes | `go test ./...` |
| Executar | `go run ./cmd/genova-file-monitoring` |
| FalkorDB | `docker compose --env-file .env up -d` |
| LLM | `llama-server -m C:\LLamaModels\gemma-4-12B-it-qat-q4_0.gguf -ngl 99 -fa on --cache-type-k q8_0 --cache-type-v q8_0 --port 11440` |

## Regras de operação

1. **Regra Zero:** nunca presuma. Requisito ambíguo vira pergunta curta ao usuário.
2. **Autorização:** nunca `commit`, `push` ou criar PR sem "sim" explícito do usuário na conversa.
3. **Nunca sobrescreva** arquivo existente sem perguntar.
4. **graphify:** com `graphify-out/graph.json` presente, consulte `graphify query "<pergunta>"` antes de varrer arquivos. Após alterar código, rode `graphify update .`.

## Fluxo de especificação

Spec Kit (`.specify/`): `/speckit.specify` → `/speckit.plan` → `/speckit.tasks` → aprovação → `/speckit.implement`.

## Convenções

A definir.

## Restrições

- O app só lê o projeto observado; nunca grava nele.
- Interface local apenas em 127.0.0.1.
- Código não sai da máquina; rede só na ingestão de documentação e na busca manual.
- No máximo 5 avisos por minuto; saída do LLM validada contra schema (JSON inválido é descartado).
- Docker é usado só para o FalkorDB.
- Segredos (`.env*`) nunca são commitados.
