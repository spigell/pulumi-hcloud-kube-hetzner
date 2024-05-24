// This template is used mostly for local development.
// It is not possible to change configuration with pulumi.Output values when using the project as golang module.
// All dependencies must be resolved!

package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
)

type clusters map[string]map[string]any

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		var clu clusters
		cfg.RequireObject("clusters", &clu)

		outputs := make(pulumi.MapMap, 0)

		for name, config := range clu {
			cluster, err := phkh.New(ctx, name, config, []pulumi.ResourceOption{})
			if err != nil {
				return fmt.Errorf("error while cluster (id:%s) initialization: %w", name, err)
			}

			deployed, err := cluster.Up()
			if err != nil {
				return fmt.Errorf("error while cluster (id:%s) creation: %w", name, err)
			}

			outputs[name] = pulumi.Map{
				phkh.PrivatekeyKey:     deployed.PrivateKey,
				phkh.HetznerServersKey: deployed.Servers,
				phkh.KubeconfigKey:     deployed.Kubeconfig,
			}

		}
		ctx.Export(phkh.PhkhKey, outputs)

		return nil
	})
}
