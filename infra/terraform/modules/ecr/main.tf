variable "repository_name" {
  description = "ECR repository name"
  type        = string
}

variable "image_tag_mutability" {
  description = "Tag mutability setting"
  type        = string
  default     = "MUTABLE"
}

variable "force_delete" {
  description = "Allow repository deletion even if images exist"
  type        = bool
  default     = false
}

variable "encryption_type" {
  description = "ECR encryption type"
  type        = string
  default     = "AES256"
}

variable "scan_on_push" {
  description = "Enable image scanning on push"
  type        = bool
  default     = true
}

resource "aws_ecr_repository" "this" {
  name                 = var.repository_name
  image_tag_mutability = var.image_tag_mutability
  force_delete         = var.force_delete

  image_scanning_configuration {
    scan_on_push = var.scan_on_push
  }

  encryption_configuration {
    encryption_type = var.encryption_type
  }
}

output "repository_name" {
  value       = aws_ecr_repository.this.name
  description = "ECR repository name"
}

output "repository_arn" {
  value       = aws_ecr_repository.this.arn
  description = "ECR repository ARN"
}

output "repository_url" {
  value       = aws_ecr_repository.this.repository_url
  description = "ECR repository URL"
}
