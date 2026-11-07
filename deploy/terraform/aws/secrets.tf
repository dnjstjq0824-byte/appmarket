# --- Secrets Manager for required API secrets (all fail closed at boot) ---
# The ECS execution role reuses AmazonECSTaskExecutionRolePolicy, which grants
# secretsmanager:GetSecretValue, so no extra IAM policy is needed.

resource "aws_secretsmanager_secret" "iam_hmac_secret" {
  name = "bn/appmarket/iam_hmac_secret"
}

resource "aws_secretsmanager_secret_version" "iam_hmac_secret" {
  secret_id     = aws_secretsmanager_secret.iam_hmac_secret.id
  secret_string = var.iam_hmac_secret
}

resource "aws_secretsmanager_secret" "distribution_token_secret" {
  name = "bn/appmarket/distribution_token_secret"
}

resource "aws_secretsmanager_secret_version" "distribution_token_secret" {
  secret_id     = aws_secretsmanager_secret.distribution_token_secret.id
  secret_string = var.distribution_token_secret
}

resource "aws_secretsmanager_secret" "pg_webhook_secret" {
  name = "bn/appmarket/pg_webhook_secret"
}

resource "aws_secretsmanager_secret_version" "pg_webhook_secret" {
  secret_id     = aws_secretsmanager_secret.pg_webhook_secret.id
  secret_string = var.pg_webhook_secret
}

resource "aws_secretsmanager_secret" "payout_enc_key" {
  name = "bn/appmarket/payout_enc_key"
}

resource "aws_secretsmanager_secret_version" "payout_enc_key" {
  secret_id     = aws_secretsmanager_secret.payout_enc_key.id
  secret_string = var.payout_enc_key
}
