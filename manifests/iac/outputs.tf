output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "oss_bucket" {
  description = "OSS 存储桶名称"
  value       = module.oss.bucket_name
}

output "fc_function_url" {
  description = "函数计算 HTTP 触发 URL"
  value       = module.fc.http_trigger_url
}

output "api_gateway_url" {
  description = "API 网关地址"
  value       = module.api_gateway.invoke_url
}
