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
  default     = "asset.cloud.quanttide.com"
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

variable "provider_code_object" {
  description = "OSS object key used by the provider Function Compute code package"
  type        = string
  default     = "provider/qtcloud-asset-provider-go-linux-amd64.zip"
}

variable "provider_code_bucket_name" {
  description = "Existing OSS bucket that stores Provider release packages"
  type        = string
  default     = "qtcloud-asset"
}

variable "provider_domain_name" {
  description = "Custom domain for the provider API"
  type        = string
  default     = "api.quanttide.com"
}

variable "enable_provider_custom_domain" {
  description = "Whether to create an optional FC custom domain; keep false when using the platform API gateway"
  type        = bool
  default     = false
}
