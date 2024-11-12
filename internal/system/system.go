package system

import (
	"fmt"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/info"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/os"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/os/microos"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type System struct {
	ctx  *program.Context
	info *info.Info
	// kubeDependecies hidden storage for keeping dependencies between modules in k8s stage.
	// For instance, wait leader to be ready before joining nodes.
	kubeDependecies map[string][]pulumi.Resource

	ID      string
	KeyPair *keypair.ECDSAKeyPair
	OS      os.OperatingSystem
}

type SysProvisioned struct {
	OS os.Provisioned
}

func New(ctx *program.Context, id string) *System {
	return &System{
		ID:  id,
		ctx: ctx,
		//		KeyPair: pair,
		info: info.New(),
	}
}

func (s *System) MicroOS() *microos.MicroOS {
	os := microos.New(s.ID)

	return os
}


func (s *System) WithOS(os os.OperatingSystem) *System {
	s.OS = os

	return s
}

func (s *System) WithCommunicationMethod(method variables.CommunicationMethod) *System {
	s.info = s.info.WithCommunicationMethod(method)

	return s
}

func (s *System) CommunicationMethod() variables.CommunicationMethod {
	return s.info.CommunicationMethod()
}

func (s *System) WithK8SEndpointType(t string) *System {
	s.info = s.info.WithK8SEndpointType(t)

	return s
}

func (s *System) MarkAsLeader() *System {
	s.info = s.info.MarkAsLeader()

	return s
}

func (s *System) Up(server *hetzner.Server) (*SysProvisioned, error) {
	os, err := s.OS.Up(s.ctx, server, s.kubeDependecies)
	if err != nil {
		err = fmt.Errorf("error while preparing: %w", err)
		return nil, err
	}

	return &SysProvisioned{
		OS: os,
	}, nil
}
