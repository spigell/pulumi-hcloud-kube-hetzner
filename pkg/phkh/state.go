package phkh

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	KeyPairKey        = "ssh:keypair"
	k3sTokenKey       = "k3s:token"
	KubeconfigKey     = "kubeconfig"
	HetznerServersKey = "hetzner:servers"
	publicKey         = "PublicKey"
	PrivateKey        = "PrivateKey"
)

type State struct {
	ctx *pulumi.Context
}

func state(ctx *pulumi.Context) (*State, error) {
	return &State{
		ctx: ctx,
	}, nil
}

func (s *State) exportHetznerInfra(deployed *hetzner.Deployed) {
	export := make(map[string]map[string]interface{})
	for k, v := range deployed.Servers {
		export[k] = make(map[string]interface{})
		export[k]["ip"] = v.Connection.IP
		export[k]["user"] = v.Connection.User
		export[k]["local-password"] = v.LocalPassword
	}

	s.ctx.Export(HetznerServersKey, pulumi.ToSecret(export))
}

func (s *State) exportKubeconfig(kube pulumi.AnyOutput) {
	s.ctx.Export(KubeconfigKey, pulumi.ToSecret(kube.ApplyT(
		func(v interface{}) (string, error) {
			kubeconfig := v.(*api.Config)

			k, _ := clientcmd.Write(*kubeconfig)
			return string(k), nil
		},
	).(pulumi.StringOutput)))
}
