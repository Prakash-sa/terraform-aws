data "aws_caller_identity" "current" {}
locals {
    account_id = data.aws_caller_identity.current.account_id
}
output "account_id" {
  value = local.account_id
}

