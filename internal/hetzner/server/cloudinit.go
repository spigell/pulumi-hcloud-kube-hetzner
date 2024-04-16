package server

import (
	"errors"
	"fmt"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/server/scripts"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gopkg.in/yaml.v3"
)

const (
	cloudConfigHeader = "#cloud-config\n"
)

var ErrMarshalYaml = errors.New("yaml marshal error")

type CloudInit struct {
	SSHPwauth  bool `yaml:"ssh_pwauth"`
	Users      []*CloudInitUserCloud
	Hostname   string
	Chpasswd   *CloudInitChpasswd
	GrowPart   *CloudInitGrowPart
	WriteFiles []*CloudInitWriteFile `yaml:"write_files,omitempty"`
	RunCMD     []string              `yaml:"runcmd,omitempty"`
	// This is internal field for storing pulumi input
	Inputs *CloudInitPulumiInputs `yaml:"-"`
}
type CloudInitWriteFile struct {
	Content     string
	Path        string
	Permissions string
}

type CloudInitPulumiInputs struct {
	Key *pulumi.StringOutput
}

type CloudInitUserCloud struct {
	Name              string
	Sudo              string
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
}

type CloudInitChpasswd struct {
	Expire bool
	Users  []*CloudInitChpasswdUser
}

type CloudInitChpasswdUser struct {
	Name     string
	Type     string
	Password string
}

type CloudInitGrowPart struct {
	Devices []string
}

// render returns rendered cloud-config
// error will be catched by Pulumi if yaml marshal failed.
func (c *CloudInit) render() pulumi.StringOutput {
	return pulumi.All(c.Inputs.Key).ApplyT(func(args []interface{}) (string, error) {
		key := args[0].(string)
		c.Users[0].SSHAuthorizedKeys = append(c.Users[0].SSHAuthorizedKeys, key)

		cfg, err := yaml.Marshal(&c)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrMarshalYaml, err)
		}

		return cloudConfigHeader + string(cfg), nil
	}).(pulumi.StringOutput)
}

func RenameInterfaceScript() *CloudInitWriteFile {
	return &CloudInitWriteFile{
		Path:        "/etc/cloud/rename_interface.sh",
		Content:     scripts.RenameInterface,
		Permissions: "0755",
	}
}
