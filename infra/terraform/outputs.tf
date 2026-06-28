output "ecr_repository_url" {
  description = "URL do repositório ECR — usar no `docker push` antes de cada deploy."
  value       = aws_ecr_repository.api.repository_url
}

output "function_url" {
  description = "URL HTTPS pública da API."
  value       = aws_lambda_function_url.api.function_url
}

output "lambda_function_name" {
  value = aws_lambda_function.api.function_name
}
