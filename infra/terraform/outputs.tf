output "ecr_repository_url" {
  description = "URL do repositório ECR — usar no `docker push` antes de cada deploy."
  value       = aws_ecr_repository.api.repository_url
}

output "api_url" {
  description = "URL HTTPS pública da API (API Gateway)."
  value       = aws_apigatewayv2_api.api.api_endpoint
}

output "lambda_function_name" {
  description = "Nome da função Lambda."
  value       = aws_lambda_function.api.function_name
}

output "github_actions_role_arn" {
  description = "Role ARN que o workflow do GitHub Actions assume via OIDC."
  value       = aws_iam_role.github_actions_deploy.arn
}
