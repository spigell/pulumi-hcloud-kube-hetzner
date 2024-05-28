PULUMI := pulumi --non-interactive
PULUMI_STACK ?=
PULUMI_SSH_KEY_FILE ?= /tmp/phkh.key
# Default is pulumi service
PULUMI_BACKEND ?=
PULUMI_EXAMPLE_NAME ?= k3s-private-non-ha-simple
PULUMI_CONFIG_SOURCE ?= ../examples/$(PULUMI_EXAMPLE_NAME).yaml
PULUMI_STACK_INIT_FLAGS ?=
HCLOUD_IMAGE ?= 

ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))

WITH_PULUMI_STACK_DEFINED := pulumi-create-stack pulumi-generate-config

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

pulumi-generate-config: export STACK = $(firstword $(ARGS))
pulumi-generate-config: pulumi-create-stack
	echo 'config-examples/${STACK}.yaml' > /Pulumi.$(PULUMI_STACK).yaml.bak
	rm ./Pulumi.$(PULUMI_STACK).yaml.bak
	@echo "Pulumi.$(PULUMI_STACK).yaml is generated"

pulumi-debug-provider:
	PULUMI_DEBUG_PROVIDERS="hcloud-kube-hetzner:$(shell sudo ss -tulnp | grep 'pulumi-resource' | awk '{print $$5}' | cut -f 2 -d ":")" pulumi $(DEBUG_COMMAND) --logtostderr -v 9 2> /tmp/log.txt

pulumi-ssh-check:
	$(PULUMI) stack output --show-secrets -j phkh | jq '.privatekey' -r > $(PULUMI_SSH_KEY_FILE)
	chmod 600 $(PULUMI_SSH_KEY_FILE)
	@JSON=$$(pulumi stack output --show-secrets -j phkh | jq '.servers') && \
	for i in `echo $${JSON} | jq -r 'keys[]'`; do \
		ssh -i $(PULUMI_SSH_KEY_FILE) -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
		-l `echo $${JSON} | jq -r --arg k $$i '.[$$k] | .user'` \
		`echo $${JSON} | jq -r --arg k $$i '.[$$k] | .ip'` \
		'echo "Greetings from `hostname`"' ; \
	done

pulumi-ssh-to-node:
	$(PULUMI) stack output --show-secrets -j phkh | jq '.privatekey' -r > $(PULUMI_SSH_KEY_FILE)
	chmod 600 $(PULUMI_SSH_KEY_FILE)
	JSON=$$(pulumi stack output --show-secrets -j phkh | jq '.servers' | jq '.[] | select(.name == "$(TARGET)")') && \
	ssh -i $(PULUMI_SSH_KEY_FILE) -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
		-l `echo $${JSON} | jq -r .user` \
		`echo $${JSON} | jq -r .ip`
