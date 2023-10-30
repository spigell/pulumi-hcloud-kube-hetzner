package server

import (
	"errors"
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gopkg.in/yaml.v3"
)

var ErrMarshalYaml = errors.New("yaml marshal error")

type CloudConfig struct {
	SSHPwauth bool `yaml:"ssh_pwauth"`
	Users     []*CloudConfigUserCloudConfig
	Hostname  string
	Chpasswd  *CloudConfigChpasswd
	GrowPart  *CloudConfigGrowPartConfig
	// This is internal field for storing pulumi input
	Inputs *CloudConfigPulumiInputs `yaml:"-"`
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
	r := "#cloud-config\n"
	return pulumi.All(c.Inputs.Key).ApplyT(func(args []interface{}) (string, error) {
		key := args[0].(string)
		c.Users[0].SSHAuthorizedKeys = []string{
			key,
		}

		cfg, err := yaml.Marshal(&c)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrMarshalYaml, err)
		}

		return r + string(cfg), nil
	}).(pulumi.StringOutput)
}
