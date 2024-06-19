package osconfig

import (
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules/journald"
)

type OSConfig struct {
	JournalD *journald.Config
}

func (o *OSConfig) WithDefaults() *OSConfig {
	if o.JournalD == nil {
		o.JournalD = &journald.Config{}
	}

	return o
}
