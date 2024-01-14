package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/config"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/network"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	autoApiApps "github.com/spigell/pulumi-automation-api-apps/common/pulumi"
	"github.com/spigell/pulumi-automation-api-apps/hetzner-snapshots-manager/sdk/snapshots"
)

const (
	defaultServerType = "cx21"
	defaultUserName   = "rancher"
	defaultLocation   = variables.DefaultLocation

	// Allow user to be superuser.
	sudo = "ALL=(ALL) NOPASSWD:ALL"

	// This labels set in build via packer.
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
	KeyName  pulumi.StringOutput
}

type Deployed struct {
	Resource *hcloud.Server
	Password string
}

func New(srv *config.Server, key *hcloud.SshKey) *Server {
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
		Hostname: srv.Hostname,
		GrowPart: &CloudConfigGrowPartConfig{
			Devices: []string{
				"/var",
			},
		},
		Users: []*CloudConfigUserCloudConfig{
			{
				Name: srv.UserName,
				Sudo: sudo,
			},
		},
		Chpasswd: &CloudConfigChpasswd{
			Expire: false,
			Users: []*CloudConfigChpasswdUser{
				{
					Name:     srv.UserName,
					Password: srv.UserPasswd,
				},
			},
		},
	}

	if userdata.Chpasswd.Users[0].Password == "" {
		// Default is hashed password, but we need plain text.
		// TODO: maybe we can use hashed password?
		// I do not how to do it with current knowledges :(
		userdata.Chpasswd.Users[0].Password = utils.GenerateRandomString(12)
	}

	if !strings.HasPrefix(userdata.Chpasswd.Users[0].Password, "$6") {
		userdata.Chpasswd.Users[0].Type = "text"
	}

	userdata.Inputs = &CloudConfigPulumiInputs{
		Key: &key.PublicKey,
	}

	return &Server{
		Config:   srv,
		Userdata: userdata,
		KeyName:  key.Name,
	}
}

func (s *Server) Validate() error {
	return nil
}

func (s *Server) Up(ctx *pulumi.Context, opts []pulumi.ResourceOption, id string, net *network.Deployed, pool string) (*Deployed, error) {
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

	args := &hcloud.ServerArgs{
		ServerType: pulumi.String(s.Config.ServerType),
		Location:   pulumi.String(s.Config.Location),
		Name:       pulumi.String(s.Config.Hostname),
		Image:      image,
		SshKeys: pulumi.StringArray{
			s.KeyName,
		},
	}

	dependencies := make([]pulumi.Resource, 0)
	if net != nil {
		s.Userdata.WriteFiles = append(s.Userdata.WriteFiles, RenameInterfaceScript())
		s.Userdata.RunCMD = append(s.Userdata.RunCMD, RenameInterfaceScript().Path)

		subnet := net.Subnets[pool]

		args.Networks = &hcloud.ServerNetworkTypeArray{
			hcloud.ServerNetworkTypeArgs{
				NetworkId: net.ID,
				Ip:        pulumi.String(subnet.IPs[id]),
			},
		}
		// Rule: id of pool is id of the needed subnet
		dependencies = append(dependencies, net.Subnets[pool].Resource)
	}

	args.UserData = pulumi.ToSecret(s.Userdata.render()).(pulumi.StringOutput)

	if os.Getenv(autoApiApps.EnvAutomaionAPIAddr) != "" {
		sn, err := snapshots.GetLastSnapshot(&http.Client{}, s.Config.Hostname)
		if err != nil {
			switch {
			case errors.Is(err, snapshots.ErrSnapshotNotFound):
			default:
				return nil, fmt.Errorf("get uncovered error for last snapshot: %w", err)
			}
		}
		args.Image = pulumi.String(rune(sn.Body.ID))
	}

	opts = append(opts, pulumi.DependsOn(dependencies))
	opts = append(opts, pulumi.IgnoreChanges([]string{
		"userData",
		"image",
	}))

	created, err := hcloud.NewServer(ctx, id, args, opts...)
	if err != nil {
		return nil, err
	}

	return &Deployed{
		Resource: created,
		Password: s.Userdata.Chpasswd.Users[0].Password,
	}, nil
}
