package config

import (
	"reflect"
	"testing"

	"dario.cat/mergo"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/k8s/k8sconfig"
	"github.com/stretchr/testify/assert"
)

func TestSingleServer(t *testing.T) {
	single := &Config{
		Nodepools: &NodepoolsConfig{
			Servers: []*NodepoolConfig{
				{
					PoolID: "servers",
					Nodes: []*NodeConfig{
						{
							NodeID: "server01",
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
					PoolID: "servers",
					Nodes: []*NodeConfig{
						{
							NodeID: "server01",
						},
						{
							NodeID: "server02",
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

func TestBoolTransformer(t *testing.T) {
	tests := []struct {
		name     string
		dst      *bool // Destination value
		src      *bool // Source value
		expected *bool // Expected result after merge
	}{
		// Inherance from upper levels
		{"src true, dst nil", nil, ptrBool(true), ptrBool(true)},
		{"src false, dst nil", nil, ptrBool(false), ptrBool(false)},
		// Specified on this level
		{"src nil, dst true", ptrBool(true), nil, ptrBool(true)},
		// Get the dst value if defined
		{"src true, dst false", ptrBool(false), ptrBool(true), ptrBool(false)},
		{"src false, dst true", ptrBool(true), ptrBool(false), ptrBool(true)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := &k8sconfig.TaintConfig{
				DisableDefaultsTaints: tt.dst,
			}
			src := &k8sconfig.TaintConfig{
				DisableDefaultsTaints: tt.src,
			}

			if err := mergo.Merge(dst, src, mergo.WithTransformers(BoolTransformer{})); err != nil {
				t.Errorf("Merge failed: %v", err)
			}

			if !reflect.DeepEqual(dst.DisableDefaultsTaints, tt.expected) {
				t.Errorf("Failed %s, expected value to be %v, got %v", tt.name, tt.expected, dst.DisableDefaultsTaints)
			}
		})
	}
}

// Helper function to easily create *bool.
func ptrBool(b bool) *bool {
	return &b
}
