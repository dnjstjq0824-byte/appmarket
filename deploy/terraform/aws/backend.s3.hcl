# S3 backend config template for the AWS production module.
# Copy to backend.s3.live.hcl (gitignored by pattern below), fill in, then:
#   terraform init -backend-config=backend.s3.live.hcl
# The state bucket must already exist with versioning enabled.
bucket         = "REPLACE-WITH-tfstate-bucket-name"
key            = "appmarket/production/terraform.tfstate"
region         = "ap-northeast-2"
encrypt        = true
dynamodb_table = "REPLACE-WITH-tflock-table-name"
