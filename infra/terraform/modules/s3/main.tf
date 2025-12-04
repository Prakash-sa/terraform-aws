module "variable_account"{
 source= "./../../variables"
}
resource "aws_s3_bucket" "my_bucket"{
    bucket="bucket-${module.variable_account.account_id}"
}
output "s3_name" {
  value = aws_s3_bucket.my_bucket.bucket
}