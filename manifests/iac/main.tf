terraform {
  required_version = ">= 1.6.0"

  required_providers {
    alicloud = {
      source  = "hashicorp/alicloud"
      version = "~> 1.230.0"
    }
  }

  backend "oss" {
    bucket     = "qtcloud-asset-terraform-state"
    prefix     = "state"
    region     = "cn-hangzhou"
    encrypt    = true
  }
}

provider "alicloud" {
  region = var.region
}

module "vpc" {
  source = "./modules/vpc"

  vpc_name   = "${var.project_name}-vpc"
  cidr_block = var.vpc_cidr
  region     = var.region
}

module "oss" {
  source = "./modules/oss"

  bucket_name = "${var.project_name}-assets"
  region      = var.region
}

module "fc" {
  source = "./modules/fc"

  service_name = "${var.project_name}-service"
  function_name = "${var.project_name}-fn"
  region       = var.region
}

module "api_gateway" {
  source = "./modules/api-gateway"

  group_name = "${var.project_name}-api"
  region     = var.region
}

module "trigger" {
  source = "./modules/trigger"

  function_name = module.fc.function_name
  service_name  = module.fc.service_name
  region        = var.region
}
