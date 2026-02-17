# Rate Limiter

Rate limiter em Go que controla o tráfego de requisições para um serviço web, limitando por **endereço IP** ou **token de acesso**.

## Como funciona

O rate limiter atua como middleware HTTP interceptando todas as requisições antes que cheguem ao handler. Para cada requisição, ele:

1. Verifica se existe um header `API_KEY`. Se existir, aplica o limite configurado para aquele token (sobrepondo o limite por IP).
2. Caso contrário, aplica o limite por IP do cliente.
3. Se o limite for excedido, bloqueia o IP/token pelo tempo configurado e retorna `429 Too Many Requests`.

### Janela de tempo

O controle usa uma **sliding window** de 1 segundo implementada com sorted sets do Redis. Cada requisição é registrada com timestamp em nanossegundos, e a contagem considera apenas os registros dentro do último segundo.

### Bloqueio

Quando o limite é excedido, o IP ou token é bloqueado por um período configurável (`BLOCK_TIME_IN_SECONDS`). Durante o bloqueio, todas as requisições são recusadas. Após o tempo expirar, o bloqueio é removido automaticamente via TTL do Redis.

### Prioridade Token > IP

Se a requisição contém o header `API_KEY`, **somente** o limite do token é avaliado. O limite por IP é ignorado, permitindo que tokens tenham limites independentes (maiores ou menores) que o padrão por IP.

## Arquitetura

```
cmd/server/main.go          → Inicialização e wiring
internal/
├── infra/
│   ├── limiter/limiter.go   → Lógica de rate limiting (regras de negócio)
│   ├── middleware/limiter.go → Middleware HTTP (extrai IP/token, retorna 429)
│   ├── model/api-key.go     → Modelo de domínio do API key
│   └── repository/
│       ├── interface.go      → Interface StoreKey (strategy pattern)
│       └── redis.go          → Implementação Redis
└── webserver/webserver.go   → Servidor HTTP e handlers
```

### Strategy Pattern

A interface `StoreKey` em `repository/interface.go` abstrai o mecanismo de persistência. Para trocar o Redis por outro backend (memória, banco relacional, etc.), basta implementar a interface:

```go
type StoreKey interface {
    SaveKey(ctx context.Context, apiKey model.ApiKey) error
    GetApiKeyAttributes(ctx context.Context, key string) (rate int, valid bool, block bool, err error)
    GetRequestsLastSecond(ctx context.Context, prefix, id string) (int, error)
    AddRequest(ctx context.Context, prefix, id string) error
    Block(ctx context.Context, prefix, id string, blockTime time.Duration) error
    IsBlocked(ctx context.Context, prefix, id string) (bool, error)
}
```

## Configuração

Variáveis de ambiente (ou arquivo `.env` na raiz):

| Variável | Descrição | Padrão |
|---|---|---|
| `MAX_REQUESTS_BY_IP_PER_SECOND` | Limite de requisições por segundo por IP | `10` |
| `BLOCK_TIME_IN_SECONDS` | Tempo de bloqueio após exceder o limite (em segundos) | `300` (5 min) |
| `REDIS_ADDR` | Endereço do Redis | `localhost:6379` |
| `REDIS_PASSWORD` | Senha do Redis | (vazio) |
| `REDIS_DB` | Database do Redis | `0` |

Para tokens, o limite por segundo é definido individualmente na criação do token via endpoint `/api-key`.

## Executando com Docker

```bash
docker compose up --build
```

O servidor estará disponível em `http://localhost:8080`.

## Endpoints

### `GET /`

Endpoint protegido pelo rate limiter. Retorna `Hello World` com status `200`.

Quando o limite é excedido:
- **Status:** `429 Too Many Requests`
- **Mensagem:** `you have reached the maximum number of requests or actions allowed within a certain time frame`

### `POST /api-key`

Cria um novo token de acesso com limite personalizado.

**Request:**
```json
{
  "duration": 3600,
  "rate": 100
}
```

- `duration`: validade do token em segundos
- `rate`: limite de requisições por segundo para este token

**Response (201):**
```json
{
  "api-key": "uuid-gerado"
}
```

### Usando o token

Envie o token no header:

```
API_KEY: <token>
```

## Exemplos

### Teste de limitação por IP

```bash
# Enviar requisições em sequência (limite padrão: 10/s)
for i in $(seq 1 11); do
  curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/
done
# As primeiras 10 retornam 200, a 11ª retorna 429
```

### Teste de limitação por token

```bash
# Criar token com limite de 5 req/s e validade de 1 hora
TOKEN=$(curl -s -X POST http://localhost:8080/api-key \
  -H "Content-Type: application/json" \
  -d '{"duration": 3600, "rate": 5}' | jq -r '.["api-key"]')

# Enviar requisições com o token
for i in $(seq 1 6); do
  curl -s -o /dev/null -w "%{http_code}\n" \
    -H "API_KEY: $TOKEN" http://localhost:8080/
done
# As primeiras 5 retornam 200, a 6ª retorna 429
```

## Estrutura de dados no Redis

| Chave | Tipo | Descrição |
|---|---|---|
| `key:<uuid>` | String | Rate limit do token (valor = req/s, com TTL = duração) |
| `requests:ip:<ip>` | Sorted Set | Timestamps das requisições recentes por IP |
| `requests:apikey:<uuid>` | Sorted Set | Timestamps das requisições recentes por token |
| `block:ip:<ip>` | String | Flag de bloqueio do IP (TTL = tempo de bloqueio) |
| `block:apikey:<uuid>` | String | Flag de bloqueio do token (TTL = tempo de bloqueio) |
