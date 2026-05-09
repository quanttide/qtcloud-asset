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

resource "alicloud_fcv3_function" "this" {
  function_name        = var.function_name
  description          = "QtCloud Asset provider"
  runtime              = "custom.debian12"
  handler              = "not-used"
  memory_size          = 512
  cpu                  = 0.5
  disk_size            = 512
  timeout              = 60
  instance_concurrency = 10
  internet_access      = true

  code {
    oss_bucket_name = var.code_bucket
    oss_object_name = var.code_object
  }

  custom_runtime_config {
    command = ["uvicorn"]
    args    = ["app.main:app", "--host", "0.0.0.0", "--port", "9000"]
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

output "service_name" {
  value = var.service_name
}

output "function_name" {
  value = alicloud_fcv3_function.this.function_name
}

output "invoke_url" {
  value = alicloud_fcv3_trigger.http_trigger.http_trigger[0].url_internet
}
