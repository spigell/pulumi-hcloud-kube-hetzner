package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	phkh "github.com/spigell/pulumi-hcloud-kube-hetzner/pulumi-component/sdk/go/hcloud-kube-hetzner"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		_, err := phkh.NewCluster(ctx, "test", &phkh.ClusterArgs{})

		return err
	})
}
