
ifeq ($(M),)
        export HCLOUD_TOKEN=$(shell bash -c 'read -s -p "Enter your HCLOUD_TOKEN: " hcloud_token; echo $$hcloud_token')
endif


build-microos:
	cd image-builder/microos && \
	packer init template.pkr.hcl && packer build template.pkr.hcl
