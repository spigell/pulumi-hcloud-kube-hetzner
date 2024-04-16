package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingleServer(t *testing.T) {
	single := &Config{
		Nodepools: &NodepoolsConfig{
			Servers: []*NodepoolConfig{
				{
					ID: "servers",
					Nodes: []*NodeConfig{
						{
							ID: "server01",
						},
					},
				},
			},
		},
	}

	single.Nodepools.SpecifyLeader()
	assert.Equal(t, true, single.Nodepools.Servers[0].Nodes[0].Leader)

	multi := &Config{
		Nodepools: &NodepoolsConfig{
			Servers: []*NodepoolConfig{
				{
					ID: "servers",
					Nodes: []*NodeConfig{
						{
							ID: "server01",
						},
						{
							ID: "server02",
						},
					},
				},
			},
		},
	}

	multi.Nodepools.SpecifyLeader()
	assert.Equal(t, false, multi.Nodepools.Servers[0].Nodes[0].Leader)
	assert.Equal(t, false, multi.Nodepools.Servers[0].Nodes[1].Leader)
}
