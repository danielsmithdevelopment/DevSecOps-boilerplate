password: 
	docker run -it --rm alpine mkpasswd -m sha-512 $(password) -s "wbg0x"

packer-build:
	packer build -var-file packer/linode/variables.auto.pkrvars.hcl packer/linode/

packer-validate:
	packer validate -var-file packer/linode/variables.auto.pkrvars.hcl packer/linode/
	
packer-ubuntu:
	DOCKER_REPOSITORY=wbg0x/wbg0x-ubuntu-base IMAGE_VERSION=latest packer build packer/docker/ubuntu-base/ubuntu-base.json

terraform-plan-dev:
	terraform -chdir=terraform/linode/dev/ init
	terraform -chdir=terraform/linode/dev/ plan -out=tfplan-create-dev

terraform-apply-dev:
	terraform -chdir=terraform/linode/dev apply "tfplan-create-dev"