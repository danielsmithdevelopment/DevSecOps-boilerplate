# Terraform

> Linode infrastructure as code — VM provisioning with cloud firewall.

## In this repo

| Path | Purpose |
|------|---------|
| `terraform/linode/dev/main.tf` | Dev environment entry |
| `terraform/linode/dev/variables.tf` | Input variables |
| `terraform/linode/modules/linode-instance/` | Ubuntu 22 hardened instance module |
| `terraform/linode/modules/cloud-firewall/` | Inbound SSH firewall module |
| `.tflint.hcl` | TFLint configuration |
| `Makefile` | `terraform-plan-dev`, `terraform-apply-dev` |

## Quick start

```bash
# Set LINODE_TOKEN, image_id (from Packer), root_pass, authorized_keys
make terraform-plan-dev
make terraform-apply-dev
```

## Configuration

### Variables (`variables.tf`)

| Variable | Description |
|----------|-------------|
| `linode_api_token` | Linode API token |
| `image_id` | Custom image ID from [Packer](packer.md) build |
| `root_pass` | Root password |
| `authorized_keys` | SSH public keys |

### Modules

- **linode-instance** — deploys hardened Ubuntu 22 from golden image
- **cloud-firewall** — restricts inbound to SSH

## Making changes

1. Edit modules or `dev/main.tf`.
2. Format: `terraform fmt -recursive terraform/`
3. Lint: `tflint --config .tflint.hcl`
4. Plan before apply: `make terraform-plan-dev`
5. CI runs fmt + tflint on every PR.

## Integration

```
Packer golden image → image_id → Terraform linode-instance module
```

## Official resources

- [Terraform documentation](https://developer.hashicorp.com/terraform/docs)
- [Linode provider](https://registry.terraform.io/providers/linode/linode/latest/docs)
- [TFLint](https://github.com/terraform-linters/tflint)
