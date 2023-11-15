package info

import "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"

type OSInfo interface {
	// sftpServerPath is a path to sftp-server binary.
	// It is used to transfer files to the server via pulumi-file plugin.
	SFTPServerPath() string
}

type Info struct {
	leader bool

	communicationMethod string
	communicationIface  string

	k8sEndpointType string
}

func New() *Info {
	return &Info{
		communicationMethod: variables.PublicCommunicationMethod,
		communicationIface:  variables.Ifaces[variables.PublicCommunicationMethod],
	}
}

func (i *Info) WithCommunicationMethod(method string) *Info {
	i.communicationMethod = method
	i.communicationIface = variables.Ifaces[method]

	return i
}

func (i *Info) WithK8SEndpointType(t string) *Info {
	i.k8sEndpointType = t

	return i
}

func (i *Info) MarkAsLeader() *Info {
	i.leader = true

	return i
}

func (i *Info) CommunicationMethod() string {
	return i.communicationMethod
}

func (i *Info) CommunicationIface() string {
	return i.communicationIface
}

func (i *Info) Leader() bool {
	return i.leader
}
