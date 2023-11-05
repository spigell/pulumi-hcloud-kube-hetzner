package info

import "github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"

type OSInfo interface {
	// sftpServerPath is a path to sftp-server binary.
	// It is used to transfer files to the server via pulumi-file plugin.
	SFTPServerPath() string
}

type Info struct {
	communicationMethod string
	communicationIface  string
}

func New() *Info {
	return &Info{
		communicationMethod: variables.DefaultCommunicationMethod,
		communicationIface:  variables.Ifaces[variables.DefaultCommunicationMethod],
	}
}

func (i *Info) WithCommunicationMethod(method string) *Info {
	i.communicationMethod = method
	i.communicationIface = variables.Ifaces[method]

	return i
}

func (i *Info) CommunicationMethod() string {
	return i.communicationMethod
}

func (i *Info) CommunicationIface() string {
	return i.communicationIface
}
