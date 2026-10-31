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
  name          = "bn-quarantine-artifacts-${var.gcp_project_id}"
  location      = var.gcp_region
  force_destroy = false
  uniform_bucket_level_access = true
}

resource "google_storage_bucket" "public_cdn" {
  name          = "bn-public-cdn-artifacts-${var.gcp_project_id}"
  location      = var.gcp_region
  force_destroy = false
  uniform_bucket_level_access = true
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

