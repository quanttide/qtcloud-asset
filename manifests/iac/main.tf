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
    prevent_destroy = false
  }
}

resource "alicloud_alidns_record" "studio_cname" {
  domain_name = "quanttide.com"
  type        = "CNAME"
  rr          = "asset"
  value       = "${alicloud_oss_bucket.studio.bucket}.${alicloud_oss_bucket.studio.extranet_endpoint}"
  ttl         = 600
  status      = "ENABLE"

  depends_on = [alicloud_oss_bucket.studio]
}

resource "alicloud_cr_namespace" "provider" {
  name               = var.provider_registry_namespace
  auto_create        = false
  default_visibility = "PRIVATE"
}

resource "alicloud_cr_repo" "provider" {
  namespace = alicloud_cr_namespace.provider.name
  name      = var.provider_registry_repo
  summary   = "QtCloud Asset provider image"
  repo_type = "PRIVATE"
  detail    = "Container image repository for the QtCloud Asset provider Function Compute service."
}

module "fc" {
  source = "./modules/fc"

  service_name  = "${var.project_name}-service"
  function_name = "provider"
  region        = var.region
  image         = "registry.cn-hangzhou.aliyuncs.com/${alicloud_cr_repo.provider.namespace}/${alicloud_cr_repo.provider.name}:latest"
}
