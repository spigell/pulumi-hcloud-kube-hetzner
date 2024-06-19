package journald

var (
	defaultGatherAuditD   = true
	defaultGatherToLeader = true
)

type Config struct {
	// GatherAuditD indicates whether auditd logs should be gathered.
	// Default is true.
	GatherAuditD *bool `json:"gather-auditd" yaml:"gather-auditd" mapstructure:"gather-auditd"`

	// GatherToLeader indicates whether journald logs should be sent to the leader node.
	// Default is true.
	GatherToLeader *bool `json:"gather-to-leader" yaml:"gather-to-leader" mapstructure:"gather-to-leader"`
}

func (c *Config) WithDefaults() *Config {
	if c.GatherAuditD == nil {
		c.GatherAuditD = &defaultGatherAuditD
	}

	if c.GatherToLeader == nil {
		c.GatherToLeader = &defaultGatherToLeader
	}

	return c
}
