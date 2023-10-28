PULUMI := pulumi --non-interactive
PULUMI_STACK ?= basic
PULUMI_SSH_KEY_FILE ?= /tmp/phkh.key
# Default is pulumi service
PULUMI_BACKEND ?=
PULUMI_CONFIG_SOURCE ?= examples/$(PULUMI_STACK).yaml
PULUMI_STACK_INIT_FLAGS ?=
HCLOUD_IMAGE ?= 

ci-pulumi-prepare: pulumi-login pulumi-stack pulumi-config

pulumi-login:
	pulumi logout
	pulumi login $(PULUMI_BACKEND)

pulumi-select:
	$(PULUMI) stack select $(PULUMI_STACK)

pulumi-stack:
	$(PULUMI) stack rm --yes --force $(PULUMI_STACK) || true
	$(PULUMI) stack init $(PULUMI_STACK) $(PULUMI_STACK_INIT_FLAGS)

pulumi-config: pulumi-stack
	@if [[ -z $${HCLOUD_IMAGE} ]]; then \
		read -s -p "Enter your HCLOUD_IMAGE: " hcloud_image; \
		export HCLOUD_IMAGE=$$(echo $${hcloud_image}); \
	fi && \
	go run ./scripts/pulumi-config-generator $(PULUMI_CONFIG_SOURCE) >> ./Pulumi.$(PULUMI_STACK).yaml
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

pulumi-wireguard-check:
	$(PULUMI) stack output --show-secrets 'wireguard:connection' > ./wg0.conf
	wg-quick up ./wg0.conf
	@JSON=$$(pulumi stack output --show-secrets -j 'wireguard:info') && \
	for i in `echo $${JSON} | jq -r 'keys[]'`; do \
		ping -c 2 `echo $${JSON} | jq -r --arg k $$i '.[$$k] | .ip'`; \
	done
