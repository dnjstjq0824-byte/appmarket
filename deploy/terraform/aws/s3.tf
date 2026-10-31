# --- S3 Buckets for Quarantine & Public CDN Distribution ---

resource "aws_s3_bucket" "quarantine" {
  bucket = "bn-quarantine-artifacts-${var.environment}"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "quarantine" {
  bucket = aws_s3_bucket.quarantine.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "quarantine" {
  bucket                  = aws_s3_bucket.quarantine.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket" "public_cdn" {
  bucket = "bn-public-cdn-artifacts-${var.environment}"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "public_cdn" {
  bucket = aws_s3_bucket.public_cdn.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "public_cdn" {
  bucket                  = aws_s3_bucket.public_cdn.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

