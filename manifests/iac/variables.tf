variable "project_name" {
  description = "项目名称"
  type        = string
  default     = "qtcloud-asset"
}

variable "region" {
  description = "阿里云区域"
  type        = string
  default     = "cn-hangzhou"
}

variable "vpc_cidr" {
  description = "VPC CIDR 块"
  type        = string
  default     = "172.16.0.0/16"
}

variable "environment" {
  description = "环境名称（dev/staging/prod）"
  type        = string
  default     = "dev"
}
