package connection

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	remotefile "github.com/spigell/pulumi-file/sdk/go/file/remote"
)

type Connection struct {
	IP         pulumi.StringOutput
	PrivateKey pulumi.StringOutput
	User       string
}

func (c *Connection) RemoteCommand() *remote.ConnectionArgs {
	return &remote.ConnectionArgs{
		Host:           c.IP,
		User:           pulumi.String(c.User),
		PrivateKey:     c.PrivateKey,
		DialErrorLimit: pulumi.Int(20),
	}
}

func (c *Connection) RemoteFile() *remotefile.ConnectionArgs {
	return &remotefile.ConnectionArgs{
		Host:       c.IP,
		User:       pulumi.String(c.User),
		PrivateKey: c.PrivateKey,
	}
}
