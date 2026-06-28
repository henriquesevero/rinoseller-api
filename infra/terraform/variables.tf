variable "aws_region" {
  description = "Região AWS onde os recursos são criados."
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Prefixo usado no nome dos recursos."
  type        = string
  default     = "rinoseller-api"
}

variable "image_tag" {
  description = "Tag da imagem no ECR a ser usada pela função Lambda (ex: a SHA do commit). Atualizada a cada deploy."
  type        = string
  default     = "latest"
}

variable "database_url" {
  description = "Connection string do Postgres (Supabase, por enquanto). Nunca commitar o valor real — passar via terraform.tfvars (gitignored) ou -var na CLI."
  type        = string
  sensitive   = true
}

variable "jwt_secret" {
  description = "Secret usado para assinar os JWTs emitidos pela API. Gerar com 'openssl rand -hex 32'."
  type        = string
  sensitive   = true
}

variable "allowed_origins" {
  description = "Origens permitidas no CORS, além de localhost (ex: https://app.rinoseller.com.br)."
  type        = string
  default     = ""
}

variable "log_retention_days" {
  description = "Retenção dos logs no CloudWatch — limitar evita custo de armazenamento crescendo indefinidamente."
  type        = number
  default     = 14
}

variable "github_repo" {
  description = "Repositório GitHub (owner/repo) autorizado a assumir a role de deploy via OIDC."
  type        = string
}
