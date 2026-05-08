output "studio_bucket_name" {
  value       = alicloud_oss_bucket.studio.bucket
  description = "The OSS bucket used by the Studio static website"
}

output "studio_bucket_endpoint" {
  value       = alicloud_oss_bucket.studio.extranet_endpoint
  description = "The public OSS endpoint used as the CNAME target"
}

output "studio_website_url" {
  value       = "https://${var.studio_domain_name}"
  description = "The Studio URL, available after OSS custom domain binding is configured"
}

output "studio_dns_record_id" {
  value       = alicloud_alidns_record.studio_cname.id
  description = "The DNS record ID"
}

output "provider_service_name" {
  value       = module.fc.service_name
  description = "Function Compute service name"
}

output "provider_function_name" {
  value       = module.fc.function_name
  description = "Function Compute function name"
}

output "provider_invoke_url" {
  value       = module.fc.invoke_url
  description = "Provider HTTP trigger invoke URL"
}
