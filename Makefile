SHELL := bash

GH_EXAMPLE ?= default

test-project: clean
	@mkdir -p test-project
	@cd test-project && \
	pulumi new ../pulumi-template -g -n pkhk --yes && \
	go mod edit -replace=github.com/spigell/pulumi-hcloud-kube-hetzner=../
	@go work use ./test-project
	@echo "Now you can create stack for test project in test-project directory"
	@echo 'Please use command `pulumi-config PULUMI_CONFIG_SOURCE=/path/to/file` to set config source for the stack'
	@echo -e "If the list of files: \033[0;31m [main.go, go.mod, go.sum] \033[0m changed, please add the changes in pulumi-template directory"
	
clean:
	go work edit -dropuse ./test-project
	rm -rf test-project

github-run:
	gh workflow run --ref $$(git rev-parse --abbrev-ref HEAD) -f example=$(GH_EXAMPLE) main-test-examples.yaml
	sleep 10
	watch gh run view $$(gh run list --workflow=main-test-examples.yaml -b $$(git rev-parse --abbrev-ref HEAD) -L 1 --json databaseId | jq .[0].databaseId -r) -v
