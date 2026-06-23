# Packer

> Golden image builds for Linode VMs and Docker base images.

## In this repo

| Path | Purpose |
|------|---------|
| `packer/linode/rootvalidator.pkr.hcl` | Linode Ubuntu 22.04 custom image |
| `packer/linode/variables.pkr.hcl` | Linode build variables |
| `packer/docker/ubuntu-base/ubuntu-base.json` | Docker Ubuntu base with Ansible |
| `packer/docker/ubuntu-base/install-ansible.sh` | Ansible bootstrap script |
| `Makefile` | `packer-build`, `packer-validate`, `packer-ubuntu` targets |

## Quick start

### Linode golden image

```bash
# Set variables in packer/linode/variables.auto.pkrvars.hcl
make packer-validate
make packer-build
```

### Docker Ubuntu base

```bash
make packer-ubuntu
make docker-push-ubuntu   # pushes wbg0x/wbg0x-ubuntu-base:latest
```

## Configuration

### Linode build

- Base image: `linode/ubuntu22.04`
- Region: `us-west`
- Instance type: `g6-nanode-1`
- Provisioners: [Ansible](ansible.md) playbooks (`limited_user_account.yml`, `common_server_setup.yml`)

### Variables

| Variable | Description |
|----------|-------------|
| `linode_api_token` | Linode API token |

Store secrets in `variables.auto.pkrvars.hcl` (gitignored) or environment.

## Making changes

1. Edit `.pkr.hcl` or `.json` templates.
2. Validate: `make packer-validate`
3. Build: `make packer-build`
4. CI runs `packer fmt -check` in [GitHub Actions CI](../cicd/github-actions-ci.md).

## Integration

```
Packer (Linode) → Ansible provision → custom Linode image
                                              ↓
                                    Terraform (image_id)
```

```
Packer (Docker) → Ansible packages → wbg0x/wbg0x-ubuntu-base image
```

## Official resources

- [Packer documentation](https://developer.hashicorp.com/packer/docs)
- [Linode Packer plugin](https://github.com/linode/packer-plugin-linode)
- [Ansible provisioner](https://developer.hashicorp.com/packer/docs/provisioners/ansible)
