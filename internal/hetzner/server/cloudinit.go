package server

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

var ErrMarshalYaml = errors.New("yaml marshal error")

type CloudConfig struct {
	SSHPwauth bool `yaml:"ssh_pwauth"`
	Users     []*UserCloudConfig
	Hostname  string
	GrowPart  *GrowPartConfig
}

type UserCloudConfig struct {
	Name              string
	Sudo              string
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
	Passwd            string   `yaml:"passwd,omitempty"`
}

type GrowPartConfig struct {
	Devices []string
}

func (c *CloudConfig) render() (string, error) {
	r := "#cloud-config\n"

	cfg, err := yaml.Marshal(&c)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrMarshalYaml, err)
	}

	return r + string(cfg), nil
}
