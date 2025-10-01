provider "aws"{
    region="us-east-1"
}
module "s3"{
    source= "./modules/s3"
}
module "eks" {
  source = "./modules/EKS"
}
module "ecr_repo" {
  source = "./modules/ecr"
}
module "codebuild_project"{
    source= "./modules/codebuild"
}
module "codepipeline"{
    source= "./modules/codepipeline"
}
