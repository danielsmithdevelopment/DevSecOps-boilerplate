variable "linode_api_token" {
  type    = string
  default = ""
}

variable "image_id" {
  type    = string
  default = ""
}

variable "root_pass" {
  type      = string
  default   = ""
  sensitive = true
}

variable "authorized_keys" {
  type    = string
  default = ""
}
