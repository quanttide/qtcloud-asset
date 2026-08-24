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

output "provider_custom_domain_url" {
  value       = module.fc.custom_domain == null ? null : "http://${module.fc.custom_domain}"
  description = "Provider custom domain URL"
}
