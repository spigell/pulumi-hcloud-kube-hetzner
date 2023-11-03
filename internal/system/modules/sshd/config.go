package sshd

import (
	"bytes"
	"fmt"
)

type Config struct {
	AcceptEnv              string
	PasswordAuthentication bool
	AllowTcpForwarding     bool //nolint:revive,stylecheck // For align with sshd naming
}

func (c *Config) String() string {
	params := make(map[string]string)
	params["PasswordAuthentication"] = boolToString(c.PasswordAuthentication)
	params["AllowTcpForwarding"] = boolToString(c.AllowTcpForwarding)

	if c.AcceptEnv != "" {
		params["AcceptEnv"] = c.AcceptEnv
	}

	b := new(bytes.Buffer)
	for k, v := range params {
		fmt.Fprintf(b, "%s %s\n", k, v)
	}

	return b.String()
}

func boolToString(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
