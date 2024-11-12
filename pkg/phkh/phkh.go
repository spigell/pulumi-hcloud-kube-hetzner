package phkh

import (
	"errors"
	//"fmt"

	"github.com/sanity-io/litter"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	//"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/distributions/k3s"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	// "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/storage/k3stoken"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/storage/sshkeypair"
	talos "github.com/spigell/pulumi-talos-cluster/sdk/go/talos-cluster"

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
	// ServerIPKey is the key used to export external IP of servers.
	ServerIPKey = "ip"
	// ServerInternalIPKey is the key used to export internal ip of servers.
	ServerInternalIPKey = "internalIP"
	ServerUserKey       = "user"
	ServerNameKey       = "name"
)

// PHKH is the main struct.
type PHKH struct {
	config   *config.Config
	compiled *Compiled
	ctx      *program.Context
}

// Cluster contains the information about the created cluster.
type Cluster struct {
	Servers    pulumi.MapArray
	Kubeconfig pulumi.StringOutput
	Privatekey pulumi.StringOutput
}

// New creates a new project instance.
// It parses the config and compiles the project.
func NewCluster(ctx *pulumi.Context, name string, configuration map[string]any, opts []pulumi.ResourceOption) (*PHKH, error) {
	cfg, err := config.ParseClusterConfig(configuration)
	if err != nil {
		return nil, err
	}

	cfg = cfg.WithInited()

	context, err := program.NewContext(ctx, name, opts...)
	if err != nil {
		if !errors.Is(err, program.ErrNoStateFile) {
			return nil, err
		}
	}

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

	//token, err := k3stoken.New(c.ctx)
	//if err != nil {
	//	return nil, err
	//}

	talosCluster, err := talos.NewCluster(c.ctx.Context(), "cluster", &talos.ClusterArgs{})

	if err != nil {
		return nil, err
	}

	userdata := talosCluster.MachineConfigurations.MapIndex(pulumi.String("test2"))

	cloud, err := c.compiled.Hetzner.Up(keypair, &userdata)
	if err != nil {
		return nil, err
	}

	_, err = talos.NewBootstrap(c.ctx.Context(), "bootstrap", &talos.BootstrapArgs{
		Node: cloud.Servers[c.compiled.SysCluster.Leader().ID].Connection.IP,
		ClientConfiguration: talosCluster.ClientConfiguration,
	})

	// err = talos.Provision(c.ctx, cloud.Servers)
	// if err != nil {
	//	return nil, err
	//}
	// sys, err := c.compiled.SysCluster.Up(token, cloud)
	//if err != nil {
	//	return nil, err
	//}

	//switch distr := c.compiled.K8S.Distr(); distr {
	//case k3s.DistrName:
	//err = c.compiled.K8S.Up(sys.K3s.KubeconfigForUsage, sys.Resources)
	//if err != nil {
	//	return nil, err
	//}

	//default:
	//return nil, fmt.Errorf("unsupported kubernetes distribution: %s", distr)
	//}

	return &Cluster{
		Servers: toExportedHetznerServers(cloud),
		// Kubeconfig: toExportedKubeconfig(sys.K3s.KubeconfigForExport),
		Privatekey: keypair.PrivateKey(),
	}, nil
}

// DumpConfig returns a string representation of the parsed config with defaults.
// This is useful for debugging.
func (c *PHKH) DumpConfig() string {
	return litter.Sdump(c.config)
}

func toExportedHetznerServers(deployed *hetzner.Deployed) pulumi.MapArray {
	export := make([]map[string]interface{}, 0)
	for k, v := range deployed.Servers {
		m := make(map[string]interface{})
		m[ServerIPKey] = v.Connection.IP
		m[ServerInternalIPKey] = v.InternalIP
		m[ServerUserKey] = v.Connection.User
		m[ServerNameKey] = k

		export = append(export, m)
	}

	return pulumi.ToMapArray(export)
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
