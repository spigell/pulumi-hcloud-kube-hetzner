PULUMI := pulumi --non-interactive
PULUMI_STACK ?= 
PULUMI_SSH_KEY_FILE ?= /tmp/phkh.key
# Default is pulumi service
PULUMI_BACKEND ?=
PULUMI_CONFIG_SOURCE ?= ../examples/k3s-private-non-ha-simple.yaml
PULUMI_STACK_INIT_FLAGS ?=
HCLOUD_IMAGE ?= 

WITH_PULUMI_STACK_DEFINED := pulumi-stack pulumi-config

ifneq (,$(filter $(MAKECMDGOALS),$(WITH_PULUMI_STACK_DEFINED)))
        ifeq ($(PULUMI_STACK),)
                PULUMI_STACK := $(shell bash -c 'read -p "Enter stack name: " pulumi_stack; echo $$pulumi_stack')
                export $(PULUMI_STACK)
        endif
endif

ci-pulumi-prepare: pulumi-login pulumi-stack pulumi-config

pulumi-login:
	pulumi logout
	pulumi login $(PULUMI_BACKEND)

pulumi-select:
	$(PULUMI) stack select $(PULUMI_STACK)

pulumi-stack:
	$(PULUMI) stack rm --yes --force -s $(PULUMI_STACK) || true
	$(PULUMI) stack init $(PULUMI_STACK) $(PULUMI_STACK_INIT_FLAGS)

pulumi-config: pulumi-stack
	cp $(PULUMI_CONFIG_SOURCE) ./Pulumi.$(PULUMI_STACK).yaml
	sed -i "s/pulumi-hcloud-kube-hetzner/pkhk/g" ./Pulumi.$(PULUMI_STACK).yaml
	@echo "Pulumi.$(PULUMI_STACK).yaml is generated"

pulumi-ssh-check:
	$(PULUMI) stack output --show-secrets -j 'ssh:keypair' | jq .PrivateKey -r > $(PULUMI_SSH_KEY_FILE)
	chmod 600 $(PULUMI_SSH_KEY_FILE)
	@JSON=$$(pulumi stack output --show-secrets -j 'hetzer:servers') && \
	for i in `echo $${JSON} | jq -r 'keys[]'`; do \
		ssh -i $(PULUMI_SSH_KEY_FILE) -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
		-l `echo $${JSON} | jq -r --arg k $$i '.[$$k] | .user'` \
		`echo $${JSON} | jq -r --arg k $$i '.[$$k] | .ip'` \
		'echo "Greetings from `hostname`"' ; \
	done

pulumi-ssh-to-node:
	$(PULUMI) stack output --show-secrets -j 'ssh:keypair' | jq .PrivateKey -r > $(PULUMI_SSH_KEY_FILE)
	chmod 600 $(PULUMI_SSH_KEY_FILE)
	@JSON=$$(pulumi stack output --show-secrets -j 'hetzner:servers' | jq '.["$(TARGET)"]') && \
	ssh -i $(PULUMI_SSH_KEY_FILE) -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
		-l `echo $${JSON} | jq -r .user` \
		`echo $${JSON} | jq -r .ip`

pulumi-wireguard-check:
	$(PULUMI) stack output --show-secrets 'wireguard:connection' > ./wg0.conf
	wg-quick up ./wg0.conf
	@JSON=$$(pulumi stack output --show-secrets -j 'wireguard:info') && \
	for i in `echo $${JSON} | jq -r 'keys[]'`; do \
		ping -c 2 `echo $${JSON} | jq -r --arg k $$i '.[$$k] | .ip'`; \
	done
