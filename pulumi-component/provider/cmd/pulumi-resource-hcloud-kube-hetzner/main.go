//go:generate go run ./generate.go

package main

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pulumi-component/pkg/provider"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pulumi-component/pkg/version"
)

func main() {
	provider.Serve(version.Version, pulumiSchema)
}
