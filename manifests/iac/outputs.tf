output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "oss_bucket" {
  description = "OSS 存储桶名称"
  value       = module.oss.bucket_name
}

output "ecs_public_ip" {
  description = "ECS 公网 IP"
  value       = module.ecs.public_ip
}
