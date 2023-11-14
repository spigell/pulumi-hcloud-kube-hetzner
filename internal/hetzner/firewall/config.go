package firewall

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Config struct {
	dedicated     bool
	dedicatedPool bool
	rules         []*Rule

	Enabled         bool
	AllowICMP       bool `json:"allow-icmp" yaml:"allow-icmp"`
	SSH             *SSH
	AdditionalRules []*Rule `json:"additional-rules" yaml:"additional-rules"`
}

type SSH struct {
	Allow      bool
	DisallowOwnIp bool `json:"disallow-own-ip"`
	AllowedIps []string `json:"allowed-ips" yaml:"allowed-ips"`
}

type Rule struct {
	pulumiSourceIps pulumi.StringArray

	Protocol    string
	Port        string
	SourceIps   []string
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

func (c *Config) AddRules(rules []*Rule) {
	c.rules = append(c.rules, rules...)
}
