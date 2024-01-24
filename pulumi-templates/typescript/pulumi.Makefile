PULUMI := pulumi --non-interactive
PULUMI_STACK ?= 
PULUMI_SSH_KEY_FILE ?= /tmp/phkh.key
# Default is pulumi service
PULUMI_BACKEND ?=
PULUMI_EXAMPLE_NAME ?= k3s-private-non-ha-simple
PULUMI_CONFIG_SOURCE ?= ../examples/$(PULUMI_EXAMPLE_NAME).yaml
PULUMI_STACK_INIT_FLAGS ?=
HCLOUD_IMAGE ?= 

WITH_PULUMI_STACK_DEFINED := pulumi-create-stack pulumi-generate-config pulumi-init-stack

ifneq (,$(filter $(MAKECMDGOALS),$(WITH_PULUMI_STACK_DEFINED)))
        ifeq ($(PULUMI_STACK),)
                PULUMI_STACK := $(shell bash -c 'read -p "Enter stack name: " pulumi_stack; echo $$pulumi_stack')
                export $(PULUMI_STACK)
        endif
endif

pulumi-ci-prepare: pulumi-login pulumi-create-stack pulumi-generate-config

pulumi-login:
	pulumi logout
	pulumi login $(PULUMI_BACKEND)

pulumi-select:
	$(PULUMI) stack select $(PULUMI_STACK)

pulumi-create-stack:
	$(PULUMI) stack rm --yes --force -s $(PULUMI_STACK) || true
	$(PULUMI) stack init $(PULUMI_STACK) $(PULUMI_STACK_INIT_FLAGS)

pulumi-generate-config: pulumi-create-stack
	cat $(PULUMI_CONFIG_SOURCE) >> ./Pulumi.$(PULUMI_STACK).yaml
	sed -i "s/pulumi-hcloud-kube-hetzner/pkhk/g" ./Pulumi.$(PULUMI_STACK).yaml
	@echo "Pulumi.$(PULUMI_STACK).yaml is generated"

pulumi-init-stack: export PROJECT_NAME := $(shell grep name Pulumi.yaml | head -1 | cut -f 2 -d ':' | tr -d ' ')
pulumi-init-stack: pulumi-create-stack
	curl -s https://raw.githubusercontent.com/spigell/pulumi-hcloud-kube-hetzner/feature/pulumi-component/examples/${PULUMI_EXAMPLE_NAME}.yaml >> Pulumi.$(PULUMI_STACK).yaml
	sed -i "s/pulumi-hcloud-kube-hetzner/$(PROJECT_NAME)/g" ./Pulumi.$(PULUMI_STACK).yaml
	@echo "Stack config generated from $(PULUMI_EXAMPLE_NAME) example"

pulumi-ssh-check:
	$(PULUMI) stack output --show-secrets -j 'privatekey' | jq . -r > $(PULUMI_SSH_KEY_FILE)
	chmod 600 $(PULUMI_SSH_KEY_FILE)
	@JSON=$$(pulumi stack output --show-secrets -j 'servers') && \
	for i in `echo $${JSON} | jq -r 'keys[]'`; do \
		ssh -i $(PULUMI_SSH_KEY_FILE) -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
		-l `echo $${JSON} | jq -r --arg k $$i '.[$$k] | .user'` \
		`echo $${JSON} | jq -r --arg k $$i '.[$$k] | .ip'` \
		'echo "Greetings from `hostname`"' ; \
	done

pulumi-ssh-to-node:
	$(PULUMI) stack output --show-secrets -j 'privatekey' | jq . -r > $(PULUMI_SSH_KEY_FILE)
	chmod 600 $(PULUMI_SSH_KEY_FILE)
	JSON=$$(pulumi stack output --show-secrets -j 'servers' | jq '.[] | select(.name == "$(TARGET)")') && \
	ssh -i $(PULUMI_SSH_KEY_FILE) -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
		-l `echo $${JSON} | jq -r .user` \
		`echo $${JSON} | jq -r .ip`
