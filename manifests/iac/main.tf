terraform {
  required_version = ">= 1.0"

  required_providers {
    alicloud = {
      source  = "aliyun/alicloud"
      version = "~> 1.212"
    }
  }
}
provider "alicloud" {
  region = var.region
}

resource "alicloud_oss_bucket" "studio" {
  bucket = var.studio_bucket_name

  versioning {
    status = "Enabled"
  }

  website {
    index_document = var.index_document
    error_document = var.error_document
  }

  lifecycle {
    prevent_destroy = true
  }
}
resource "alicloud_oss_bucket_acl" "studio_public_read" {
  bucket = alicloud_oss_bucket.studio.bucket
  acl    = "public-read"
}

module "fc" {
  source = "./modules/fc"

  service_name  = "${var.project_name}-service"
  function_name = "provider-package"
  region        = var.region
  code_bucket   = var.provider_code_bucket_name
  code_object   = var.provider_code_object
  domain_name   = var.provider_domain_name
  enable_domain = var.enable_provider_custom_domain
}
