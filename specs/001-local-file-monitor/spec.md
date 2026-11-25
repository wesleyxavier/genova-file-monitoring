# Feature Specification: Monitor Local de Boas Práticas (v1)

**Feature Branch**: `001-local-file-monitor`

**Created**: 2026-11-24

**Status**: Draft

**Input**: Constituição do projeto (`.specify/memory/constitution.md` v1.1.0), visão geral (`README.md`) e definições do usuário (Q1: interface local em `127.0.0.1`, Q2: FalkorDB para configurações e histórico, Q3: extensões configuráveis + ignore padrão + `.gitignore`).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Monitorar pasta e notificar problemas críticos/altos ao salvar (Priority: P1)

Como desenvolvedor usando Windows, quero iniciar o monitor apontando para uma pasta de projeto para receber notificações nativas quase em tempo real sempre que eu salvar um arquivo com problemas de qualidade ou segurança, sem que meu código saia da máquina.

**Why this priority**: É o propósito central do produto: dar feedback de qualidade no momento do save, quando a correção custa menos, mantendo privacidade total e operação somente leitura.

**Independent Test**: Pode ser testado iniciando o monitor em uma pasta de teste com um arquivo elegível contendo um bug evidente ou risco de segurança, salvando o arquivo e verificando que uma notificação nativa do Windows é exibida com a severidade correta sem qualquer alteração no arquivo observado.

**Acceptance Scenarios**:

1. **Given** o monitor está ativo observando uma pasta de projeto e o serviço de análise local está pronto, **When** o usuário salva um arquivo elegível com problema de severidade Crítica, **Then** o sistema exibe uma notificação no Windows com som indicando o arquivo, a regra violada e o problema encontrado.
2. **Given** o monitor está ativo e o usuário salva um arquivo elegível com problema de severidade Alta, **When** a análise é concluída, **Then** o sistema exibe uma notificação no Windows (sem som de alerta crítico) descrevendo o achado.
3. **Given** o usuário realiza múltiplos saves rápidos no mesmo arquivo dentro da janela de debounce (500 ms), **When** os eventos de salvamento ocorrem em sequência, **Then** o sistema agrupa os eventos em uma única análise após a estabilização do save, sem bloquear a IDE.
4. **Given** qualquer análise executada pelo monitor, **When** o arquivo é lido e avaliado, **Then** nenhum arquivo da pasta observada é criado, modificado, movido ou apagado, e nenhum dado do código é enviado para fora da máquina.

---

### User Story 2 - Controlar ruído e evitar fadiga de notificações (Priority: P2)

Como desenvolvedor, quero que o monitor limite a frequência de alertas, evite repetir avisos idênticos e respeite a severidade mínima configurada, para não ser interrompido por excesso de notificações durante o trabalho.

**Why this priority**: Sem controle de ruído, alertas repetitivos ou de baixa relevância tornam o monitor inutilizável no dia a dia.

**Independent Test**: Pode ser testado salvando repetidamente arquivos com múltiplos problemas (incluindo severidade Baixa e Média) e verificando que: no máximo 5 notificações são disparadas em 1 minuto, alertas duplicados para o mesmo arquivo e regra no mesmo trecho são suprimidos por 30 minutos, e achados abaixo da severidade mínima vão apenas para o histórico.

**Acceptance Scenarios**:

1. **Given** mais de 5 achados notificáveis são detectados dentro de 1 minuto, **When** o teto de 5 avisos por minuto é atingido, **Then** os avisos excedentes não geram notificação na tela e são registrados apenas no histórico persistido.
2. **Given** um aviso já foi emitido para um arquivo e regra específicos, **When** o usuário salva o mesmo arquivo novamente dentro de 30 minutos sem alterar o trecho relacionado, **Then** o sistema suprime a nova notificação para essa mesma combinação de arquivo e regra.
3. **Given** um aviso foi suprimido por deduplicação dentro dos 30 minutos, **When** o usuário altera o trecho do arquivo relacionado àquela regra e o problema persiste ou muda, **Then** o sistema permite nova avaliação e notificação conforme as regras de severidade.
4. **Given** a severidade mínima configurada é Alta, **When** a análise encontra problemas de severidade Média ou Baixa, **Then** o sistema não exibe notificação e registra os achados no histórico persistido.
5. **Given** o usuário silenciou uma regra específica, **When** essa regra é violada em novos saves, **Then** o sistema não exibe notificação para a regra silenciada.

---

### User Story 3 - Ciclo de vida automático do analisador local e validação estrita (Priority: P2)

Como usuário do monitor, quero que o aplicativo gerencie automaticamente a inicialização e o encerramento do modelo local de análise e descarte respostas malformadas, para ter uma experiência confiável sem precisar gerenciar processos manualmente.

**Why this priority**: Garante que o app funcione como ferramenta autocontida e que alucinações ou saídas inválidas do modelo nunca virem falsos alertas na tela.

**Independent Test**: Pode ser testado iniciando e encerrando o monitor e verificando que o processo de análise local sobe na faixa de porta esperada, é encerrado junto com o app, e que uma resposta inválida do modelo é descartada e registrada em log sem gerar notificação.

**Acceptance Scenarios**:

1. **Given** o usuário inicia o monitor com a configuração de ambiente local válida, **When** o aplicativo inicializa, **Then** ele inicia o servidor de modelo local como processo filho em uma porta dentro da faixa permitida (11440–11450, nunca 11435) e aguarda prontidão antes de processar análises.
2. **Given** o monitor está em execução com o servidor de modelo local ativo, **When** o usuário encerra o monitor, **Then** o aplicativo encerra o processo filho do servidor de modelo de forma limpa antes de finalizar.
3. **Given** o modelo local retorna uma resposta fora do formato esperado ou inválida, **When** o sistema valida o resultado da análise, **Then** o resultado inválido é descartado e registrado em log, sem gerar notificação ao usuário.
4. **Given** um arquivo observado contém instruções maliciosas ou comandos em comentários/strings tentando alterar o comportamento da análise, **When** o arquivo é analisado, **Then** o conteúdo do arquivo é tratado estritamente como dado analisado e não como instrução.

---

### User Story 4 - Configurar monitoramento e consultar histórico via interface local (Priority: P3)

Como usuário, quero acessar uma interface web local (escutando apenas em `127.0.0.1`) para definir a pasta observada, ajustar severidade mínima, silenciar regras, gerenciar extensões monitoradas e consultar o histórico de achados persistido no FalkorDB.

**Why this priority**: Completa o ciclo de uso permitindo personalização dinâmica e auditoria dos achados (incluindo os de severidade Baixa e os suprimidos por limite de ruído).

**Independent Test**: Pode ser testado abrindo a interface local em `127.0.0.1`, configurando a pasta de projeto e a severidade mínima para Média, salvando um arquivo com achado Baixo, reiniciando ou recarregando a página e verificando que a configuração e o achado permanecem disponíveis no histórico.

**Acceptance Scenarios**:

1. **Given** o usuário acessa a interface local em `127.0.0.1`, **When** informa a pasta do projeto a observar, a severidade mínima, as regras silenciadas e as extensões desejadas, **Then** o sistema persiste essas configurações no FalkorDB local e as aplica imediatamente ao monitoramento.
2. **Given** ocorreram achados (incluindo severidade Baixa e excedentes do teto de ruído), **When** o usuário consulta o histórico pela interface local em `127.0.0.1`, **Then** o sistema exibe os registros persistidos no FalkorDB contendo arquivo, severidade, regra, descrição e data/hora, permitindo filtragem.
3. **Given** o aplicativo é reiniciado, **When** o usuário inicia o monitor novamente, **Then** as configurações salvas anteriormente no FalkorDB são restauradas automaticamente.

### Edge Cases

- **Arquivos binários, vazios ou em pastas ignoradas**: O sistema ignora por padrão pastas como `.git`, `node_modules`, `bin`, `obj`, `dist`, `build` e `vendor`, respeita regras do `.gitignore` da pasta observada, ignora arquivos binários ou vazios e analisa apenas extensões de código habilitadas na lista configurável (padrão inicial: `.go`, `.ts`, `.tsx`, `.js`, `.jsx`, `.py`, `.cs`, `.java`, `.rs`, `.sql`).
- **Múltiplos saves simultâneos**: Quando vários arquivos diferentes são salvos quase ao mesmo tempo (ex.: formatador ou troca de branch), a fila processa um job por vez sem travar o watcher e sem perder eventos além do agrupamento por debounce de cada arquivo.
- **Falha ou indisponibilidade do modelo local ou do FalkorDB**: Se o servidor do modelo local falhar ao iniciar, cair ou expirar timeout, ou se o FalkorDB estiver indisponível, o monitor registra o erro em log, informa o estado na interface local sem travar o processo e não gera notificações falsas.
- **Arquivo apagado ou renomeado antes da leitura**: O sistema trata a ausência do arquivo graciosamente, registra em log se necessário e segue para o próximo item da fila.
- **Regras determinísticas antes do modelo**: Quando existirem regras determinísticas aplicáveis à linguagem do arquivo salvo, elas executam antes da chamada ao modelo local.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST observar alterações de salvamento em arquivos dentro da pasta de projeto configurada pelo usuário, operando exclusivamente em modo somente leitura sobre a pasta observada.
- **FR-002**: O sistema MUST filtrar os arquivos observados aplicando lista configurável de extensões de código, ignorando por padrão pastas de controle/dependências/build (`.git`, `node_modules`, `bin`, `obj`, `dist`, `build`, `vendor`), arquivos binários/vazios e respeitando o `.gitignore` do projeto observado.
- **FR-003**: O sistema MUST aplicar debounce por arquivo (padrão de 500 ms) para agrupar múltiplos eventos de um mesmo salvamento antes de enfileirar a análise.
- **FR-004**: O sistema MUST processar análises em segundo plano em fila sequencial (um job por vez) para não bloquear a IDE nem sobrecarregar o modelo local.
- **FR-005**: O sistema MUST executar regras determinísticas aplicáveis à linguagem do arquivo antes de invocar o modelo local.
- **FR-006**: O sistema MUST analisar o conteúdo do arquivo utilizando o guia de boas práticas embutido como texto fixo, sem busca web, sem ingestão de URLs, sem RAG, sem embeddings e sem chamadas de rede externas.
- **FR-007**: O sistema MUST tratar todo conteúdo lido dos arquivos observados (código, strings, comentários) exclusivamente como dado de entrada, nunca como instrução ao analisador.
- **FR-008**: O sistema MUST validar a saída do modelo local contra um schema estruturado; saídas inválidas MUST ser descartadas e registradas em log, sem gerar notificação.
- **FR-009**: O sistema MUST iniciar o servidor de modelo local como processo filho usando o comando definido na configuração de ambiente local, em porta na faixa 11440–11450 (nunca 11435), e encerrá-lo ao finalizar o aplicativo.
- **FR-010**: O sistema MUST classificar achados em quatro níveis de severidade (Crítica, Alta, Média, Baixa) e notificar no Windows apenas conforme a regra de cada nível e a severidade mínima configurada (Crítica com som; Alta sem som crítico; Média apenas se mínimo configurado for Média; Baixa nunca notifica).
- **FR-011**: O sistema MUST impor teto máximo de 5 notificações por minuto; achados excedentes MUST ser registrados apenas no histórico persistido e em log.
- **FR-012**: O sistema MUST deduplicar notificações por combinação de arquivo + regra durante 30 minutos, liberando nova notificação antes desse prazo apenas se o trecho relacionado tiver sido alterado.
- **FR-013**: O sistema MUST expor uma interface web local escutando exclusivamente em `127.0.0.1` para configurar pasta observada, severidade mínima, regras silenciadas e extensões monitoradas, bem como consultar o histórico de achados.
- **FR-014**: O sistema MUST persistir as configurações de monitoramento e o histórico de achados no FalkorDB local (sem uso de RAG ou embeddings na v1).

### Key Entities *(include if feature involves data)*

- **Arquivo Observado**: Representa um arquivo elegível na pasta monitorada (caminho relativo, extensão, momento do último save, hash/conteúdo do trecho para deduplicação).
- **Achado (Finding)**: Resultado validado de uma análise persistido no FalkorDB, contendo arquivo, regra violada, severidade (Crítica, Alta, Média, Baixa), trecho/localização, descrição objetiva e timestamp.
- **Notificação**: Alerta nativo emitido no Windows a partir de um achado elegível, sujeito ao teto de 5/min, severidade mínima e janela de deduplicação de 30 minutos.
- **Configuração de Monitoramento**: Preferências ativas do usuário persistidas no FalkorDB (pasta observada, severidade mínima, regras silenciadas, extensões habilitadas, janela de debounce).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% dos arquivos salvos dentro da pasta observada permanecem intactos (zero operações de escrita, movimentação ou exclusão na pasta do projeto).
- **SC-002**: 100% da comunicação de rede durante a v1 ocorre exclusivamente em loopback local (`127.0.0.1` / `localhost`), sem qualquer tráfego externo.
- **SC-003**: O usuário recebe feedback nativo no Windows para problemas Críticos/Altos logo após a janela de debounce de 500 ms e a conclusão da análise em fila, sem congelamento ou bloqueio da IDE.
- **SC-004**: Em cenários de salvamentos intensivos com múltiplos problemas, o sistema emite no máximo 5 notificações por minuto e zero notificações repetidas para a mesma regra e trecho dentro de 30 minutos.
- **SC-005**: 100% das respostas malformadas do modelo local são descartadas silenciosamente na UI (zero alertas falsos por erro de formato), com registro em log.
- **SC-006**: 100% das configurações salvas e dos achados registrados permanecem disponíveis após reiniciar o aplicativo por meio da persistência no FalkorDB local.

## Assumptions

- O usuário utiliza Windows 10 ou Windows 11 com suporte a notificações nativas do sistema operacional.
- O binário do servidor de modelo local (`llama-server`), o arquivo do modelo GGUF e a instância local do FalkorDB (`docker compose`) estão disponíveis localmente na máquina do usuário e configurados via `.env` (não commitado).
- A v1 é de uso estritamente pessoal e local, sem autenticação multiusuário, sem empacotamento para distribuição pública e sem RAG/embeddings.
- Alterações em arquivos são detectadas a partir de eventos do sistema de arquivos após o início do monitor (não há varredura retroativa obrigatória de todo o repositório ao iniciar).
