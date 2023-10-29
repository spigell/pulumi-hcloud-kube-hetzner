package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"pulumi-hcloud-kube-hetzner/internal/config"
	"pulumi-hcloud-kube-hetzner/internal/utils/ssh/keypair"
	"strings"

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

	// ServerName must be a valid hostname.
	// Since ctx.Project() can be a quite long string, prefix for server name is 4 character.
	serverNamePrefix = "phkh"

	// This labels set in build via packer
	selector = "microos-snapshot=yes"
)

var (
	ErrUserDataRender = errors.New("userdata render error")

	ImageNotFoundSuggestion = fmt.Sprintf(strings.Join([]string{
		"please provide image ID manually in configuration",
		"create image with `%s` selector",
	}, " or "), selector)
	ImageNotFoundMessage = strings.Join([]string{
		"can not obtain image ID automatically",
		"failed to get image",
	}, ": ")
)

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
	name := fmt.Sprintf("%s-%s-%s", serverNamePrefix, ctx.Stack(), id)
	s.Userdata.Hostname = name

	// Get image ID from user input
	image := pulumi.String(s.Config.Image)

	// If image is not provided from user, get latest microos snapshot.
	if s.Config.Image == "" {
		got, err := hcloud.GetImage(ctx, &hcloud.GetImageArgs{
			WithSelector: pulumi.StringRef(selector),
			MostRecent:   pulumi.BoolRef(true),
		})

		if err != nil {
			return nil, fmt.Errorf(
				strings.Join([]string{
					ImageNotFoundMessage,
					ImageNotFoundSuggestion,
					"%w",
				}, ": "), err,
			)
		}

		image = pulumi.String(fmt.Sprintf("%d", got.Id))
	}

	// Error is already checked.
	ud, _ := s.Userdata.render()

	args := &hcloud.ServerArgs{
		ServerType: pulumi.String(s.Config.ServerType),
		Location:   pulumi.String(s.Config.Location),
		Name:       pulumi.String(name),
		UserData:   pulumi.String(ud),
		Image:      image,
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
