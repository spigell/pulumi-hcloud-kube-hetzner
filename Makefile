SHELL := bash
TEMPLATE ?= go/library

GH_EXAMPLE ?= k3s-private-non-ha-simple

# Used in CI
test-go-project: clean
	@mkdir -p test-project
	@cd test-project && \
	pulumi new ../pulumi-templates/$(TEMPLATE) -g -n pkhk --yes && \
	go mod edit -replace=github.com/spigell/pulumi-hcloud-kube-hetzner=../
	@go work use ./test-project
	@echo "Now you can create stack for test project in test-project directory"
	@echo 'Please use command `pulumi-config PULUMI_CONFIG_SOURCE=/path/to/file` to set config source for the stack'
	@echo -e "If the list of files: \033[0;31m [main.go, go.mod, go.sum] \033[0m changed, please add the changes in pulumi-template directory"

test-ts-project: clean
	@mkdir -p test-project
	@cd test-project && \
	pulumi new ../pulumi-templates/typescript -g -n pkhk --yes && \
	yarn link --cwd ../pulumi-component/sdk/nodejs && \
	sed -i '/\@spigell\/hcloud-kube-hetzner/d' package.json && \
	yarn link "@spigell/hcloud-kube-hetzner" && \
	yarn install
	@echo "Now you can create stack for test project in test-project directory"
	@echo 'Please use command `pulumi-config PULUMI_CONFIG_SOURCE=/path/to/file` to set config source for the stack'
	@echo -e "If the list of files: \033[0;31m [main.go, go.mod, go.sum] \033[0m changed, please add the changes in pulumi-template directory"
	
clean:
	go work edit -dropuse ./test-project || true
	yarn unlink --cwd pulumi-component/sdk/nodejs || true
	rm -rf test-project

github-run:
	gh workflow run --ref $$(git rev-parse --abbrev-ref HEAD) -f example=$(GH_EXAMPLE) main-test-examples.yaml
	sleep 10
	watch gh run view $$(gh run list --workflow=main-test-examples.yaml -b $$(git rev-parse --abbrev-ref HEAD) -L 1 --json databaseId | jq .[0].databaseId -r) -v


up-go-lib-template-versions: TEMPLATE = go/library
up-go-lib-template-versions: clean test-go-project
	cd test-project && go mod edit -dropreplace=github.com/spigell/pulumi-hcloud-kube-hetzner
	cd test-project && go get -u && go get github.com/spigell/pulumi-hcloud-kube-hetzner@main && go mod tidy
	cp ./test-project/go.mod ./pulumi-templates/$(TEMPLATE)/go.mod
	sed -i "1s/.*/module \\\$${PROJECT}/" ./pulumi-templates/$(TEMPLATE)/go.mod
	cp ./test-project/go.sum ./pulumi-templates/$(TEMPLATE)/go.sum

# This stage syncs templates with the GO library template
sync-templates:
	for a in go/component typescript; do \
		cd pulumi-templates && \
		cp -vr \
			go/library/README.md \
			go/library/pulumi.Makefile \
			go/library/versions \
			go/library/Makefile \
			go/library/scripts \
			go/library/image-builder \
			./$${a}/ ; \
		cd - ; \
	done

unit-tests:
	set -o pipefail ; go test $$(go list ./... | grep -v integration | grep -v crds/generated) | grep -v 'no test files'
