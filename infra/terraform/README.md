# Deploy da API na AWS (Lambda + ECR + API Gateway via Terraform)

Infra como código para rodar a API RinoSeller em AWS Lambda (container image),
usando o [AWS Lambda Web Adapter](https://github.com/awslabs/aws-lambda-web-adapter)
— o binário Go não tem nenhum código específico de Lambda; o mesmo Dockerfile
(`docker build --target runtime`) também roda em EC2/ECS sem alterações.

A função é exposta publicamente via **API Gateway HTTP API** (não Function
URL — ver nota sobre restrição de conta nova ao final).

A função Lambda **precisa ser criada em `us-east-1`** (North Virginia).

## 0. Pré-requisitos (uma única vez)

```bash
brew install awscli terraform

aws configure

aws s3api create-bucket --bucket rinoseller-terraform-state --region us-east-1
aws s3api put-bucket-versioning --bucket rinoseller-terraform-state \
  --versioning-configuration Status=Enabled
aws s3api put-bucket-encryption --bucket rinoseller-terraform-state \
  --server-side-encryption-configuration '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'
aws s3api put-public-access-block --bucket rinoseller-terraform-state \
  --public-access-block-configuration BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
```

## 1. Configurar variáveis

```bash
cd infra/terraform
cp terraform.tfvars.example terraform.tfvars
# editar terraform.tfvars: DATABASE_URL (pooler do Supabase, porta 6543),
# JWT_SECRET (openssl rand -hex 32), github_repo (owner/repo no GitHub)
```

## 2. Provisionar o ECR (primeira vez)

O Lambda exige que a imagem já exista no ECR antes de poder ser criado —
então a primeira aplicação só sobe o repositório ECR:

```bash
terraform init
terraform apply -target=aws_ecr_repository.api -target=aws_ecr_lifecycle_policy.api
```

## 3. Build e push da imagem

```bash
cd ../..  # raiz do repo
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin <ECR_REPOSITORY_URL>

# --platform linux/amd64 é necessário em Mac Apple Silicon — a função Lambda
# está configurada para x86_64 (arquitetura padrão dos runners do GitHub
# Actions, que fazem o deploy contínuo depois do setup inicial).
# --provenance=false --sbom=false evita o manifest OCI de attestation que o
# Lambda não aceita ("image manifest ... not supported").
docker build --platform linux/amd64 --provenance=false --sbom=false \
  --target lambda -t <ECR_REPOSITORY_URL>:latest .
docker push <ECR_REPOSITORY_URL>:latest
```

(`<ECR_REPOSITORY_URL>` é a saída `ecr_repository_url` do passo 2.)

## 4. Provisionar o restante (Lambda, IAM, logs, API Gateway, role do GitHub Actions)

```bash
cd infra/terraform
terraform apply
```

A saída `api_url` é o endpoint HTTPS público da API.

## 5. Rodar a migração do banco

A migração não roda no boot do Lambda (ver `cmd/migrate`). Rodar uma única
vez, de qualquer máquina com acesso ao banco:

```bash
DATABASE_URL="<a mesma string usada em terraform.tfvars>" go run ./cmd/migrate
```

## Deploys seguintes

Depois do setup inicial (passos 0–5), os deploys seguintes acontecem
automaticamente via GitHub Actions a cada push em `main` — ver a seção
"CI/CD" no README principal do repositório. Não é necessário repetir os
passos manuais acima.

Para forçar um deploy manual: `Actions` → workflow `Deploy` → `Run workflow`.

## Migrar para EC2/ECS no futuro

Sem mudar uma linha de Go: `docker build --target runtime` gera a mesma
aplicação sem a camada do Lambda Web Adapter. Substituir os recursos deste
diretório (`aws_lambda_function`, `aws_apigatewayv2_*`) por
`aws_ecs_service`/`aws_instance` apontando para essa imagem.

## Nota: restrição de conta nova da AWS

Contas AWS recém-criadas têm uma restrição anti-fraude padrão que bloqueia
a criação de recursos públicos (Lambda Function URL, e aparentemente também
`apigatewayv2:CreateApi` e `iam:CreateOpenIDConnectProvider`) com
`AccessDeniedException`, mesmo com a IAM policy correta — é uma camada
acima do IAM, só removida pelo AWS Support (gratuito, categoria "Account
and Billing"). Se `terraform apply` falhar nesses recursos especificamente
com `AccessDenied`/`AccessDeniedException` apesar da policy estar certa,
essa é a causa — não há ajuste de Terraform ou IAM que resolva por este
lado. Depois da liberação pelo suporte, `terraform apply` de novo resolve
sem mudança de código.
