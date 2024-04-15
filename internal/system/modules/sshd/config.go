package sshd

import (
	"bytes"
	"fmt"

	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils"
)

type Params struct {
	AcceptEnv              string
	PasswordAuthentication bool
	AllowTcpForwarding     bool //nolint:revive,stylecheck // For align with sshd naming
}

func (p *Params) String() string {
	params := make(map[string]string)
	params["PasswordAuthentication"] = boolToString(p.PasswordAuthentication)
	params["AllowTcpForwarding"] = boolToString(p.AllowTcpForwarding)

	if p.AcceptEnv != "" {
		params["AcceptEnv"] = p.AcceptEnv
	}

	b := new(bytes.Buffer)
	for _, k := range utils.SortedMapKeys(params) {
		fmt.Fprintf(b, "%s %s\n", k, params[k])
	}

	return b.String()
}

func boolToString(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
