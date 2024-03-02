terraform {
  required_providers {
    linode = {
      source = "linode/linode"
      version = "2.5.2"
    }
  }
}

provider "linode" {
  token = var.linode_api_token
}

resource "linode_instance" "wbg0x-private" {
        image = var.image_id
        label = "wbg0x-private"
        group = "wbg0x"
        region = "us-west"
        type = "g7-premium-8"
        authorized_keys = [ var.authorized_keys ]
        root_pass = var.root_pass
}