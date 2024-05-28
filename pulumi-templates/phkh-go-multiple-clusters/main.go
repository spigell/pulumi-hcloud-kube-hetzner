package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	phkhlib "github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	phkh "github.com/spigell/pulumi-hcloud-kube-hetzner/pulumi-component/sdk/go/hcloud-kube-hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pulumi-component/sdk/go/hcloud-kube-hetzner/cluster"
)

type clusters map[string]map[string]any

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		var clu clusters
		cfg.RequireObject("clusters", &clu)

		small, err := phkh.NewCluster(ctx, "small", &phkh.ClusterArgs{
			Config: cluster.ConfigConfigArgs{
				Network: &cluster.ConfigNetworkConfigArgs{
					Hetzner: &cluster.NetworkConfigArgs{
						Enabled: pulumi.Bool(true),
					},
				},
				Nodepools: &cluster.ConfigNodepoolsConfigArgs{
					Servers: &cluster.ConfigNodepoolConfigArray{
						&cluster.ConfigNodepoolConfigArgs{
							PoolID: pulumi.String("servers"),
							Nodes: cluster.ConfigNodeConfigArray{
								&cluster.ConfigNodeConfigArgs{
									NodeID: pulumi.String("server-01"),
								},
							},
						},
					},
				},
			},
		},
		)

		if err != nil {
			return fmt.Errorf("error while the small cluster initialization: %w", err)
		}

		outputs := make(pulumi.MapMap, 0)

		for name, cfg := range clu {
			cluster, err := phkh.NewCluster(ctx, name, &phkh.ClusterArgs{
				Config:               pulumi.ToMap(cfg),
				UseKebabConfigFormat: pulumi.Bool(true),
			})
			if err != nil {
				return fmt.Errorf("error while cluster (id:%s) initialization: %w", name, err)
			}

			outputs[name] = pulumi.Map{
				phkhlib.PrivatekeyKey:     cluster.Privatekey,
				phkhlib.HetznerServersKey: cluster.Servers,
				phkhlib.KubeconfigKey:     cluster.Kubeconfig,
			}

		}
		outputs["small"] = pulumi.Map{
			phkhlib.PrivatekeyKey:     small.Privatekey,
			phkhlib.HetznerServersKey: small.Servers,
			phkhlib.KubeconfigKey:     small.Kubeconfig,
		}
		ctx.Export(phkhlib.PhkhKey, outputs)

		return nil
	})
}
