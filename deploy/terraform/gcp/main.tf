# ==============================================================================
# Terraform GCP Production Infrastructure Blueprint
# BookingNetwork App Market — Sovereign Multi-Region Cloud Deployment
# ==============================================================================

terraform {
  required_version = ">= 1.7.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.20.0"
    }
  }
  # Remote state. Copy backend.gcs.hcl to backend.gcs.live.hcl, fill in the
  # placeholders (never commit real values), then:
  #   terraform init -backend-config=backend.gcs.live.hcl
  backend "gcs" {}
}

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

# --- Cloud Run v2 High Availability Service ---
resource "google_cloud_run_v2_service" "api" {
  name     = "bn-appmarket-api-${var.environment}"
  location = var.gcp_region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    scaling {
      min_instance_count = 3
      max_instance_count = 20
    }

    containers {
      image = var.container_image
      ports {
        container_port = 8080
      }
      resources {
        limits = {
          cpu    = "2"
          memory = "2Gi"
        }
      }
      env {
        name  = "ALLOWED_ORIGIN"
        value = "https://appmarket.dnjstjqppk.workers.dev"
      }
      env {
        name  = "TRUSTED_PROXY_CIDRS"
        value = "10.0.0.0/8"
      }
      env {
        name  = "OPENSEARCH_URL"
        value = "https://search.appmarket.local:9200"
      }
      env {
        name  = "S3_HEALTH_URL"
        value = "https://storage.googleapis.com"
      }
      env {
        name  = "OBJECT_STORAGE"
        value = "s3"
      }
      env {
        name  = "S3_ENDPOINT"
        value = "https://storage.googleapis.com"
      }
      env {
        name  = "S3_REGION"
        value = var.gcp_region
      }
      env {
        name  = "S3_USE_PATH_STYLE"
        value = "false"
      }
      env {
        name  = "MEDIA_BUCKET"
        value = google_storage_bucket.public_cdn.name
      }
      env {
        name  = "QUARANTINE_BUCKET"
        value = google_storage_bucket.quarantine.name
      }
      env {
        name = "DATABASE_URL"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.db_url.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "REDIS_URL"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.redis_url.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "IAM_HMAC_SECRET"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.iam_hmac_secret.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "DISTRIBUTION_TOKEN_SECRET"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.distribution_token_secret.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "PG_WEBHOOK_SECRET"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.pg_webhook_secret.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "PAYOUT_ENC_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.payout_enc_key.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "S3_ACCESS_KEY_ID"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.s3_access_key_id.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "S3_SECRET_ACCESS_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.s3_secret_access_key.secret_id
            version = "latest"
          }
        }
      }
      liveness_probe {
        http_get {
          path = "/healthz/live"
          port = 8080
        }
        initial_delay_seconds = 10
        period_seconds        = 15
      }
      startup_probe {
        http_get {
          path = "/healthz/ready"
          port = 8080
        }
        initial_delay_seconds = 5
        period_seconds        = 5
        failure_threshold     = 6
      }
    }
  }
}

resource "google_cloud_run_service_iam_binding" "public_access" {
  location = google_cloud_run_v2_service.api.location
  service  = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  members  = ["allUsers"]
}

# --- Cloud SQL PostgreSQL 16 (HA Multi-Zone) ---
resource "google_sql_database_instance" "postgres" {
  name             = "bn-postgres-instance-${var.environment}"
  database_version = "POSTGRES_16"
  region           = var.gcp_region

  settings {
    tier              = "db-custom-4-16384"
    availability_type = "REGIONAL"
    disk_size         = 100
    disk_type         = "PD_SSD"
    disk_autoresize   = true

    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = true
      start_time                     = "03:00"
    }

    ip_configuration {
      ipv4_enabled = true
    }
  }
}

resource "google_sql_database" "db" {
  name     = "bn_appmarket"
  instance = google_sql_database_instance.postgres.name
}

resource "google_sql_user" "app_user" {
  name     = "bn_app"
  instance = google_sql_database_instance.postgres.name
  password = var.db_password
}

# --- Memorystore Redis 7 (Standard HA) ---
resource "google_redis_instance" "redis" {
  name           = "bn-redis-instance-${var.environment}"
  tier           = "STANDARD_HA"
  memory_size_gb = 4
  region         = var.gcp_region
  redis_version  = "REDIS_7_0"
  auth_enabled   = true
}

# --- GCS Buckets for Quarantine & Public Delivery ---
resource "google_storage_bucket" "quarantine" {
  name                        = "bn-quarantine-artifacts-${var.gcp_project_id}"
  location                    = var.gcp_region
  force_destroy               = false
  uniform_bucket_level_access = true
}

resource "google_storage_bucket" "public_cdn" {
  name                        = "bn-public-cdn-artifacts-${var.gcp_project_id}"
  location                    = var.gcp_region
  force_destroy               = false
  uniform_bucket_level_access = true
}

resource "google_storage_bucket" "kyb_vault" {
  name                        = "bn-kyb-vault-${var.gcp_project_id}"
  location                    = var.gcp_region
  force_destroy               = false
  uniform_bucket_level_access = true
}

# Cloud Run runs as the Compute Engine default service account unless a custom
# SA is attached. It needs Secret Manager read access (all mounted secrets)
# and object admin on the three app buckets (S3-compatible GCS access).
data "google_project" "current" {}

locals {
  cloud_run_sa = "serviceAccount:${data.google_project.current.number}-compute@developer.gserviceaccount.com"
}

resource "google_project_iam_member" "run_secrets_reader" {
  project = var.gcp_project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = local.cloud_run_sa
}

resource "google_storage_bucket_iam_member" "quarantine_objects" {
  bucket = google_storage_bucket.quarantine.name
  role   = "roles/storage.objectAdmin"
  member = local.cloud_run_sa
}

resource "google_storage_bucket_iam_member" "public_cdn_objects" {
  bucket = google_storage_bucket.public_cdn.name
  role   = "roles/storage.objectAdmin"
  member = local.cloud_run_sa
}

resource "google_storage_bucket_iam_member" "kyb_vault_objects" {
  bucket = google_storage_bucket.kyb_vault.name
  role   = "roles/storage.objectAdmin"
  member = local.cloud_run_sa
}

# HMAC key for the S3-compatible GCS endpoint (consumed via Secret Manager).
resource "google_storage_hmac_key" "run_s3" {
  service_account_email = "${data.google_project.current.number}-compute@developer.gserviceaccount.com"
}

resource "google_secret_manager_secret" "s3_access_key_id" {
  secret_id = "bn-s3-access-key-id"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "s3_access_key_id" {
  secret      = google_secret_manager_secret.s3_access_key_id.id
  secret_data = google_storage_hmac_key.run_s3.access_id
}

resource "google_secret_manager_secret" "s3_secret_access_key" {
  secret_id = "bn-s3-secret-access-key"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "s3_secret_access_key" {
  secret      = google_secret_manager_secret.s3_secret_access_key.id
  secret_data = google_storage_hmac_key.run_s3.secret
}

# --- Secret Manager ---
resource "google_secret_manager_secret" "db_url" {
  secret_id = "bn-database-url"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "db_url" {
  secret      = google_secret_manager_secret.db_url.id
  secret_data = "postgres://bn_app:${var.db_password}@${google_sql_database_instance.postgres.public_ip_address}:5432/bn_appmarket?sslmode=require"
}

resource "google_secret_manager_secret" "redis_url" {
  secret_id = "bn-redis-url"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "redis_url" {
  secret      = google_secret_manager_secret.redis_url.id
  secret_data = "redis://:${google_redis_instance.redis.auth_string}@${google_redis_instance.redis.host}:${google_redis_instance.redis.port}/0"
}

# The API exits at boot unless every one of these is non-empty.
resource "google_secret_manager_secret" "iam_hmac_secret" {
  secret_id = "bn-iam-hmac-secret"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "iam_hmac_secret" {
  secret      = google_secret_manager_secret.iam_hmac_secret.id
  secret_data = var.iam_hmac_secret
}

resource "google_secret_manager_secret" "distribution_token_secret" {
  secret_id = "bn-distribution-token-secret"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "distribution_token_secret" {
  secret      = google_secret_manager_secret.distribution_token_secret.id
  secret_data = var.distribution_token_secret
}

resource "google_secret_manager_secret" "pg_webhook_secret" {
  secret_id = "bn-pg-webhook-secret"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "pg_webhook_secret" {
  secret      = google_secret_manager_secret.pg_webhook_secret.id
  secret_data = var.pg_webhook_secret
}

resource "google_secret_manager_secret" "payout_enc_key" {
  secret_id = "bn-payout-enc-key"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "payout_enc_key" {
  secret      = google_secret_manager_secret.payout_enc_key.id
  secret_data = var.payout_enc_key
}

variable "gcp_project_id" {
  type    = string
  default = "bookingnetwork-prod"
}

variable "gcp_region" {
  type    = string
  default = "asia-northeast3"
}

variable "environment" {
  type    = string
  default = "production"
}

variable "container_image" {
  type    = string
  default = "gcr.io/bookingnetwork-prod/appmarket-api:latest"
}

variable "db_password" {
  type      = string
  sensitive = true
}

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

output "cloud_run_url" {
  description = "Public Cloud Run URL. API base for smoke tests."
  value       = google_cloud_run_v2_service.api.uri
}

output "sql_ip" {
  description = "Cloud SQL public IP."
  value       = google_sql_database_instance.postgres.public_ip_address
}

output "redis_host" {
  description = "Memorystore host."
  value       = google_redis_instance.redis.host
}

