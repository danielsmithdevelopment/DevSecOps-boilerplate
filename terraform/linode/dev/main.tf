terraform {
  required_version = ">= 1.9.0"

  required_providers {
    linode = {
      source  = "linode/linode"
      version = "3.14.0"
    }
  }
}

provider "linode" {
  token = var.linode_api_token
}

module "ubuntu22-instance" {
  source          = "../modules/linode-instance"
  image_id        = var.image_id
  root_pass       = var.root_pass
  authorized_keys = var.authorized_keys
}
