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
