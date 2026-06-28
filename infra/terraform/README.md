# Deploy da API na AWS (Lambda + ECR via Terraform)

Infra como código para rodar a API RinoSeller em AWS Lambda (container image),
usando o [AWS Lambda Web Adapter](https://github.com/awslabs/aws-lambda-web-adapter)
— o binário Go não tem nenhum código específico de Lambda; o mesmo Dockerfile
(`docker build --target runtime`) também roda em EC2/ECS sem alterações.

Nada disso foi aplicado ainda. Esses passos exigem AWS CLI configurado com
credenciais de uma conta AWS real.

## 0. Pré-requisitos (uma única vez)

```bash
# Instalar AWS CLI e Terraform (macOS)
brew install awscli terraform

# Configurar credenciais
aws configure

# Criar o bucket do state remoto (nome em backend.tf) — precisa existir
# antes do primeiro `terraform init`
aws s3api create-bucket --bucket rinoseller-terraform-state --region us-east-1
aws s3api put-bucket-versioning --bucket rinoseller-terraform-state \
  --versioning-configuration Status=Enabled
```

## 1. Configurar variáveis

```bash
cd infra/terraform
cp terraform.tfvars.example terraform.tfvars
# editar terraform.tfvars com DATABASE_URL, JWT_SECRET (openssl rand -hex 32) etc.
```

## 2. Provisionar o ECR (primeira vez)

O Lambda exige que a imagem já exista no ECR antes de poder ser criado —
então a primeira aplicação só sobe o repositório ECR:

```bash
terraform init
terraform apply -target=aws_ecr_repository.api
```

## 3. Build e push da imagem

```bash
cd ../..  # raiz do repo
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin <ECR_REPOSITORY_URL>

docker build --target lambda -t rinoseller-api .
docker tag rinoseller-api:latest <ECR_REPOSITORY_URL>:latest
docker push <ECR_REPOSITORY_URL>:latest
```

(`<ECR_REPOSITORY_URL>` é a saída `ecr_repository_url` do passo 2.)

## 4. Provisionar o restante (Lambda, IAM, logs, function URL)

```bash
cd infra/terraform
terraform apply
```

A saída `function_url` é o endpoint HTTPS público da API.

## 5. Rodar a migração do banco

A migração não roda no boot do Lambda (ver `cmd/migrate`). Rodar uma única
vez, de qualquer máquina com acesso ao banco:

```bash
DATABASE_URL="<a mesma string usada em terraform.tfvars>" go run ./cmd/migrate
```

## Deploys seguintes

Repetir só os passos 3 e 4 (build, push, `terraform apply` — o Terraform
detecta a nova tag/imagem e atualiza a função). Usar uma tag única por deploy
(ex: o SHA do commit) em vez de `latest` evita ambiguidade sobre qual versão
está no ar; nesse caso, atualize `image_tag` em `terraform.tfvars` antes do
apply.

## Migrar para EC2/ECS no futuro

Sem mudar uma linha de Go: `docker build --target runtime` gera a mesma
aplicação sem a camada do Lambda Web Adapter. Substituir os recursos deste
diretório (`aws_lambda_function`, `aws_lambda_function_url`) por
`aws_ecs_service`/`aws_instance` apontando para essa imagem.
