
module "variable_account"{
 source= "./../../../variables"
}
resource "aws_iam_role" "aws-codepipeline-lab-pipeline-role" {
  name = "aws-codepipeline-lab-pipeline-role"
  
  assume_role_policy = jsonencode({
    Version= "2012-10-17",
       Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
            Service: "codepipeline.amazonaws.com"
        }
      }
      ]
})
}

resource "aws_iam_role_policy_attachment" "AWSCodePipeline" {
  policy_arn = "arn:aws:iam::${module.variable_account.account_id}:policy/AWSCodePipeline"  # Replace 'XYZ' with the actual policy name or ARN

  role = aws_iam_role.aws-codepipeline-lab-pipeline-role.name
}

output "codepipeline_arn"{
  value=aws_iam_role.aws-codepipeline-lab-pipeline-role.arn
}




