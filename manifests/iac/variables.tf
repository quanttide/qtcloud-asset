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

variable "provider_image" {
  description = "Container image used by the provider Function Compute function"
  type        = string
  default     = "registry.cn-hangzhou.aliyuncs.com/quanttide/qtcloud-asset-provider:latest"
}
