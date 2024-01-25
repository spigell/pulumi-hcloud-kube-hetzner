package phkh

import (
	"fmt"

	"github.com/sanity-io/litter"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/distributions/k3s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/storage/k3stoken"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/storage/sshkeypair"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	// Name of project.
	PhkhKey = "phkh"
	// PrivateKeyKey is the key used to export the private key.
	PrivatekeyKey = "privatekey"
	// KubeconfigKey is the key used to export the kubeconfig.
	KubeconfigKey = "kubeconfig"
	// HetznerServersKey is the key used to export the hetzner servers.
	HetznerServersKey = "servers"
)

// PHKH is the main struct.
type PHKH struct {
	config   *config.Config
	compiled *Compiled
	ctx      *program.Context
}

// Cluster contains the information about the created cluster.
type Cluster struct {
	Servers    []map[string]interface{}
	Kubeconfig pulumi.StringOutput
	PrivateKey pulumi.StringOutput
}

// New creates a new project instance.
// It parses the config and compiles the project.
func New(ctx *pulumi.Context, opts []pulumi.ResourceOption) (*PHKH, error) {
	cfg := config.New(ctx).WithInited()

	context := program.NewContext(ctx, opts...)

	compiled, err := compile(context, cfg)
	if err != nil {
		return nil, err
	}

	return &PHKH{
		config:   cfg,
		compiled: compiled,
		ctx:      context,
	}, nil
}

// Up creates a new cluster and returns some information about it.
func (c *PHKH) Up() (*Cluster, error) {
	keypair, err := sshkeypair.New(c.ctx)
	if err != nil {
		return nil, err
	}

	token, err := k3stoken.New(c.ctx)
	if err != nil {
		return nil, err
	}

	cloud, err := c.compiled.Hetzner.Up(keypair)
	if err != nil {
		return nil, err
	}
	sys, err := c.compiled.SysCluster.Up(token, cloud)
	if err != nil {
		return nil, err
	}

	outputs := pulumi.Map{
		PrivatekeyKey:     pulumi.ToSecret(keypair.PrivateKey()),
		HetznerServersKey: pulumi.ToMapArray(toExportedHetznerServers(cloud)),
	}

	switch distr := c.compiled.K8S.Distr(); distr {
	case k3s.DistrName:
		outputs[KubeconfigKey] = pulumi.ToSecret(toExportedKubeconfig(sys.K3s.KubeconfigForExport))
		err = c.compiled.K8S.Up(sys.K3s.KubeconfigForUsage, sys.Resources)

		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported kubernetes distribution: %s", distr)
	}

	c.ctx.Context().Export(PhkhKey, outputs)

	return &Cluster{
		Servers:    toExportedHetznerServers(cloud),
		Kubeconfig: toExportedKubeconfig(sys.K3s.KubeconfigForExport),
		PrivateKey: keypair.PrivateKey(),
	}, nil
}

// DumpConfig returns a string representation of the parsed config with defaults.
// This is useful for debugging.
func (c *PHKH) DumpConfig() string {
	return litter.Sdump(c.config)
}

func toExportedHetznerServers(deployed *hetzner.Deployed) []map[string]interface{} {
	export := make([]map[string]interface{}, 0)
	for k, v := range deployed.Servers {
		m := make(map[string]interface{})
		m["ip"] = v.Connection.IP
		m["user"] = v.Connection.User
		m["name"] = k

		export = append(export, m)
	}

	return export
}

func toExportedKubeconfig(kube pulumi.AnyOutput) pulumi.StringOutput {
	return kube.ApplyT(
		func(v interface{}) (string, error) {
			kubeconfig := v.(*api.Config)

			k, _ := clientcmd.Write(*kubeconfig)
			return string(k), nil
		},
	).(pulumi.StringOutput)
}
