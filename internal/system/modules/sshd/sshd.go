package sshd

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	remotefile "github.com/spigell/pulumi-file/sdk/go/file/remote"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/os/info"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"
)

type SSHD struct {
	order int
	ID    string
	OS    info.Info

	Config *Config
}

type Provisioned struct {
	resources []pulumi.Resource
}

func New(id string, os info.Info, config *Config) *SSHD {
	return &SSHD{
		ID:     id,
		OS:     os,
		Config: config,
	}
}

func (s *SSHD) SetOrder(order int) {
	s.order = order
}

func (s *SSHD) Order() int {
	return s.order
}

// Up configures sshd.
// It deletes default sshd config file and creates new one with config provided in Config.
func (s *SSHD) Up(ctx *pulumi.Context, con *connection.Connection, deps []pulumi.Resource) (modules.Output, error) {
	resources := make([]pulumi.Resource, 0)

	// Delete default sshd config file.
	// It blocks SetEnv from working.
	deleted, err := remote.NewCommand(ctx, fmt.Sprintf("delete-default-sshd-%s", s.ID), &remote.CommandArgs{
		Connection: con.RemoteCommand(),
		Create:     pulumi.String("sudo rm -rfv /etc/ssh/sshd_config"),
	}, pulumi.DependsOn(deps),
		pulumi.DeleteBeforeReplace(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh configuration file: %w", err)
	}
	resources = append(resources, deleted)

	deployed, err := remotefile.NewFile(ctx, fmt.Sprintf("add-sshd-config-%s", s.ID), &remotefile.FileArgs{
		Connection: con.RemoteFile(),
		UseSudo:    pulumi.Bool(true),
		Path:       pulumi.String("/etc/ssh/sshd_config.d/phkh.conf"),
		Content:    pulumi.String(s.Config.String()),
		SftpPath:   pulumi.String(s.OS.SFTPServerPath()),
	}, pulumi.RetainOnDelete(true), pulumi.DependsOn(deps))
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh configuration file: %w", err)
	}
	resources = append(resources, deployed)

	restarted, err := remote.NewCommand(ctx, fmt.Sprintf("restart-sshd-%s", s.ID), &remote.CommandArgs{
		Connection: con.RemoteCommand(),
		Create:     pulumi.String("sudo systemctl restart sshd"),
		Triggers: pulumi.Array{
			deployed.Md5sum,
			deployed.Permissions,
			deployed.Connection,
			deployed.Path,
			deployed.Connection,
			deleted.Create,
		},
	}, pulumi.DependsOn(resources),
		pulumi.DeleteBeforeReplace(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to restart sshd: %w", err)
	}

	resources = append(resources, restarted)

	return &Provisioned{
		resources: resources,
	}, nil
}

func (p *Provisioned) Value() interface{} {
	return nil
}

func (p *Provisioned) Resources() []pulumi.Resource {
	return p.resources
}
