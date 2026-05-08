variable "project_name" {
  description = "Project name"
  type        = string
  default     = "qtcloud-asset"
}

variable "region" {
  description = "Aliyun region"
  type        = string
  default     = "cn-hangzhou"
}

variable "studio_bucket_name" {
  description = "OSS bucket name for the Flutter Studio static website"
  type        = string
  default     = "qtcloud-asset-studio"
}

variable "studio_domain_name" {
  description = "Custom domain for the Studio website"
  type        = string
  default     = "asset.quanttide.com"
}

variable "index_document" {
  description = "Static website index document"
  type        = string
  default     = "index.html"
}

variable "error_document" {
  description = "Static website error document"
  type        = string
  default     = "index.html"
}

variable "provider_registry_namespace" {
  description = "ACR namespace used by the provider image"
  type        = string
  default     = "quanttide"
}

variable "provider_registry_repo" {
  description = "ACR repository used by the provider image"
  type        = string
  default     = "qtcloud-asset-provider"
}
