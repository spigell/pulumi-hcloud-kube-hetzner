SHELL := /bin/bash

include pulumi.Makefile

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
