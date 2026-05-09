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

module "fc" {
  source = "./modules/fc"

  service_name  = "${var.project_name}-service"
  function_name = "provider-package"
  region        = var.region
  code_bucket   = alicloud_oss_bucket.studio.bucket
  code_object   = var.provider_code_object
}
