package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	phkhlib "github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
	phkh "github.com/spigell/pulumi-hcloud-kube-hetzner/pulumi-component/sdk/go/hcloud-kube-hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pulumi-component/sdk/go/hcloud-kube-hetzner/cluster"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
	        outputs := make(pulumi.MapMap, 0)

                clusterName := "simple"

		cluster, err := phkh.NewCluster(ctx, clusterName, &phkh.ClusterArgs{
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
			return fmt.Errorf("error while the `%s` cluster initialization: %w", %clusterName, err)
		}


		outputs[clusterName] = pulumi.Map{
			phkhlib.PrivatekeyKey:     cluster.Privatekey,
			phkhlib.HetznerServersKey: cluster.Servers,
			phkhlib.KubeconfigKey:     cluster.Kubeconfig,
		}
		ctx.Export(phkhlib.PhkhKey, outputs)

		return nil
	})
}
