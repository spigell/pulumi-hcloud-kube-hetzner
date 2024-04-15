package firewall

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Config struct {
	dedicated     bool
	dedicatedPool bool
	rules         []*RuleConfig

	Enabled         bool
	AllowICMP       bool `json:"allow-icmp" yaml:"allow-icmp"`
	SSH             *SSHConfig
	AdditionalRules []*RuleConfig `json:"additional-rules" yaml:"additional-rules"`
}

type SSHConfig struct {
	Allow         bool
	DisallowOwnIP bool     `json:"disallow-own-ip" yaml:"disallow-own-ip"`
	AllowedIps    []string `json:"allowed-ips" yaml:"allowed-ips"`
}

type RuleConfig struct {
	pulumiSourceIps pulumi.StringArray

	Protocol    string
	Port        string
	SourceIps   []string `json:"source-ips" yaml:"source-ips"`
	Direction   string
	Description string
}

func (c *Config) MarkAsDedicated() {
	c.dedicated = true
}

func (c *Config) Dedicated() bool {
	return c.dedicated
}

func (c *Config) MarkWithDedicatedPool() {
	c.dedicatedPool = true
}

func (c *Config) DedicatedPool() bool {
	return c.dedicatedPool
}

func (c *Config) AddRules(rules []*RuleConfig) {
	c.rules = append(c.rules, rules...)
}
