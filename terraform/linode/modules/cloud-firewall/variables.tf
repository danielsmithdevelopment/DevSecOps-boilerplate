variable "linodes" {
  description = "List of Linode ids to which the rule sets will be applied"
  type        = list(string)
  default     = []
}

variable "firewall_label" {
  description = "This firewall's human-readable firewall_label"
  type        = string
  default     = "my-firewall"
}

variable "tags" {
  description = "List of tags to apply to this Firewall"
  type        = list(string)
  default     = []
}
