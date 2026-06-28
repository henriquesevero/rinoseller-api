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

  package_type = "Image"
  image_uri    = "${aws_ecr_repository.api.repository_url}:${var.image_tag}"

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

# URL pública e gratuita (sem custo de API Gateway). authorization_type =
# NONE porque a API faz sua própria autenticação via JWT — quem chama sem
# token recebe 401 da própria aplicação, não do Lambda. Sem bloco `cors`
# aqui de propósito: CORS é tratado uma única vez, pelo middleware Gin
# (ALLOWED_ORIGINS), para ter o mesmo comportamento em Lambda/ECS/EC2 — uma
# configuração de CORS na Function URL divergiria da regra real da app.
resource "aws_lambda_function_url" "api" {
  function_name      = aws_lambda_function.api.function_name
  authorization_type = "NONE"
}
