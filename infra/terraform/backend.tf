# State remoto no S3 com lock nativo (Terraform >= 1.10, sem precisar de
# tabela DynamoDB). O bucket precisa existir antes do primeiro `terraform
# init` — veja README.md para o comando de criação.
terraform {
  backend "s3" {
    bucket       = "rinoseller-terraform-state"
    key          = "api/terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true
    encrypt      = true
  }
}
