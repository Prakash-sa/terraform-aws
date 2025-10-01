module "aws-codepipeline-lab-pipeline-role"{
    source= "./../iam_roles/aws-codepipeline-lab-pipeline-role"
}
module "variable_account"{
 source= "./../../variables"
}
locals {
    s3_name = "bucket-${module.variable_account.account_id}"
}
resource "aws_codestarconnections_connection" "CodeStar_connection" {
  name          = "CodeStar_connection"
  provider_type = "GitHub"
}

resource "aws_codepipeline" "codepipeline" {
  name     = "aws-codepipeline-lab"
  role_arn = module.aws-codepipeline-lab-pipeline-role.codepipeline_arn


  artifact_store {
    location = local.s3_name
    type     = "S3"

  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["SourceArtifact"]   # Optional output artifact
      

configuration = {
  ConnectionArn =  aws_codestarconnections_connection.CodeStar_connection.arn # Replace with your connection ARN
  FullRepositoryId = "<YOUR_GITHUB_PROFILE_NAME/REPOSITORY_NAME>"                       # Combines owner and repo
  BranchName = "eks-web-app"                                                                # Replace with your desired branch
}
    }
  }

  stage {
    name = "Build"

    action {
      name            = "Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      version         = "1"
      input_artifacts  = ["SourceArtifact"]  # Specify input artifact from previous stage
      output_artifacts = ["BuildArtifact"]   # Optional output artifact



      configuration = {
        ProjectName = "aws-codepipeline-lab-codebuild-project"
      }
    }
  }
}
