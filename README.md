# RinoSeller API

API do sistema RinoSeller — gestão de vendas, clientes, produtos, orçamentos e finanças de um pequeno negócio de revenda. Escrita em Go, seguindo Arquitetura Hexagonal (Ports & Adapters) com elementos de DDD tático.

## Stack

- **Go 1.25** + [Gin](https://github.com/gin-gonic/gin) (HTTP)
- **PostgreSQL** via [pgx](https://github.com/jackc/pgx) (hoje hospedado no Supabase)
- **JWT** (golang-jwt) para autenticação, bcrypt para senhas
- **Swagger/OpenAPI** gerado via [swaggo](https://github.com/swaggo/swag)
- **Docker** multi-stage (mesma imagem roda em AWS Lambda, ECS, EC2 ou App Runner)
- **Terraform** para a infraestrutura AWS (`infra/terraform/`)

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

### Deploy: um binário, múltiplos ambientes

```
                    ┌─────────────────────┐
                    │   Dockerfile único    │
                    │  (stage "builder")    │
                    └──────────┬───────────┘
                ┌──────────────┴──────────────┐
                ▼                              ▼
   ┌─────────────────────────┐    ┌─────────────────────────────┐
   │  target "runtime"        │    │  target "lambda"              │
   │  binário Go puro          │    │  binário Go + AWS Lambda      │
   │  (EC2 / ECS / App Runner) │    │  Web Adapter (extension)      │
   └─────────────────────────┘    └─────────────────────────────┘
```

O binário Go é **idêntico** nos dois casos — nenhum código depende do SDK ou dos eventos do Lambda. O [AWS Lambda Web Adapter](https://github.com/awslabs/aws-lambda-web-adapter) traduz invocações do Lambda em requisições HTTP normais contra o mesmo servidor Gin. Trocar de Lambda para ECS/EC2 no futuro é só usar o outro target do Dockerfile — zero mudança de código.

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
infra/terraform/  # infraestrutura AWS (ECR, Lambda, IAM, logs)
Dockerfile        # multi-stage: targets "runtime" e "lambda"
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
# build (target "runtime", o mesmo usado em EC2/ECS/App Runner)
docker build --target runtime -t rinoseller-api:runtime .

# build do target Lambda (com o AWS Lambda Web Adapter embutido)
docker build --target lambda -t rinoseller-api:lambda .

# rodar o target runtime localmente
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgres://..." \
  -e JWT_SECRET="..." \
  rinoseller-api:runtime
```

## Checks de qualidade

```bash
go build ./...
go vet ./...
gofmt -l .
golangci-lint run ./...   # brew install golangci-lint
```

## Deploy na AWS

Infraestrutura como código em `infra/terraform/` (ECR, Lambda, IAM, CloudWatch Logs, Function URL). Passo a passo completo — criação do bucket de state, build/push da imagem, `terraform apply` — está em [`infra/terraform/README.md`](infra/terraform/README.md).

Resumo do fluxo:
1. Provisionar o ECR.
2. Build da imagem com `--target lambda`, push para o ECR.
3. `terraform apply` cria a função Lambda apontando para a imagem.
4. Rodar `cmd/migrate` uma vez contra o banco de produção.

O banco de dados continua no Supabase nesta fase — a migração para RDS/outro provedor, se vier a acontecer, é um passo independente deste deploy.
