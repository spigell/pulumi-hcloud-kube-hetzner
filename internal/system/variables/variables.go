// This package contains variables used by all system subpackages.
// Please do not import another packages here.
package variables

const (
	AgentRole  = "agent"
	ServerRole = "server"
	// Name of interfaces.
	PublicIface  = "eth0"
	PrivateIface = "eth1"
	WGIface      = "kubewg0"
	// Name of modules.
	K3s       = "k3s"
	SSHD      = "sshd"
	Wireguard = "wireguard"
	// Communication methods.
)

type CommunicationMethod string

var (
	PublicCommunicationMethod   CommunicationMethod = "public"
	InternalCommunicationMethod CommunicationMethod = "internal"
	WgCommunicationMethod       CommunicationMethod = Wireguard

	Ifaces = map[CommunicationMethod]string{
		PublicCommunicationMethod:   PublicIface,
		InternalCommunicationMethod: PrivateIface,
		WgCommunicationMethod:       WGIface,
	}
)

func (c CommunicationMethod) String() string {
	return string(c)
}

func (c CommunicationMethod) HetznerBased() bool {
	return c != WgCommunicationMethod
}
