resource "aws_iam_openid_connect_provider" "github" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = ["6938fd4d98bab03faadb97b34396831e3780aea1"]
}

data "aws_iam_policy_document" "github_actions_trust" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github.arn]
    }
    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }
    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["repo:${var.github_repo}:ref:refs/heads/main"]
    }
  }
}

resource "aws_iam_role" "github_actions_deploy" {
  name               = "${var.project_name}-github-actions"
  assume_role_policy = data.aws_iam_policy_document.github_actions_trust.json
}

data "aws_iam_policy_document" "github_actions_permissions" {
  statement {
    sid       = "ECRAuth"
    actions   = ["ecr:GetAuthorizationToken"]
    resources = ["*"]
  }
  statement {
    sid       = "ECRRepository"
    actions   = ["ecr:*"]
    resources = [aws_ecr_repository.api.arn]
  }
  statement {
    sid       = "Lambda"
    actions   = ["lambda:*"]
    resources = ["*"]
  }
  # API Gateway v2 não suporta escopo por ARN de recurso na maioria das
  # ações de gerenciamento (CreateApi, tagging) — "*" é o que a própria AWS
  # recomenda nos exemplos de policy para esse serviço.
  statement {
    sid       = "ApiGateway"
    actions   = ["apigateway:*"]
    resources = ["*"]
  }
  statement {
    sid = "IAMForOwnRoles"
    actions = [
      "iam:CreateRole", "iam:GetRole", "iam:DeleteRole", "iam:UpdateRole",
      "iam:TagRole", "iam:UntagRole",
      "iam:PutRolePolicy", "iam:GetRolePolicy", "iam:DeleteRolePolicy",
      "iam:ListRolePolicies", "iam:ListAttachedRolePolicies", "iam:ListInstanceProfilesForRole",
      "iam:PassRole",
    ]
    resources = ["arn:aws:iam::*:role/${var.project_name}-*"]
  }
  statement {
    # DescribeLogGroups usa filtro por prefixo, não por ARN exato — IAM exige
    # Resource "*" para essa ação specific, mesmo restringindo as outras.
    sid = "Logs"
    actions = [
      "logs:CreateLogGroup", "logs:DeleteLogGroup", "logs:DescribeLogGroups",
      "logs:PutRetentionPolicy", "logs:TagResource", "logs:UntagResource",
      "logs:ListTagsForResource", "logs:ListTagsLogGroup",
    ]
    resources = ["*"]
  }
  statement {
    sid       = "TerraformStateBucket"
    actions   = ["s3:*"]
    resources = ["arn:aws:s3:::rinoseller-terraform-state", "arn:aws:s3:::rinoseller-terraform-state/*"]
  }
  statement {
    sid       = "WhoAmI"
    actions   = ["sts:GetCallerIdentity"]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "github_actions_deploy" {
  name   = "${var.project_name}-github-actions-deploy"
  role   = aws_iam_role.github_actions_deploy.id
  policy = data.aws_iam_policy_document.github_actions_permissions.json
}
