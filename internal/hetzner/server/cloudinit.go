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

type CloudConfig struct {
	SSHPwauth  bool `yaml:"ssh_pwauth"`
	Users      []*CloudConfigUserCloudConfig
	Hostname   string
	Chpasswd   *CloudConfigChpasswd
	GrowPart   *CloudConfigGrowPartConfig
	WriteFiles []*CloudConfigWriteFile `yaml:"write_files,omitempty"`
	RunCMD     []string                `yaml:"runcmd,omitempty"`
	// This is internal field for storing pulumi input
	Inputs *CloudConfigPulumiInputs `yaml:"-"`
}
type CloudConfigWriteFile struct {
	Content     string
	Path        string
	Permissions string
}

type CloudConfigPulumiInputs struct {
	Key *pulumi.StringOutput
}

type CloudConfigUserCloudConfig struct {
	Name              string
	Sudo              string
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
}

type CloudConfigChpasswd struct {
	Expire bool
	Users  []*CloudConfigChpasswdUser
}

type CloudConfigChpasswdUser struct {
	Name     string
	Type     string
	Password string
}

type CloudConfigGrowPartConfig struct {
	Devices []string
}

// render returns rendered cloud-config
// error will be catched by Pulumi if yaml marshal failed.
func (c *CloudConfig) render() pulumi.StringOutput {
	return pulumi.All(c.Inputs.Key).ApplyT(func(args []interface{}) (string, error) {
		key := args[0].(string)
		c.Users[0].SSHAuthorizedKeys = []string{
			key,
		}

		cfg, err := yaml.Marshal(&c)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrMarshalYaml, err)
		}

		return cloudConfigHeader + string(cfg), nil
	}).(pulumi.StringOutput)
}

func RenameInterfaceScript() *CloudConfigWriteFile {
	return &CloudConfigWriteFile{
		Path:        "/etc/cloud/rename_interface.sh",
		Content:     scripts.RenameInterface,
		Permissions: "0755",
	}
}
