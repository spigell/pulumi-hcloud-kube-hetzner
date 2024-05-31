package provider

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
)

type Cluster struct {
	pulumi.ResourceState
	ClusterArgs

	HetznerServers pulumi.MapArrayOutput `pulumi:"servers"`
	Kubeconfig     pulumi.StringOutput   `pulumi:"kubeconfig"`
	Privatekey     pulumi.StringOutput   `pulumi:"privatekey"`
}

func (c *Cluster) Type() string {
	return ProviderName + ":index:Cluster"
}

type ClusterArgs struct {
	Config pulumi.MapOutput `pulumi:"config"`
}

func construct(ctx *pulumi.Context, c *Cluster, name string,
	args *ClusterArgs, inputs provider.ConstructInputs, opts ...pulumi.ResourceOption,
) (*provider.ConstructResult, error) {
	// Blit the inputs onto the arguments struct.
	if err := inputs.CopyTo(args); err != nil {
		return nil, errors.Wrap(err, "setting args")
	}

	// Register our component resource.
	if err := ctx.RegisterComponentResource(c.Type(), name, c, opts...); err != nil {
		return nil, err
	}

	finalizer, err := local.NewCommand(ctx, fmt.Sprintf("%s:finalizer", name), &local.CommandArgs{
		Create: pulumi.All(args.Config).ApplyT(
			func(args []any) (v pulumi.StringOutput, err error) {
				cfg := args[0].(map[string]any)

				cluster, err := phkh.NewCluster(ctx, name, cfg, opts)
				if err != nil {
					return v, err
				}

				deployed, err := cluster.Up()
				if err != nil {
					return v, err
				}

				// Create json map manually since json.Marshal can't process output values.
				outputs := pulumi.Sprintf(`{
					"%s": %s,
					"%s": %s,
					"%s": %s
				}`,
					phkh.KubeconfigKey,
					pulumi.JSONMarshal(deployed.Kubeconfig),
					phkh.HetznerServersKey,
					pulumi.JSONMarshal(deployed.Servers),
					phkh.PrivatekeyKey,
					pulumi.JSONMarshal(deployed.Privatekey),
				)

				return pulumi.Sprintf("echo '%s' ", outputs.ApplyT(
					func(v string) string {
						return base64.StdEncoding.EncodeToString([]byte(v))
					})), nil
			}).(pulumi.StringOutput),
		Logging: local.LoggingStderr,
	}, pulumi.AdditionalSecretOutputs([]string{"stdout"}),
	)
	if err != nil {
		return nil, err
	}

	c.HetznerServers = getPulumiKey(finalizer.Stdout, phkh.HetznerServersKey).AsMapArrayOutput()
	c.Kubeconfig = pulumi.ToSecret(getPulumiKey(finalizer.Stdout, phkh.KubeconfigKey).AsStringOutput()).(pulumi.StringOutput)
	c.Privatekey = pulumi.ToSecret(getPulumiKey(finalizer.Stdout, phkh.PrivatekeyKey).AsStringOutput()).(pulumi.StringOutput)

	if err := ctx.RegisterResourceOutputs(c, pulumi.Map{
		phkh.HetznerServersKey: c.HetznerServers,
		phkh.KubeconfigKey:     c.Kubeconfig,
		phkh.PrivatekeyKey:     c.Privatekey,
	}); err != nil {
		return nil, err
	}

	return provider.NewConstructResult(c)
}

func getPulumiKey(state pulumi.StringOutput, key string) pulumi.AnyOutput {
	return state.ApplyT(func(keys string) (any, error) {
		var c map[string]any

		decoded, err := base64.StdEncoding.DecodeString(keys)
		if err != nil {
			return "", nil
		}

		err = json.Unmarshal(decoded, &c)
		if err != nil {
			return "", err
		}

		return c[key], nil
	}).(pulumi.AnyOutput)
}
