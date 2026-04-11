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

module "ecs" {
  source = "./modules/ecs"

  instance_name = "${var.project_name}-server"
  vpc_id        = module.vpc.vpc_id
  vswitch_id    = module.vpc.vswitch_id
  region        = var.region
}
