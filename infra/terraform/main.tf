resource "aws_ecr_repository" "api" {
  name                 = var.project_name
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

# Mantém só as últimas 10 imagens — evita acumular custo de storage no ECR
# com builds antigos que nunca mais serão referenciados por um deploy.
resource "aws_ecr_lifecycle_policy" "api" {
  repository = aws_ecr_repository.api.name

  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Manter apenas as 10 imagens mais recentes"
      selection = {
        tagStatus   = "any"
        countType   = "imageCountMoreThan"
        countNumber = 10
      }
      action = { type = "expire" }
    }]
  })
}

resource "aws_cloudwatch_log_group" "api" {
  name              = "/aws/lambda/${var.project_name}"
  retention_in_days = var.log_retention_days
}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda_exec" {
  name               = "${var.project_name}-lambda-exec"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

# Permissão mínima: só escrever logs no próprio log group. Nenhuma permissão
# adicional (S3, DynamoDB etc.) é concedida — o princípio de menor privilégio
# evita que um bug na API se transforme num acesso indevido a outros recursos
# da conta AWS.
data "aws_iam_policy_document" "lambda_logs" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = ["${aws_cloudwatch_log_group.api.arn}:*"]
  }
}

resource "aws_iam_role_policy" "lambda_logs" {
  name   = "${var.project_name}-lambda-logs"
  role   = aws_iam_role.lambda_exec.id
  policy = data.aws_iam_policy_document.lambda_logs.json
}

resource "aws_lambda_function" "api" {
  function_name = var.project_name
  role          = aws_iam_role.lambda_exec.arn

  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.api.repository_url}:${var.image_tag}"
  architectures = ["x86_64"]

  memory_size = 512
  timeout     = 30

  environment {
    variables = {
      DATABASE_URL    = var.database_url
      JWT_SECRET      = var.jwt_secret
      ALLOWED_ORIGINS = var.allowed_origins
      # Lambda Web Adapter: porta em que o servidor Go (Gin) escuta.
      AWS_LWA_PORT = "8080"
    }
  }

  depends_on = [aws_iam_role_policy.lambda_logs, aws_cloudwatch_log_group.api]
}

resource "aws_apigatewayv2_api" "api" {
  name          = var.project_name
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_integration" "lambda" {
  api_id                 = aws_apigatewayv2_api.api.id
  integration_type       = "AWS_PROXY"
  integration_uri        = aws_lambda_function.api.invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "default" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "$default"
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.api.id
  name        = "$default"
  auto_deploy = true
}

resource "aws_lambda_permission" "apigateway_invoke" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*/*"
}
