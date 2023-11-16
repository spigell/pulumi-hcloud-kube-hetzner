package mcc

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	Name	     = "mcc"
	enabledByDefault = false
)

type Config struct {
	Enabled bool
	Version string
}

type MCC struct {
	enabled bool
	clusterCIDR string
}

func New(cfg *Config) *MCC {
	var m *MCC

	if cfg == nil {
		m = &MCC{
			enabled: enabledByDefault,
		}
	}

	return m 
}

func (m *MCC) Name() string {
	return Name
}

func (m *MCC) IsEnabled() bool {
	return m.enabled
}

func (m *MCC) IsSupported(distr string) bool {
	switch distr {
	case "k3s":
		return true
	default:
		return false
	}
}

func (m *MCC) SetClusterCIDR (cidr string) {
	m.clusterCIDR = cidr 
}


func (m *MCC) Manage(ctx *pulumi.Context, prov *kubernetes.Provider) error {
	return nil
}