package network

const (
	stateFilePrefix = "network-state"
)

func dumpNetworkState(ipam string, destination string) error {

	return nil
}

func loadNetworkStateFile(stack string) ([]*allocatedSubnet, error) {
	// load yaml file from pulumi state

	return []*allocatedSubnet{}, nil

}