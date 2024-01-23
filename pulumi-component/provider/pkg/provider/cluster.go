package provider

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
)

type Cluster struct {
	pulumi.ResourceState

	HetznerServers pulumi.MapArrayOutput `pulumi:"servers"`
	Kubeconfig     pulumi.StringOutput   `pulumi:"kubeconfig"`
	Privatekey     pulumi.StringOutput   `pulumi:"privatekey"`
}

func (c *Cluster) Type() string { 
	return ComponentName 
}

type ClusterArgs struct {}

func construct(ctx *pulumi.Context, c *Cluster, typ, name string,
	args *ClusterArgs, inputs provider.ConstructInputs, opts ...pulumi.ResourceOption) (*provider.ConstructResult, error) {

	// Ensure we have the right token.
	if et := c.Type(); typ != et {
		return nil, errors.Errorf("unknown resource type %s; expected %s", typ, et)
	}

	// Blit the inputs onto the arguments struct.
	if err := inputs.CopyTo(args); err != nil {
		return nil, errors.Wrap(err, "setting args")
	}

	// Register our component resource.
	if err := ctx.RegisterComponentResource(typ, name, c, opts...); err != nil {
		return nil, err
	}

	opts = append(opts, pulumi.Parent(c))

	cluster, err := phkh.New(ctx, opts)
	if err != nil {
		return nil, err
	}

	deployed, err := cluster.Up()
	if err != nil {
		return nil, err
	}

	c.HetznerServers = pulumi.ToMapArray(deployed.Servers).ToMapArrayOutput()
	c.Kubeconfig = deployed.Kubeconfig
	c.Privatekey = deployed.PrivateKey

	if err := ctx.RegisterResourceOutputs(c, pulumi.Map{
		phkh.HetznerServersKey: c.HetznerServers,
		phkh.KubeconfigKey: c.Kubeconfig,
		phkh.PrivatekeyKey: c.Privatekey,
	}); err != nil {
		return nil, err
	}

	return provider.NewConstructResult(c)
}
