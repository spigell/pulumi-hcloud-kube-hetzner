package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
)



func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cluster, err := phkh.New(ctx)

		if err != nil {
			return err
		}

		return cluster.Up()
	})
}
