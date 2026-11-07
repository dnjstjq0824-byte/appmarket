# GCS backend config template for the GCP production module.
# Copy to backend.gcs.live.hcl (gitignored by pattern), fill in, then:
#   terraform init -backend-config=backend.gcs.live.hcl
# The state bucket must already exist with versioning enabled.
bucket = "REPLACE-WITH-tfstate-bucket-name"
prefix = "appmarket/production"
