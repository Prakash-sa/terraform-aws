# Production-ready Terraform configuration for EKS deployment
terraform {
  required_version = ">= 1.5.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"
    }
  }

  # Uncomment for remote state
  # backend "s3" {
  #   bucket         = "your-terraform-state-bucket"
  #   key            = "eks/terraform.tfstate"
  #   region         = "us-east-1"
  #   encrypt        = true
  #   dynamodb_table = "terraform-state-lock"
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = var.environment
      Project     = var.project_name
      ManagedBy   = "Terraform"
      Owner       = "DevOps"
    }
  }
}

# Data source for EKS cluster authentication
data "aws_eks_cluster" "cluster" {
  name       = module.eks.cluster_name
  depends_on = [module.eks]
}

data "aws_eks_cluster_auth" "cluster" {
  name       = module.eks.cluster_name
  depends_on = [module.eks]
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.cluster.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.cluster.token
}

provider "helm" {
  kubernetes {
    host                   = data.aws_eks_cluster.cluster.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority[0].data)
    token                  = data.aws_eks_cluster_auth.cluster.token
  }
}

# S3 bucket for artifacts
module "s3" {
  source = "./modules/s3"
}

# ECR repository for container images
module "ecr_repo" {
  source = "./modules/ecr"
  
  repository_name = var.ecr_repository_name
  environment     = var.environment
}

# EKS cluster
module "eks" {
  source = "./modules/EKS"
  
  cluster_name    = var.cluster_name
  cluster_version = var.cluster_version
  environment     = var.environment
  
  vpc_cidr           = var.vpc_cidr
  azs                = var.availability_zones
  private_subnets    = var.private_subnets
  public_subnets     = var.public_subnets
  
  node_instance_types = var.node_instance_types
  node_desired_size   = var.node_desired_size
  node_min_size       = var.node_min_size
  node_max_size       = var.node_max_size
}

# CodeBuild for CI/CD
module "codebuild_project" {
  source = "./modules/codebuild"
  
  project_name = "${var.project_name}-build"
  environment  = var.environment
  
  ecr_repository_url = module.ecr_repo.repository_url
  s3_bucket_name     = module.s3.bucket_name
}

# CodePipeline for automated deployments
module "codepipeline" {
  source = "./modules/codepipeline"
  
  pipeline_name = "${var.project_name}-pipeline"
  environment   = var.environment
  
  s3_bucket_name     = module.s3.bucket_name
  codebuild_project  = module.codebuild_project.project_name
  github_repo        = var.github_repository
  github_branch      = var.github_branch
}
