module "variable_account" {
  source = "./../../../variables"
}

# Singular Role for EKS and CodeBuild
resource "aws_iam_role" "eks_codebuild_role" {
  name = "eks-codebuild-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid    = "S01",
        Effect = "Allow",
        Action = "sts:AssumeRole",
        Principal = {
          Service = [
            "eks.amazonaws.com",        # Allow EKS to assume this role
            "codebuild.amazonaws.com"   # Allow CodeBuild to assume this role
          ]
        }
      }
    ]
  })
}

# Attach the Amazon EKS Cluster Policy to the role
resource "aws_iam_role_policy_attachment" "eks_cluster_role_policy" {
  role       = aws_iam_role.eks_codebuild_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
}

# Attach the CodeBuild policy to the role
resource "aws_iam_role_policy_attachment" "codebuild_role_policy" {
  role       = aws_iam_role.eks_codebuild_role.name
  policy_arn = "arn:aws:iam::${module.variable_account.account_id}:policy/AWSCodeBuild"  # Replace with your actual policy ARN
}

# Output the role ARN
output "eks_codebuild_role_arn" {
  value = aws_iam_role.eks_codebuild_role.arn
}
