package firewall

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Config struct {
	// dedicated indicates whether the server is on dedicated hardware.
	dedicated bool

	// dedicatedPool specifies if the server is part of a dedicated pool.
	dedicatedPool bool

	// rules is a slice of pointers to RuleConfig detailing specific configuration rules.
	rules []*RuleConfig

	// Enabled specifies if the configuration is active.
	// Default is false.
	Enabled *bool

	// AllowICMP indicates whether ICMP traffic is allowed.
	// Default is false.
	AllowICMP *bool `json:"allow-icmp" yaml:"allow-icmp" mapstructure:"allow-icmp"`

	// SSH holds the SSH specific configurations.
	SSH *SSHConfig

	// AdditionalRules is a list of additional rules to be applied.
	AdditionalRules []*RuleConfig `json:"additional-rules" yaml:"additional-rules" mapstructure:"additional-rules"`
}

type SSHConfig struct {
	// Allow indicates whether SSH access is permitted.
	// Default is false.
	Allow *bool

	// DisallowOwnIP specifies whether SSH access from the deployer's own IP address is disallowed.
	// Default is false.
	DisallowOwnIP *bool `json:"disallow-own-ip" yaml:"disallow-own-ip" mapstructure:"disallow-own-ip"`

	// AllowedIps lists specific IP addresses that are permitted to access via SSH.
	AllowedIps []string `json:"allowed-ips" yaml:"allowed-ips" mapstructure:"allowed-ips"`
}

type RuleConfig struct {
	// pulumiSourceIps holds a list of source IPs managed by Pulumi, typically used for infrastructure as code deployments.
	pulumiSourceIps pulumi.StringArray

	// Protocol specifies the network protocol (e.g., TCP, UDP) applicable for the rule.
	// Default is TCP.
	Protocol string

	// Port specifies the network port number or range applicable for the rule.
	// Required.
	Port string

	// SourceIps lists IP addresses or subnets from which traffic is allowed or to which traffic is directed, based on the Direction.
	// Required.
	SourceIps []string `json:"source-ips" yaml:"source-ips" mapstructure:"source-ips"`

	// Description provides a human-readable explanation of what the rule is intended to do.
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
