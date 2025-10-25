# API-to-MCP Server

## Descrizione

API-to-MCP è un server che converte automaticamente specifiche OpenAPI/Swagger in strumenti MCP (Model Context Protocol) esposti tramite JSON-RPC. Il server analizza uno schema OpenAPI e genera dinamicamente tools che permettono di interagire con API REST esistenti attraverso il protocollo MCP.

## Obiettivo

Semplificare l'integrazione di API REST con assistenti AI e altri client MCP, automatizzando la creazione di strumenti basati su specifiche OpenAPI esistenti.

## Funzionalità Principali

- **Parsing OpenAPI**: Analizza file OpenAPI (JSON/YAML) per estrarre endpoint, parametri e schemi
- **Generazione Tools**: Converte automaticamente endpoint REST in tools MCP
- **Server JSON-RPC**: Espone i tools generati tramite protocollo JSON-RPC
- **Mapping Tipi**: Mappa automaticamente tipi di dati OpenAPI in tipi MCP
- **Configurazione Flessibile**: Supporta filtri e personalizzazioni per la generazione dei tools

## Architettura

```
OpenAPI Spec → Parser → Tool Generator → MCP Server → JSON-RPC
```

## Tecnologie

- **Linguaggio**: Go
- **OpenAPI Parsing**: kin-openapi
- **JSON-RPC**: gorilla/rpc
- **HTTP Client**: go-resty
- **Configurazione**: viper

## Caso d'Uso

1. Fornisci un file OpenAPI/Swagger
2. Il server analizza la specifica
3. Genera automaticamente tools MCP per ogni endpoint
4. Espone i tools via JSON-RPC
5. I client possono utilizzare i tools per interagire con l'API originale

## Esempio

Dato un endpoint OpenAPI:
```yaml
/pets/{petId}:
  get:
    summary: Get pet by ID
    parameters:
      - name: petId
        in: path
        required: true
        schema:
          type: integer
```

Il server genera automaticamente un tool MCP:
```json
{
  "name": "get_pet_by_id",
  "description": "Get pet by ID",
  "inputSchema": {
    "type": "object",
    "properties": {
      "petId": {"type": "integer"}
    }
  }
}
```

## Piano di Sviluppo

### Fase 1: Setup e Foundation
- [x] Inizializzazione progetto Go con moduli
- [x] Setup struttura directory e package
- [x] Configurazione dipendenze (go.mod)
- [x] Setup logging e configurazione base
- [x] Creazione entry point principale

### Fase 2: Parser OpenAPI
- [ ] Implementazione parser OpenAPI base
- [ ] Supporto per file JSON e YAML
- [ ] Estrazione endpoint e metodi HTTP
- [ ] Parsing parametri e schemi
- [ ] Validazione struttura OpenAPI
- [ ] Gestione errori di parsing

### Fase 3: Generazione Tools MCP
- [ ] Definizione strutture dati MCP
- [ ] Mapping endpoint → tools MCP
- [ ] Conversione tipi di dati OpenAPI → MCP
- [ ] Generazione metadata tools (nome, descrizione)
- [ ] Gestione parametri richiesti/opzionali
- [ ] Validazione schema input

### Fase 4: Server JSON-RPC
- [ ] Implementazione server JSON-RPC base
- [ ] Registrazione tools MCP come metodi
- [ ] Gestione richieste e risposte
- [ ] Routing delle chiamate ai tools
- [ ] Gestione errori JSON-RPC
- [ ] Supporto per batch requests

### Fase 5: HTTP Client e Integrazione
- [ ] Implementazione HTTP client per chiamate API
- [ ] Gestione autenticazione (API key, Bearer token)
- [ ] Mapping parametri MCP → parametri HTTP
- [ ] Gestione response e errori HTTP
- [ ] Timeout e retry logic
- [ ] Supporto per diversi content types

### Fase 6: Configurazione e Filtri
- [ ] Sistema di configurazione (YAML/JSON)
- [ ] Filtri per includere/escludere endpoint
- [ ] Configurazione base URL API
- [ ] Gestione variabili ambiente
- [ ] Validazione configurazione
- [ ] Hot reload configurazione

### Fase 7: Testing e Validazione
- [ ] Test unitari per parser OpenAPI
- [ ] Test unitari per generazione tools
- [ ] Test integrazione server JSON-RPC
- [ ] Test end-to-end con API reali
- [ ] Test performance e carico
- [ ] Validazione con diversi schemi OpenAPI

### Fase 8: Documentazione e Deployment
- [ ] Documentazione API completa
- [ ] Esempi di utilizzo
- [ ] Guida configurazione
- [ ] Docker container
- [ ] Script di build e deploy
- [ ] README con quick start

### Fase 9: Ottimizzazioni e Features Avanzate
- [ ] Caching delle risposte API
- [ ] Rate limiting
- [ ] Monitoring e metrics
- [ ] Supporto per WebSocket
- [ ] Batch operations
- [ ] Supporto per streaming responses

### Fase 10: Release e Distribuzione
- [ ] Versioning semantico
- [ ] Release notes
- [ ] Distribuzione binari
- [ ] Package manager (Homebrew, etc.)
- [ ] CI/CD pipeline
- [ ] Security audit
