package k3s

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChooseDNSIP(t *testing.T) {
	ip, err := chooseDNSIP("192.168.0.0/24")

	require.NoError(t, err)

	require.Equal(t, ip, "192.168.0.10")
}
