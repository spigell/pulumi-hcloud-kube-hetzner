package server

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"pulumi-hcloud-kube-hetzner/internal/config"
	"strings"
	"time"

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
		userdata.Chpasswd.Users[0].Password = generatePassword()
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

func (s *Server) Up(ctx *pulumi.Context, id string) (*Deployed, error) {
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

	args := &hcloud.ServerArgs{
		ServerType: pulumi.String(s.Config.ServerType),
		Location:   pulumi.String(s.Config.Location),
		Name:       pulumi.String(name),
		UserData:   s.Userdata.render(),
		Image:      image,
		SshKeys: pulumi.StringArray{
			s.KeyName,
		},
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

	return &Deployed{
		Resource: created,
		Password: s.Userdata.Chpasswd.Users[0].Password,
	}, nil
}

func generatePassword() string {
	charset := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	//nolint: gosec
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	userPassword := make([]byte, 12)
	for i := range userPassword {
		userPassword[i] = charset[seededRand.Intn(len(charset))]
	}
	fmt.Println("Plain password:" + string(userPassword))

	return string(userPassword)
}
