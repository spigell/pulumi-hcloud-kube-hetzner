package phkh

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/wireguard"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	KeyPairKey        = "ssh:keypair"
	WGInfoKey         = "wireguard:info"
	k3sTokenKey       = "k3s:token"
	KubeconfigKey     = "kubeconfig"
	WGMasterConKey    = "wireguard:connection"
	HetznerServersKey = "hetzner:servers"
	publicKey         = "PublicKey"
	PrivateKey        = "PrivateKey"
)

type State struct {
	ctx   *pulumi.Context
	Stack *pulumi.StackReference
}

func state(ctx *pulumi.Context) (*State, error) {
	self, err := pulumi.NewStackReference(ctx, fmt.Sprintf("%s/%s/%s", "spigell", ctx.Project(), ctx.Stack()), nil)
	if err != nil {
		return nil, err
	}

	return &State{
		ctx:   ctx,
		Stack: self,
	}, nil
}

func (s *State) hetznerInfra() (*hetzner.Deployed, error) {
	info := &hetzner.Deployed{Servers: make(map[string]*hetzner.Server)}

	decoded, err := s.Stack.GetOutputDetails(HetznerServersKey)
	if err != nil {
		return nil, err
	}

	mapped, ok := decoded.SecretValue.(map[string]interface{})
	if !ok {
		// Do not return an error code, because it is not an error.
		// We do not have any server info in the state yet.
		return info, nil
	}

	for k, v := range mapped {
		p, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("error while decoding server info")
		}

		if p["local-password"] == nil {
			p["local-password"] = ""
		}

		info.Servers[k] = &hetzner.Server{
			Connection: &connection.Connection{
				IP:   pulumi.String(p["ip"].(string)).ToStringOutput(),
				User: p["user"].(string),
			},
			LocalPassword: p["local-password"].(string),
		}
	}

	return info, nil
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

func (s *State) sshKeyPair() (*pulumi.StringOutput, error) {
	created, err := keypair.NewECDSA()
	if err != nil {
		return nil, err
	}

	cmd, _ := json.MarshalIndent(created, "  ", "  ")

	out, err := local.NewCommand(s.ctx, "store-generated-ssh-key", &local.CommandArgs{
		Create: pulumi.Sprintf("echo '%s'", cmd),
	}, pulumi.IgnoreChanges([]string{"create"}))

	if err != nil {
		return nil, err
	}

	return &out.Stdout, nil
}

func (s *State) exportSSHKeyPair(keyPair *keypair.ECDSAKeyPair) {
	s.ctx.Export(KeyPairKey, pulumi.ToSecret(pulumi.ToMap(
		map[string](interface{}){
			PrivateKey: keyPair.PrivateKey,
			publicKey:  keyPair.PublicKey,
		},
	)))
}

func (s *State) wgInfo() (map[string]*wireguard.WgConfig, error) {
	info := make(map[string]*wireguard.WgConfig)
	decoded, err := s.Stack.GetOutputDetails(WGInfoKey)
	if err != nil {
		return nil, err
	}

	mapped, ok := decoded.SecretValue.(map[string]interface{})
	if !ok {
		// Do not return an error code, because it is not an error.
		// We do not have any wireguard info in the state yet.
		return info, nil
	}

	for k, v := range mapped {
		p := v.(map[string]interface{})
		info[k] = &wireguard.WgConfig{
			Interface: wireguard.WgInterface{
				Address:    p["ip"].(string),
				PrivateKey: p[PrivateKey].(string),
			},
		}
	}

	return info, nil
}

func (s *State) exportWGInfo(cluster *system.WgCluster) {
	s.ctx.Export(WGInfoKey, pulumi.ToSecret(cluster.Peers.ToMapOutput().ApplyT(func(v map[string]interface{}) map[string]map[string]string {
		m := make(map[string]map[string]string)
		for name, cfg := range v {
			p := cfg.(*wireguard.WgConfig)

			pk, _ := wgtypes.ParseKey(p.Interface.PrivateKey)
			m[name] = make(map[string]string)
			m[name]["ip"] = p.Interface.Address
			m[name][PrivateKey] = p.Interface.PrivateKey
			m[name][publicKey] = pk.PublicKey().String()
		}
		return m
	}).(pulumi.StringMapMapOutput)))

	s.ctx.Export(WGMasterConKey, pulumi.ToSecret(cluster.MasterConnection))
}

func (s *State) exportK3SToken(token string) {
	s.ctx.Export(k3sTokenKey, pulumi.ToSecret(pulumi.String(token)))
}

func (s *State) k3sToken() (string, error) {
	decoded, err := s.Stack.GetOutputDetails(k3sTokenKey)
	if err != nil {
		return "", err
	}

	token, ok := decoded.SecretValue.(string)
	if !ok {
		token = utils.GenerateRandomString(48)
	}

	return token, nil
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
