package info

type Info interface {
	// sftpServerPath is a path to sftp-server binary.
	// It is used to transfer files to the server via pulumi-file plugin.
	SFTPServerPath() string
}
