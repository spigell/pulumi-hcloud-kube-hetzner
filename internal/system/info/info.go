package info

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/variables"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/pki"
)

type OSInfo interface {
	// sftpServerPath is a path to sftp-server binary.
	// It is used to transfer files to the server via pulumi-file plugin.
	SFTPServerPath() string
}

type Info struct {
	leader   bool
	leaderIP pulumi.StringOutput

	communicationMethod variables.CommunicationMethod
	communicationIface  string

	k8sEndpointType string

	journaldLeader *JournaldLeader
}

type JournaldLeader struct {
	Issuer  *pki.PKI
	Restart *remote.Command
}

func New() *Info {
	return &Info{
		communicationMethod: variables.PublicCommunicationMethod,
		communicationIface:  variables.Ifaces[variables.PublicCommunicationMethod],
	}
}

func (i *Info) WithCommunicationMethod(method variables.CommunicationMethod) *Info {
	i.communicationMethod = method
	i.communicationIface = variables.Ifaces[method]

	return i
}

func (i *Info) WithK8SEndpointType(t string) *Info {
	i.k8sEndpointType = t

	return i
}

func (i *Info) WithLeaderIP(ip pulumi.StringOutput) *Info {
	i.leaderIP = ip

	return i
}

func (i *Info) WithJournaldLeader(leader *JournaldLeader) *Info {
	i.journaldLeader = leader

	return i
}

func (i *Info) K8SEndpointType() string {
	return i.k8sEndpointType
}

func (i *Info) MarkAsLeader() *Info {
	i.leader = true

	return i
}

func (i *Info) CommunicationMethod() variables.CommunicationMethod {
	return i.communicationMethod
}

func (i *Info) CommunicationIface() string {
	return i.communicationIface
}

func (i *Info) Leader() bool {
	return i.leader
}

func (i *Info) LeaderIP() pulumi.StringOutput {
	return i.leaderIP
}

func (i *Info) JournaldLeader() *JournaldLeader {
	return i.journaldLeader
}
