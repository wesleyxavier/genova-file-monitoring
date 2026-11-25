<!--
Sync Impact Report (remover antes do commit)
- Versão: (template não preenchido) → 1.0.0
- Princípios: 11 adicionados (I–XI); nenhum modificado/removido
- Seções adicionadas: Restrições Técnicas e de Escopo da v1; Fluxo de Desenvolvimento
- Seções removidas: nenhuma
- TODOs adiados: nome da variável do .env com o comando do llama-server (A definir)
-->
# genova-file-monitoring Constitution

## Core Principles

### I. Local-First e Privacidade

O código observado NUNCA sai da máquina. A v1 NÃO faz nenhuma chamada de rede externa: sem busca
web e sem ingestão de URLs. Única comunicação de rede permitida: loopback (llama-server local).

### II. Somente Leitura

O app NUNCA grava, move, apaga nem reescreve arquivos do projeto observado. Isso evita o loop de
escrita (o app alterar o que observa) e garante que o app só avisa, nunca age.

### III. Interface Local Mínima

Se houver UI/HTTP, ela MUST escutar apenas em `127.0.0.1`. Nenhum listener em `0.0.0.0`.

### IV. O Modelo Sugere, o App Decide

A saída do LLM MUST ser validada contra schema. JSON inválido é descartado e registrado em log,
sem gerar aviso. O conteúdo analisado (código, comentários) é dado, nunca instrução (mitigação de
prompt injection).

### V. Análise Direta com LLM, sem RAG na v1

O contexto da análise = conteúdo do arquivo + guia de boas práticas embutido como texto fixo.
RAG, embeddings e ingestão de documentação estão FORA do escopo da v1.
O FalkorDB (`docker-compose.yml` e `.env`) é usado na v1 exclusivamente para persistência
de configurações e histórico de achados, permanecendo proibido o uso para busca vetorial ou RAG na v1.

### VI. LLM Gerenciado pelo App

O app Go inicia o `llama-server` (llama.cpp) como processo filho e o encerra ao sair. O comando de
inicialização (modelo e flags) vem do arquivo `.env` (nome da variável: A definir). A
porta do llama-server MUST ficar na faixa 11440–11450 e NUNCA ser 11435.

### VII. Ruído Controlado

Teto de 5 avisos por minuto; deduplicação por arquivo + regra; severidade mínima configurável.
Excedentes e achados abaixo do mínimo vão para log/histórico, não para notificação.

### VIII. Não Bloqueia a IDE

Análises rodam em segundo plano, em fila com um job por vez. Debounce por arquivo, padrão 500 ms,
para agrupar os eventos de um mesmo save.

### IX. Determinismo antes do Modelo

Regras determinísticas, quando existirem para a linguagem, MUST rodar antes do LLM.

### X. Dependências e Segredos

Preferir a stdlib do Go. Dependências externas MUST ter versão fixada no `go.mod`. Segredos ficam
somente em arquivos `.env*` ignorados pelo git (exceto `.env.example`, sem valores).

### XI. Fluxo de Trabalho

Mudanças seguem Spec Kit, com worktree por feature. Commit, push e PR exigem autorização explícita
do usuário na conversa.

## Restrições Técnicas e de Escopo da v1

- Plataforma: Windows 10/11; Go conforme `go.mod`.
- Notificações nativas via biblioteca definida na spec (PRD cita beeep).
- Fora da v1: busca web, ingestão de URLs, RAG, embeddings.
- FalkorDB na v1: permitido apenas para persistência de configurações e histórico de achados.
- O app NUNCA grava no projeto observado (Princípio II).

## Fluxo de Desenvolvimento

- Sequência: `/speckit.specify` → `/speckit.plan` → `/speckit.tasks` → aprovação do usuário →
  `/speckit.implement`.
- Todo `tasks.md` começa com `## 0. Branch` (worktree) e termina com atualização do README,
  `graphify update .` e commit final + push + PR.
- Commit, push e PR somente com autorização explícita.

## Governance

Esta constituição prevalece sobre outras práticas do projeto. Emendas são feitas via
`/speckit.constitution`, com versionamento semântico: MAJOR para remoção/redefinição incompatível
de princípios; MINOR para novo princípio ou expansão material; PATCH para esclarecimentos.
Toda spec, plan e PR MUST verificar conformidade com os princípios; desvios exigem justificativa
documentada. Orientação operacional em `AGENTS.md`.

**Version**: 1.1.0 | **Ratified**: 2026-11-24 | **Last Amended**: 2026-11-24
