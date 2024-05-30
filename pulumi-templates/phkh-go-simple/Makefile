SHELL := /bin/bash

export HCLOUD_TOKEN ?= ""
WITH_HCLOUD_TOKEN := microos

ifneq (,$(filter $(MAKECMDGOALS),$(WITH_HCLOUD_TOKEN)))
        ifeq ($(HCLOUD_TOKEN),"")
                export HCLOUD_TOKEN=$(shell bash -c 'read -s -p "Enter your HCLOUD_TOKEN: " hcloud_token; echo $$hcloud_token')
        endif
endif

microos:
	@cd image-builder/microos && \
	packer init template.pkr.hcl && packer build template.pkr.hcl

pulumi-debug-provider:
	PULUMI_DEBUG_PROVIDERS="hcloud-kube-hetzner:$(shell sudo ss -tulnp | grep 'pulumi-resource' | awk '{print $$5}' | cut -f 2 -d ":")" pulumi $(DEBUG_COMMAND) --logtostderr -v 9 2> /tmp/log.txt
