package program

import (
	"errors"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/hetzner/network/ipam"
)

const (
	stateFilePrefix = "state"
)

var (
	// ErrNoStateFile is returned when the state file is not found.
	ErrNoStateFile = errors.New("no state file")
)

type State struct {
	IPAM *ipam.IPAMData `yaml:",omitempty"`
}
