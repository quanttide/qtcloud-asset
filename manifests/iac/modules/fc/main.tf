terraform {
  required_providers {
    alicloud = {
      source = "aliyun/alicloud"
    }
  }
}

variable "service_name" {
  type = string
}

variable "function_name" {
  type = string
}

variable "region" {
  type = string
}

variable "code_bucket" {
  type = string
}

variable "code_object" {
  type = string
}

variable "domain_name" {
  type = string
}

variable "enable_domain" {
  type = bool
}

resource "alicloud_fcv3_function" "this" {
  function_name        = var.function_name
  description          = "QtCloud Asset provider"
  runtime              = "custom.debian12"
  handler              = "not-used"
  memory_size          = 1024
  cpu                  = 1
  disk_size            = 512
  timeout              = 300
  instance_concurrency = 10
  internet_access      = true

  code {
    oss_bucket_name = var.code_bucket
    oss_object_name = var.code_object
  }

  custom_runtime_config {
    command = ["./bootstrap"]
    args    = []
    port    = 9000

    health_check_config {
      http_get_url          = "/health"
      initial_delay_seconds = 1
      period_seconds        = 10
      timeout_seconds       = 3
      failure_threshold     = 3
      success_threshold     = 1
    }
  }
}

resource "alicloud_fcv3_trigger" "http_trigger" {
  function_name = alicloud_fcv3_function.this.function_name
  trigger_name  = "http-trigger"
  trigger_type  = "http"
  qualifier     = "LATEST"
  description   = "Public HTTP trigger for QtCloud Asset provider"

  trigger_config = jsonencode({
    authType = "anonymous"
    methods  = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  })
}

resource "alicloud_fcv3_custom_domain" "api" {
  count = var.enable_domain ? 1 : 0

  custom_domain_name = var.domain_name
  protocol           = "HTTP"

  route_config {
    routes {
      function_name = alicloud_fcv3_function.this.function_name
      path          = "/*"
      qualifier     = "LATEST"
    }
  }
}

output "service_name" {
  value = var.service_name
}

output "function_name" {
  value = alicloud_fcv3_function.this.function_name
}

output "invoke_url" {
  value = alicloud_fcv3_trigger.http_trigger.http_trigger[0].url_internet
}

output "custom_domain" {
  value = var.enable_domain ? alicloud_fcv3_custom_domain.api[0].custom_domain_name : null
}
