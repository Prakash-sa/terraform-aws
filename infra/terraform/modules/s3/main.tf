variable "bucket_prefix" {
  description = "Prefix for the artifacts bucket"
  type        = string
  default     = "go-api-artifacts"
}

variable "force_destroy" {
  description = "Force destroy the bucket (including objects)"
  type        = bool
  default     = false
}

resource "random_id" "suffix" {
  byte_length = 4
}

resource "aws_s3_bucket" "artifacts" {
  bucket        = format("%s-%s", var.bucket_prefix, random_id.suffix.hex)
  force_destroy = var.force_destroy
}

resource "aws_s3_bucket_versioning" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

output "bucket_name" {
  value       = aws_s3_bucket.artifacts.bucket
  description = "Artifacts bucket name"
}

output "bucket_arn" {
  value       = aws_s3_bucket.artifacts.arn
  description = "Artifacts bucket ARN"
}
