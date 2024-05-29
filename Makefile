SHELL := bash
DEV_TEMPLATE ?= dev/go-module-development
DEV_PROJECT := phkh-dev
TAG ?= $(shell git describe --tags --abbrev=0)

GH_EXAMPLE ?= k3s-private-non-ha-simple

# Used in CI
test-go-project: clean
	@mkdir -p test-project
	@cd test-project && \
	pulumi new ../pulumi-templates/$(DEV_TEMPLATE) -g -n $(DEV_PROJECT) --yes && \
	go mod edit -replace=github.com/spigell/pulumi-hcloud-kube-hetzner=../
	@go work use ./test-project
	@echo "Now you can create stack for test project in test-project directory"
	@echo 'Please use command `pulumi-config PULUMI_CONFIG_SOURCE=/path/to/file` to set config source for the stack'
	@echo -e "If the list of files: \033[0;31m [main.go, go.mod, go.sum] \033[0m changed, please add the changes in pulumi-template directory"

test-ts-project: clean
	@mkdir -p test-project
	@cd test-project && \
	pulumi new ../pulumi-templates/$(DEV_TEMPLATE) -g -n $(DEV_PROJECT) --yes && \
	yarn link --cwd ../pulumi-component/sdk/nodejs/bin && \
	sed -i '/\@spigell\/hcloud-kube-hetzner/d' package.json && \
	yarn link "@spigell/hcloud-kube-hetzner" && \
	yarn install
	@echo "Now you can create stack for test project in test-project directory"
	@echo -e "If the list of files: \033[0;31m [index.ts, package.json] \033[0m changed, please add the changes in pulumi-template directory"
	
clean:
	go work edit -dropuse ./test-project || true
	yarn unlink --cwd pulumi-component/sdk/nodejs/bin || true
	rm -rf test-project

github-run:
	gh workflow run --ref $$(git rev-parse --abbrev-ref HEAD) -f example=$(GH_EXAMPLE) main-test-examples.yaml
	sleep 10
	watch gh run view $$(gh run list --workflow=main-test-examples.yaml -b $$(git rev-parse --abbrev-ref HEAD) -L 1 --json databaseId | jq .[0].databaseId -r) -v

up-template-versions:: up-go-template-versions up-typescript-template-versions

up-go-template-versions: export TEMPLATES = dev/go-module-development phkh-go-multiple-clusters phkh-go-simple phkh-go-cluster-files
up-go-template-versions:
	@for TMP in $(TEMPLATES); do \
		$(MAKE) test-go-project DEV_TEMPLATE=$${TMP} && \
		cd test-project && \
		go mod edit -dropreplace=github.com/spigell/pulumi-hcloud-kube-hetzner && \
		go get -u && go get github.com/spigell/pulumi-hcloud-kube-hetzner@$(TAG) && go mod tidy && \
		cp go.mod ../pulumi-templates/$${TMP}/go.mod && \
		sed -i "1s/.*/module \\\$${PROJECT}/" ../pulumi-templates/$${TMP}/go.mod && \
		cp go.sum ../pulumi-templates/$${TMP}/go.sum && \
		cd .. && \
		$(MAKE) clean ; \
	done

up-typescript-template-versions: export TEMPLATES = phkh-typescript-cluster-files phkh-typescript-simple phkh-typescript-pulumi-config
up-typescript-template-versions:
	@for TMP in $(TEMPLATES); do \
		$(MAKE) test-ts-project DEV_TEMPLATE=$${TMP} && \
		cd test-project && \
		yarn add @spigell/hcloud-kube-hetzner@$(TAG) && \
		yarn add yaml && \
		cp package.json ../pulumi-templates/$${TMP}/package.json && \
		sed -i "2s/.*/    \"name\": \\\"\$${PROJECT}\",/" ../pulumi-templates/$${TMP}/package.json && \
		cp yarn.lock ../pulumi-templates/$${TMP}/yarn.lock && \
		cd .. && \
		$(MAKE) clean ; \
	done

sync-cluster-files: export SOURCE = phkh-go-cluster-files
sync-cluster-files: export TARGETS = phkh-typescript-cluster-files phkh-typescript-pulumi-config
sync-cluster-files:	
	@for a in $(TARGETS); do \
		cd pulumi-templates && \
		cp -vr \
			$(SOURCE)/cluster-examples \
			./$${a}/ ; \
		cd - ; \
	done

# This stage syncs templates with the GO temlates
sync-templates: sync-cluster-files
sync-templates: export SOURCE = phkh-go-simple
sync-templates: export TARGETS = phkh-go-cluster-files phkh-go-multiple-clusters phkh-typescript-cluster-files phkh-typescript-simple dev/go-module-development phkh-typescript-pulumi-config
sync-templates:
	@for a in $(TARGETS); do \
		cd pulumi-templates && \
		cp -vr \
			$(SOURCE)/versions \
			$(SOURCE)/Makefile \
			$(SOURCE)/.gitignore \
			$(SOURCE)/image-builder \
			./$${a}/ ; \
		cd - ; \
	done

unit-tests:
	cd pulumi-component && make generate_schema
	set -o pipefail ; go test $$(go list ./... | grep -v integration | grep -v crds/generated) | grep -v 'no test files'
