# RinoSeller API

API do sistema RinoSeller — gestão de vendas, clientes, produtos, orçamentos e finanças de um pequeno negócio de revenda. Escrita em Go, seguindo Arquitetura Hexagonal (Ports & Adapters) com elementos de DDD tático.

## Stack

- **Go 1.25** + [Gin](https://github.com/gin-gonic/gin) (HTTP)
- **PostgreSQL** via [pgx](https://github.com/jackc/pgx) (hoje hospedado no Supabase)
- **JWT** (golang-jwt) para autenticação, bcrypt para senhas
- **Swagger/OpenAPI** gerado via [swaggo](https://github.com/swaggo/swag)
- **Docker** para build e deploy

## Arquitetura

Hexagonal: o domínio não conhece HTTP nem banco de dados. Os adapters de entrada (HTTP) e saída (Postgres) só se comunicam com o núcleo através de interfaces (ports).

```
                         ┌───────────────────────────┐
                         │   Cliente HTTP (SPA, curl,  │
                         │   Swagger UI)                │
                         └─────────────┬─────────────┘
                                       │ JSON / HTTP
┌──────────────────────────────────────▼──────────────────────────────────────┐
│ internal/adapters/in/http              Adapter de entrada (driving)         │
│ Gin: routes.go, middleware.go, *_handler.go, DTOs de request/response       │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ implementa
┌──────────────────────────────────────▼──────────────────────────────────────┐
│ internal/core/ports                    Input ports (use cases)              │
│ AuthUseCase, ProductUseCase, OrderUseCase, ClientUseCase, QuoteUseCase...    │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ implementado por
┌──────────────────────────────────────▼──────────────────────────────────────┐
│ internal/core/services                 Casos de uso (orquestração)          │
│ AuthService, ProductService, OrderService, ClientService, QuoteService...    │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ opera sobre
┌──────────────────────────────────────▼──────────────────────────────────────┐
│ internal/core/domain                   Domínio                              │
│ Entidades (Quote, Order, Client, Product...) com invariantes como métodos   │
│ Value Objects: Money, CpfCnpj, Phone                                        │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ persistido via
┌──────────────────────────────────────▼──────────────────────────────────────┐
│ internal/core/ports                    Output ports (repositórios)          │
│ UserRepository, ProductRepository, ClientPaymentRepository...               │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ implementa
┌──────────────────────────────────────▼──────────────────────────────────────┐
│ internal/adapters/out/repository       Adapter de saída (driven)            │
│ PostgresXxxRepository (pgx)                                                 │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
                              ┌────────▼────────┐
                              │   PostgreSQL     │
                              │   (Supabase)     │
                              └─────────────────┘
```

A composição (wiring manual de todas as dependências) acontece em `cmd/server/main.go` — não há container de DI nem reflection, é só código explícito.

### Domínio

| Entidade | Responsabilidade |
|---|---|
| `User` | Conta de acesso (admin/seller), autenticação JWT |
| `Product` | Catálogo, incluindo kits (produto composto por outros produtos) |
| `Order` | Pedido do catálogo público (fluxo simples, baixa de estoque imediata) |
| `Quote` | Orçamento/pedido de venda, com máquina de estados (`Aguardando Aprovação → Aprovado → Faturado(/Gradual) → Entregue`, ou `→ Cancelado`) |
| `Client` | Cliente, com controle de limite de dívida e histórico de pagamentos |
| `ClientPayment` | Recebimento registrado para um cliente |
| `Expense` | Conta a pagar (compra de fornecedor, despesa) |
| `CapitalContribution` | Aporte ou retirada manual de capital |

Os invariantes de negócio vivem nas próprias entidades (`Quote.Approve()`, `Client.EnsureCanReceiveNewOrder()`, `Expense.Pay()` etc.), não nos services — os services orquestram, as entidades protegem suas próprias regras.

## Estrutura de diretórios

```
cmd/
  server/        # entrypoint da API (HTTP)
  migrate/        # entrypoint que aplica o schema (roda separado do boot)
internal/
  core/
    domain/        # entidades, value objects, erros de domínio — zero dependências externas
    ports/          # interfaces (use cases + repositórios)
    services/       # implementação dos use cases
  adapters/
    in/http/        # Gin: rotas, middleware, handlers, DTOs
    out/database/   # conexão pgx, schema SQL
    out/repository/ # implementação Postgres dos ports de repositório
docs/             # Swagger gerado (swag init) — não editar manualmente
Dockerfile
railway.toml      # config de build/healthcheck do Railway
```

## Setup

### Pré-requisitos

- Go 1.25+
- Docker (para build/teste de imagem, ou para rodar Postgres localmente)
- Uma instância Postgres (local via Docker, ou Supabase)

### Variáveis de ambiente

Copie `.env.example` para `.env` e preencha:

```bash
cp .env.example .env
```

| Variável | Obrigatória | Descrição |
|---|---|---|
| `DATABASE_URL` | sim | Connection string do Postgres |
| `JWT_SECRET` | sim | Secret para assinar os JWTs — a API **falha no boot** se não estiver definida (gerar com `openssl rand -hex 32`) |
| `PORT` | não (default `8080`) | Porta HTTP |
| `ALLOWED_ORIGINS` | não | Origens extras de CORS, separadas por vírgula (localhost já é permitido por padrão) |

## Rodando localmente

```bash
# 1. Suba um Postgres (se não tiver um já rodando)
docker run -d --name rinoseller-pg -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=rinoseller -p 5432:5432 postgres:16-alpine

# 2. Configure o .env (DATABASE_URL apontando para o passo 1)
cp .env.example .env

# 3. Aplique o schema (uma vez, e sempre que o schema mudar)
go run ./cmd/migrate

# 4. Rode a API
go run ./cmd/server
```

A API sobe em `http://localhost:8080`. No primeiro boot, se não existir nenhum admin, um é criado com senha aleatória — a senha aparece **uma única vez** no log (`✓ Admin padrão criado: admin@rinoseller.com / senha temporária: ...`). Troque-a após o primeiro login.

- Swagger UI: `http://localhost:8080/docs/index.html`
- Health check: `http://localhost:8080/health`

## Rodando via Docker

```bash
docker build -t rinoseller-api .

docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgres://..." \
  -e JWT_SECRET="..." \
  rinoseller-api
```

## Checks de qualidade

```bash
go build ./...
go vet ./...
gofmt -l .
golangci-lint run ./...   # brew install golangci-lint
```

## Deploy (Railway)

O Railway detecta o `Dockerfile` automaticamente. `railway.toml` define o builder e o healthcheck (`/health`).

No painel do Railway, configurar as variáveis de ambiente do serviço:

| Variável | Valor |
|---|---|
| `DATABASE_URL` | Connection string do Postgres (pooler do Supabase) |
| `JWT_SECRET` | Secret de produção dos JWTs |

Rodar a migração uma vez contra o banco de produção (`DATABASE_URL=... go run ./cmd/migrate`) antes do primeiro deploy.
