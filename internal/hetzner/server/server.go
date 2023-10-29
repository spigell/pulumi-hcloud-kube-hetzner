package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"pulumi-hcloud-kube-hetzner/internal/config"
	"pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	autoApiApps "github.com/spigell/pulumi-automation-api-apps/common/pulumi"
	"github.com/spigell/pulumi-automation-api-apps/hetzner-snapshots-manager/sdk/snapshots"
)

const (
	defaultServerType = "cx21"
	defaultLocation   = "hel1"
	defaultUserName   = "rancher"

	// Allow user to be superuser.
	sudo = "ALL=(ALL) NOPASSWD:ALL"
)

var ErrUserDataRender = errors.New("userdata render error")

type Server struct {
	Config   *config.Server
	Userdata *CloudConfig
}

func New(srv *config.Server, keys *keypair.ECDSAKeyPair) *Server {
	if srv.ServerType == "" {
		srv.ServerType = defaultServerType
	}

	if srv.Location == "" {
		srv.Location = defaultLocation
	}
	if srv.UserName == "" {
		srv.UserName = defaultUserName
	}

	userdata := &CloudConfig{
		GrowPart: &GrowPartConfig{
			Devices: []string{
				"/var",
			},
		},
		Users: []*UserCloudConfig{
			{
				Name: srv.UserName,
				Sudo: sudo,
				SSHAuthorizedKeys: []string{
					keys.PublicKey,
				},
				Passwd: srv.UserPasswd,
			},
		},
	}
	return &Server{
		Config:   srv,
		Userdata: userdata,
	}
}

func (s *Server) Validate() error {
	_, err := s.Userdata.render()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUserDataRender, err)
	}
	return nil
}

func (s *Server) Up(ctx *pulumi.Context, id string) (*hcloud.Server, error) {
	name := fmt.Sprintf("%s-%s-%s", ctx.Project(), ctx.Stack(), id)
	s.Userdata.Hostname = name

	// Error is already checked.
	ud, _ := s.Userdata.render()

	args := &hcloud.ServerArgs{
		ServerType: pulumi.String(s.Config.ServerType),
		Location:   pulumi.String(s.Config.Location),
		Name:       pulumi.String(name),
		UserData:   pulumi.String(ud),
		Image:      pulumi.String(s.Config.Image),
	}

	if os.Getenv(autoApiApps.EnvAutomaionAPIAddr) != "" {
		sn, err := snapshots.GetLastSnapshot(&http.Client{}, name)
		if err != nil {
			switch {
			case errors.Is(err, snapshots.ErrSnapshotNotFound):
			default:
				return nil, fmt.Errorf("get uncovered error for last snapshot: %w", err)
			}
		}
		args.Image = pulumi.String(rune(sn.Body.ID))
	}

	created, err := hcloud.NewServer(ctx, id, args)
	if err != nil {
		return nil, err
	}

	return created, nil
}
