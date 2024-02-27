locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

source "linode" "trn-seed" {
  image             = "linode/ubuntu22.04"
  image_description = "This image was created using Packer."
  image_label       = "packer-ubuntu22-${local.timestamp}"
  instance_label    = "temp-packer-ubuntu22-${local.timestamp}"
  instance_type     = "g6-nanode-1"
  swap_size         = 256
  linode_token      = "${var.linode_api_token}"
  region            = "us-west"
  ssh_username      = "root"
}

build {
  sources = ["source.linode.trn-seed"]

  provisioner "ansible" {
    playbook_file = "ansible/limited_user_account.yml"
    user          = "root"
    use_proxy     = false
  }

  provisioner "ansible" {
    playbook_file = "ansible/common_server_setup.yml"
    use_proxy     = false
  }
}

packer {
  required_plugins {
    linode = {
      version = ">= 1.0.1"
      source  = "github.com/linode/linode"
    }

    ansible = {
      version = "~> 1"
      source = "github.com/hashicorp/ansible"
    }
  }
}
