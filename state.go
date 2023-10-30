package main

import (
	"fmt"
	"pulumi-hcloud-kube-hetzner/internal/hetzner"
	"pulumi-hcloud-kube-hetzner/internal/system"
	"pulumi-hcloud-kube-hetzner/internal/system/modules/wireguard"
	"pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"
	"pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	keyPairKey        = "ssh:keypair"
	wgInfoKey         = "wireguard:info"
	wgMasterConKey    = "wireguard:connection"
	hetznerServersKey = "hetzer:servers"
	publicKey         = "PublicKey"
	privateKey        = "PrivateKey"
)

type State struct {
	ctx   *pulumi.Context
	Stack *pulumi.StackReference
}

func NewState(ctx *pulumi.Context) (*State, error) {
	self, err := pulumi.NewStackReference(ctx, fmt.Sprintf("%s/%s/%s", ctx.Organization(), ctx.Project(), ctx.Stack()), nil)
	if err != nil {
		return nil, err
	}

	return &State{
		ctx:   ctx,
		Stack: self,
	}, nil
}

func (s *State) HetznerInfra() (*hetzner.Deployed, error) {
	info := &hetzner.Deployed{Servers: make(map[string]*hetzner.Server)}

	decoded, err := s.Stack.GetOutputDetails(hetznerServersKey)
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

func (s *State) ExportHetznerInfra(deployed *hetzner.Deployed) {
	export := make(map[string]map[string]interface{})
	for k, v := range deployed.Servers {
		export[k] = make(map[string]interface{})
		export[k]["ip"] = v.Connection.IP
		export[k]["user"] = v.Connection.User
		export[k]["local-password"] = v.LocalPassword
	}

	s.ctx.Export(hetznerServersKey, pulumi.ToSecret(export))
}

func (s *State) SSHKeyPair() (*keypair.ECDSAKeyPair, error) {
	decoded, err := s.Stack.GetOutputDetails(keyPairKey)
	if err != nil {
		return nil, err
	}

	keys, ok := decoded.SecretValue.(map[string]interface{})
	if !ok {
		created, err := keypair.NewECDSA()
		keys = make(map[string]interface{})
		keys[publicKey] = created.PublicKey
		keys[privateKey] = created.PrivateKey
		if err != nil {
			return nil, err
		}
	}

	return &keypair.ECDSAKeyPair{
		// It can be only strings
		PublicKey:  keys[publicKey].(string),
		PrivateKey: keys[privateKey].(string),
	}, nil
}

func (s *State) ExportSSHKeyPair(keyPair *keypair.ECDSAKeyPair) {
	s.ctx.Export(keyPairKey, pulumi.ToSecret(pulumi.ToMap(
		map[string](interface{}){
			privateKey: keyPair.PrivateKey,
			publicKey:  keyPair.PublicKey,
		},
	)))
}

func (s *State) WGInfo() (map[string]*wireguard.WgConfig, error) {
	info := make(map[string]*wireguard.WgConfig)
	decoded, err := s.Stack.GetOutputDetails(wgInfoKey)
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
				PrivateKey: p["privatekey"].(string),
			},
		}
	}

	return info, nil
}

func (s *State) ExportWGInfo(cluster *system.WgCluster) {
	s.ctx.Export(wgInfoKey, pulumi.ToSecret(cluster.Peers.ToMapOutput().ApplyT(func(v map[string]interface{}) map[string]map[string]string {
		m := make(map[string]map[string]string)
		for name, cfg := range v {
			p := cfg.(*wireguard.WgConfig)

			pk, _ := wgtypes.ParseKey(p.Interface.PrivateKey)
			m[name] = make(map[string]string)
			m[name]["ip"] = p.Interface.Address
			m[name]["privatekey"] = p.Interface.PrivateKey
			m[name]["publickey"] = pk.PublicKey().String()
		}
		return m
	}).(pulumi.StringMapMapOutput)))

	s.ctx.Export(wgMasterConKey, pulumi.ToSecret(cluster.MasterConnection))
}
