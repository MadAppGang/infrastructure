output "subnet_ids" {
  value = data.aws_subnets.all.ids
}
output "account_id" {
  value = data.aws_caller_identity.current.account_id
}

output "region" {
  value = data.aws_region.current.name
}

output "alb_dns_name" {
  value = module.workloads.alb_dns_name
}

output "backend_ecr_repo_url" {
  value = module.workloads.backend_ecr_repo_url
}

