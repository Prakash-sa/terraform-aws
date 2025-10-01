resource "aws_ecr_repository" "ecr_repo" {
  name                 = "ecr_repo"

}
output "ecr_name" {
  value = aws_ecr_repository.ecr_repo.name
}