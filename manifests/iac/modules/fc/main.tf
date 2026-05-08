variable "service_name" {
  type = string
}

variable "function_name" {
  type = string
}

variable "region" {
  type = string
}

variable "image" {
  type = string
}

resource "alicloud_ram_role" "this" {
  role_name = "${var.service_name}-fc-role"

  assume_role_policy_document = jsonencode({
    Version = "1"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = ["fc.aliyuncs.com"]
        }
      }
    ]
  })
}

resource "alicloud_ram_role_policy_attachment" "this" {
  role_name   = alicloud_ram_role.this.role_name
  policy_name = "AliyunFCFullAccess"
  policy_type = "System"
}

resource "alicloud_fc_service" "this" {
  name = var.service_name
  role = alicloud_ram_role.this.arn
}

resource "alicloud_fc_function" "this" {
  service     = alicloud_fc_service.this.name
  name        = var.function_name
  runtime     = "custom-container"
  handler     = "main.handler"
  memory_size = 512
  timeout     = 60
  ca_port     = 9000

  custom_container_config {
    image = var.image
  }
}

resource "alicloud_fc_trigger" "http_trigger" {
  service  = alicloud_fc_service.this.name
  function = alicloud_fc_function.this.name
  name     = "http-trigger"
  type     = "http"

  config = jsonencode({
    authType = "anonymous"
    methods  = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  })
}

output "service_name" {
  value = alicloud_fc_service.this.name
}

output "function_name" {
  value = alicloud_fc_function.this.name
}

output "invoke_url" {
  value = "https://${alicloud_fc_service.this.name}.${var.region}.fc.aliyuncs.com/2016-08-15/proxy/${alicloud_fc_function.this.name}/"
}
