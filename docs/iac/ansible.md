# Ansible

> Configuration management for Packer builds and server hardening.

## In this repo

| Path | Purpose |
|------|---------|
| `ansible/limited_user_account.yml` | Create `wbg0x` user with sudo |
| `ansible/common_server_setup.yml` | Hostname, timezone, apt upgrade |
| `ansible/apt_install_packages.yml` | Package installation for Docker base |
| `ansible/requirements.yml` | Collections (`ansible.posix`) |
| `.ansible-lint` | Lint configuration |

## Quick start

Playbooks run automatically via [Packer](packer.md) provisioners. To run manually:

```bash
ansible-galaxy install -r ansible/requirements.yml
ansible-playbook -i <inventory> ansible/common_server_setup.yml
```

## Playbooks

### `limited_user_account.yml`

Creates the `wbg0x` service user with SSH key and sudo access. Run as root during Packer build.

### `common_server_setup.yml`

- UTC timezone
- Hostname / FQDN (`wbg0x-private`, `private.wbg0x.com`)
- `/etc/hosts` entries
- `apt upgrade`

### `apt_install_packages.yml`

Installs packages for the Docker Ubuntu base image build.

## Making changes

1. Edit playbooks under `ansible/`.
2. Lint: `ansible-lint` (runs in CI).
3. Test via `make packer-build` or against a staging VM.
4. Pin collection versions in `requirements.yml`.

## Integration

- Called by [Packer](packer.md) Linode and Docker builds
- Linted in [GitHub Actions CI](../cicd/github-actions-ci.md)

## Official resources

- [Ansible documentation](https://docs.ansible.com/)
- [ansible-lint](https://ansible.readthedocs.io/projects/lint/)
- [ansible.posix collection](https://docs.ansible.com/ansible/latest/collections/ansible/posix/)
