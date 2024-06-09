package journald

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/info"
)

type JournalD struct {
	order  int
	ID     string
	OS     info.OSInfo
	System *info.Info

	Config *Config
}

func New(id string, os info.OSInfo, config *Config) *JournalD {
	return &JournalD{
		ID:     id,
		OS:     os,
		Config: config.WithDefaults(),
	}
}

func (j *JournalD) RequiredPkgs() []string {
	packages := make([]string, 0)
	if *j.Config.GatherToLeader {
		packages = append(packages, "systemd-journal-remote")
	}
	return packages
}

func (j *JournalD) SetOrder(order int) {
	j.order = order
}

func (j *JournalD) Order() int {
	return j.order
}

func (j *JournalD) WithSysInfo(info *info.Info) *JournalD {
	j.System = info

	return j
}

type Provisioned struct {
	resources []pulumi.Resource
	outputs   *Outputs
}

type Outputs struct {
	JournalDLeader *info.JournaldLeader
}

func (p *Provisioned) Resources() []pulumi.Resource {
	return p.resources
}

func (p *Provisioned) Value() any {
	return p.outputs
}
