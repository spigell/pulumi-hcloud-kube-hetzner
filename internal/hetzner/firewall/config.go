package firewall

type Config struct {
	dedicated bool
	rules     []*Rule

	Enabled         bool
	AllowICMP       bool `json:"allow-icmp" yaml:"allow-icmp"`
	SSH             *SSH
	AdditionalRules []*Rule `json:"additional-rules" yaml:"additional-rules"`
}

type SSH struct {
	Allow     bool
	SourceIps []string
}

type Rule struct {
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

func (c *Config) AddRules(rules []*Rule) {
	c.AdditionalRules = append(c.AdditionalRules, rules...)
}
