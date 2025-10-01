resource "aws_iam_role" "eks_cluster_role" {
  name = "eks-cluster-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "eks.amazonaws.com"
      },
      Effect = "Allow",
      Sid    = "S01"
    }]
  })
}

# Attach the Amazon EKS Cluster Policy to the cluster role
resource "aws_iam_role_policy_attachment" "eks_cluster_role_policy" {
  role       = aws_iam_role.eks_cluster_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
}

# Attach the Amazon EC2 Full Access Policy to the cluster role
resource "aws_iam_role_policy_attachment" "eks_cluster_role_ec2_policy" {
  role       = aws_iam_role.eks_cluster_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2FullAccess" # Change to a more restrictive policy if needed
}

output "eks_cluster_role_arn" {
  value = aws_iam_role.eks_cluster_role.arn
}
