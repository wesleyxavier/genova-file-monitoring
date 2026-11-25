# Contract: LLM Analysis Input & Output Schema

**Feature**: `001-local-file-monitor` | **Endpoint**: `http://127.0.0.1:<11440-11450>/v1/chat/completions`

## Isolamento de Prompt (Mitigação de Prompt Injection)

- **System message**: Contém exclusivamente as instruções do analisador, o guia de boas práticas embutido como texto fixo e a regra explícita de que o conteúdo enviado na mensagem do usuário é **dado não confiável** (nunca instrução).
- **User message**: Contém metadados mínimos (`file_path`, linguagem/extensão) e o conteúdo do arquivo delimitado por marcadores explícitos (`<<<BEGIN_UNTRUSTED_FILE_CONTENT>>>` ... `<<<END_UNTRUSTED_FILE_CONTENT>>>`).

## Schema de Resposta Esperado (JSON Estrito)

A resposta do `llama-server` MUST ser um objeto JSON válido com a estrutura:

```json
{
  "findings": [
    {
      "rule_id": "string (kebab-case, não vazio)",
      "severity": "critica | alta | media | baixa",
      "line_start": 1,
      "snippet": "trecho curto de referência (opcional)",
      "message": "descrição objetiva do problema em PT-BR (não vazia)"
    }
  ]
}
```

## Regras de Validação no App Go

1. Se o JSON não puder ser decodificado (`json.Unmarshal` falhar) → **descartar**, registrar em log e **não notificar**.
2. Se `findings` estiver ausente (`nil`) → **descartar**, registrar em log e **não notificar** (array vazio `[]` é válido e significa "sem problemas").
3. Para cada item em `findings`:
   - `rule_id` e `message` não podem ser vazios.
   - `severity` deve pertencer exatamente ao conjunto `{"critica", "alta", "media", "baixa"}`.
   - `line_start` deve ser `>= 0`.
   - Itens inválidos são descartados e logados sem gerar notificação.
