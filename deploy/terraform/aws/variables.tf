variable "aws_region" {
  type    = string
  default = "ap-northeast-2"
}

variable "environment" {
  type    = string
  default = "production"
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

variable "container_image" {
  type    = string
  default = "ghcr.io/bookingnetwork/appmarket-api:latest"
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "redis_auth_token" {
  type      = string
  sensitive = true
}

# The API exits at boot unless every one of these is non-empty.
variable "iam_hmac_secret" {
  type        = string
  sensitive   = true
  description = "IAM_HMAC_SECRET: 32+ random chars for blind indexes and dual-control HMACs."
}

variable "distribution_token_secret" {
  type        = string
  sensitive   = true
  description = "DISTRIBUTION_TOKEN_SECRET: 32+ random chars for HMAC download tokens."
}

variable "pg_webhook_secret" {
  type        = string
  sensitive   = true
  description = "PG_WEBHOOK_SECRET: HMAC secret shared with the payment gateway."
}

variable "payout_enc_key" {
  type        = string
  sensitive   = true
  description = "PAYOUT_ENC_KEY: exactly 32 bytes for payout account encryption."

  validation {
    condition     = length(var.payout_enc_key) == 32
    error_message = "PAYOUT_ENC_KEY must be exactly 32 bytes (the API refuses to boot otherwise)."
  }
}

