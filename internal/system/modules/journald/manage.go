package journald

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	remotefile "github.com/spigell/pulumi-file/sdk/go/file/remote"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/info"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/system/modules"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/pki"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/utils/ssh/connection"
)

var enableSystemdServiceCMD = strings.Join([]string{
	"sudo systemctl daemon-reload",
	"sudo systemctl disable --now %[1]s",
	"sudo systemctl enable --now %[1]s",
	"sudo systemctl status %[1]s",
	"echo 'systemctl status command returned' $? exit code",
}, " && ")

var disableSystemdServiceCMD = strings.Join([]string{
	"sudo systemctl daemon-reload",
	"sudo systemctl disable --now %[1]s",
}, " && ")

var (
	jourlandCAPemFile      = "/etc/systemd/journald.conf.d/ca.pem"
	uploaderCertPath       = "/etc/systemd/journal-upload.conf.d/cert.pem"
	uploaderPrivateKeyPath = "/etc/systemd/journal-upload.conf.d/key.pem"
	receiverCertPath       = "/etc/systemd/journal-remote.conf.d/cert.pem"
	receiverPrivateKeyPath = "/etc/systemd/journal-remote.conf.d/key.pem"
)

var auditCMD = strings.Join([]string{
	"sudo systemctl %s systemd-journald-audit.socket",
	"sudo systemctl restart systemd-journald",
	"sudo systemctl restart auditd",
}, " && ")

// Up configures journald.
func (j *JournalD) Up(ctx *program.Context, con *connection.Connection, deps []pulumi.Resource, _ []interface{}) (modules.Output, error) {
	resources := make([]pulumi.Resource, 0)

	auditAction := "enable"
	if !*j.Config.GatherAuditD {
		auditAction = "disable --now"
	}

	auditd, err := program.PulumiRun(ctx, remote.NewCommand, fmt.Sprintf("manage-auditd-socket:%s", j.ID), &remote.CommandArgs{
		Connection: con.RemoteCommand(),
		Create:     pulumi.Sprintf(auditCMD, auditAction),
	}, pulumi.DependsOn(deps), pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return nil, fmt.Errorf("failed to manage journald-auditd integration: %w", err)
	}
	resources = append(resources, auditd)

	outputs := &Outputs{}

	if j.System.Leader() {
		issuer, err := pki.New(ctx, "journald-ca")
		if err != nil {
			return nil, fmt.Errorf("failed to create journald pki: %w", err)
		}
		cert, err := issuer.NewCertificate(fmt.Sprintf("journald-receiver:%s", j.ID), []string{"server_auth"},
			pki.WithIPAddesses(pulumi.StringArray{
				j.System.LeaderIP(),
			}),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create server cert: %w", err)
		}

		receiver, err := j.setupReceiver(ctx, con, cert, resources)
		if err != nil {
			return nil, fmt.Errorf("failed to create journald recevier: %w", err)
		}

		resources = append(resources, receiver)

		journalDleader := &info.JournaldLeader{
			Issuer:  issuer,
			Restart: receiver,
		}

		j.System = j.System.WithJournaldLeader(journalDleader)
		outputs.JournalDLeader = journalDleader
	}

	_, err = program.PulumiRun(ctx, remotefile.NewFile, fmt.Sprintf("add-journald-ca-cert:%s", j.ID), &remotefile.FileArgs{
		Connection:  con.RemoteFile(),
		UseSudo:     pulumi.Bool(true),
		Path:        pulumi.String(jourlandCAPemFile),
		Content:     j.System.JournaldLeader().Issuer.CertificatePem,
		SftpPath:    pulumi.String(j.OS.SFTPServerPath()),
		Permissions: pulumi.String("644"),
	}, pulumi.RetainOnDelete(true), pulumi.DependsOn(deps))
	if err != nil {
		return nil, fmt.Errorf("failed to deploy journald CA cert: %w", err)
	}

	cert, err := j.System.JournaldLeader().Issuer.NewCertificate(
		fmt.Sprintf("journald-uploader:%s", j.ID), []string{"server_auth"},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create journald uploader cert: %w", err)
	}

	uploader, err := j.setupUploader(ctx, con, cert, j.System.JournaldLeader().Restart)
	if err != nil {
		return nil, fmt.Errorf("failed to create journald uploader: %w", err)
	}
	resources = append(resources, uploader...)

	return &Provisioned{
		resources: resources,
		outputs:   outputs,
	}, nil
}

func (j *JournalD) setupReceiver(ctx *program.Context, con *connection.Connection, cert *pki.Certificate, deps []pulumi.Resource) (*remote.Command, error) {
	serviceName := "systemd-journal-remote.service"
	socketName := "systemd-journal-remote.socket"

	certDeployed, err := program.PulumiRun(ctx, remotefile.NewFile, fmt.Sprintf("add-journald-receiver-cert:%s", j.ID), &remotefile.FileArgs{
		Connection:  con.RemoteFile(),
		UseSudo:     pulumi.Bool(true),
		Path:        pulumi.String(receiverCertPath),
		Content:     cert.CertificatePem,
		SftpPath:    pulumi.String(j.OS.SFTPServerPath()),
		Permissions: pulumi.String("644"),
	}, pulumi.RetainOnDelete(true), pulumi.DependsOn(deps))
	if err != nil {
		return nil, fmt.Errorf("failed to deploy receiver cert: %w", err)
	}

	keyDeployed, err := program.PulumiRun(ctx, remotefile.NewFile, fmt.Sprintf("add-journald-receiver-privatekey:%s", j.ID), &remotefile.FileArgs{
		Connection:  con.RemoteFile(),
		UseSudo:     pulumi.Bool(true),
		Path:        pulumi.String(receiverPrivateKeyPath),
		Content:     cert.PrivateKeyPem,
		SftpPath:    pulumi.String(j.OS.SFTPServerPath()),
		Permissions: pulumi.String("600"),
	}, pulumi.RetainOnDelete(true), pulumi.DependsOn(deps))
	if err != nil {
		return nil, fmt.Errorf("failed to deploy receiver private key: %w", err)
	}

	service, err := program.PulumiRun(ctx, remotefile.NewFile, fmt.Sprintf("add-journald-remote-service:%s", j.ID), &remotefile.FileArgs{
		Connection:  con.RemoteFile(),
		UseSudo:     pulumi.Bool(true),
		Path:        pulumi.Sprintf("/etc/systemd/system/%s", serviceName),
		Content:     pulumi.String(receiverServiceTemplate),
		SftpPath:    pulumi.String(j.OS.SFTPServerPath()),
		Permissions: pulumi.String("644"),
	}, pulumi.RetainOnDelete(true), pulumi.DependsOn(deps))
	if err != nil {
		return nil, fmt.Errorf("failed to create systemd unit configuration file: %w", err)
	}

	config, err := program.PulumiRun(ctx, remotefile.NewFile, fmt.Sprintf("add-journald-remote-config:%s", j.ID), &remotefile.FileArgs{
		Connection:  con.RemoteFile(),
		UseSudo:     pulumi.Bool(true),
		Path:        pulumi.String("/etc/systemd/journal-remote.conf.d/phkh.conf"),
		Content:     pulumi.Sprintf(receiverConfigTemplate, jourlandCAPemFile, certDeployed.Path, keyDeployed.Path),
		SftpPath:    pulumi.String(j.OS.SFTPServerPath()),
		Permissions: pulumi.String("600"),
	}, pulumi.RetainOnDelete(true))
	if err != nil {
		return nil, fmt.Errorf("failed to create receiver configuration file: %w", err)
	}

	enable, err := program.PulumiRun(ctx, remote.NewCommand, fmt.Sprintf("manage-journald-remote-service:%s", j.ID), &remote.CommandArgs{
		Connection: con.RemoteCommand(),
		Create:     pulumi.Sprintf(enableSystemdServiceCMD, socketName),
		Triggers: pulumi.Array{
			service.Md5sum,
			service.Connection,
			service.Path,
			certDeployed.Md5sum,
			certDeployed.Connection,
			certDeployed.Permissions,
			certDeployed.Path,
			keyDeployed.Md5sum,
			keyDeployed.Connection,
			keyDeployed.Permissions,
			keyDeployed.Path,
			config.Md5sum,
			config.Connection,
			config.Path,
		},
	},
		pulumi.DeleteBeforeReplace(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to manage journald remote server: %w", err)
	}

	return enable, nil
}

func (j *JournalD) setupUploader(ctx *program.Context, con *connection.Connection, cert *pki.Certificate, receiver *remote.Command) ([]pulumi.Resource, error) {
	serviceName := "systemd-journal-upload.service"

	certDeployed, err := program.PulumiRun(ctx, remotefile.NewFile, fmt.Sprintf("add-journald-uploader-cert:%s", j.ID), &remotefile.FileArgs{
		Connection:  con.RemoteFile(),
		UseSudo:     pulumi.Bool(true),
		Path:        pulumi.String(uploaderCertPath),
		Content:     cert.CertificatePem,
		SftpPath:    pulumi.String(j.OS.SFTPServerPath()),
		Permissions: pulumi.String("644"),
	}, pulumi.RetainOnDelete(true), pulumi.DependsOn([]pulumi.Resource{receiver}))
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh configuration file: %w", err)
	}

	keyDeployed, err := program.PulumiRun(ctx, remotefile.NewFile, fmt.Sprintf("add-journald-uploader-privatekey:%s", j.ID), &remotefile.FileArgs{
		Connection: con.RemoteFile(),
		UseSudo:    pulumi.Bool(true),
		Path:       pulumi.String(uploaderPrivateKeyPath),
		Content:    cert.PrivateKeyPem,
		SftpPath:   pulumi.String(j.OS.SFTPServerPath()),
		// TO DO: fix owner
		Permissions: pulumi.String("644"),
	}, pulumi.RetainOnDelete(true), pulumi.DependsOn([]pulumi.Resource{receiver}))
	if err != nil {
		return nil, fmt.Errorf("failed to create uploader cert: %w", err)
	}
	config, err := program.PulumiRun(ctx, remotefile.NewFile, fmt.Sprintf("add-journald-upload-config:%s", j.ID), &remotefile.FileArgs{
		Connection: con.RemoteFile(),
		UseSudo:    pulumi.Bool(true),
		Path:       pulumi.String("/etc/systemd/journal-upload.conf.d/phkh.conf"),
		Content: pulumi.Sprintf(uploadTemplate,
			j.System.LeaderIP(),
			jourlandCAPemFile,
			certDeployed.Path,
			keyDeployed.Path,
		),
		SftpPath:    pulumi.String(j.OS.SFTPServerPath()),
		Permissions: pulumi.String("644"),
	}, pulumi.RetainOnDelete(true), pulumi.DependsOn([]pulumi.Resource{receiver}))
	if err != nil {
		return nil, fmt.Errorf("failed to create uploader configuration file: %w", err)
	}

	cmd := disableSystemdServiceCMD
	if *j.Config.GatherToLeader {
		cmd = enableSystemdServiceCMD
	}

	status, err := program.PulumiRun(ctx, remote.NewCommand, fmt.Sprintf("manage-journald-upload-service:%s", j.ID), &remote.CommandArgs{
		Connection: con.RemoteCommand(),
		Create:     pulumi.Sprintf(cmd, serviceName),
		Triggers: pulumi.Array{
			receiver.Create,
			config.Md5sum,
			config.Connection,
			config.Path,
			certDeployed.Md5sum,
			certDeployed.Connection,
			certDeployed.Permissions,
			certDeployed.Path,
			keyDeployed.Md5sum,
			keyDeployed.Connection,
			keyDeployed.Permissions,
			keyDeployed.Path,
		},
	},
		pulumi.DeleteBeforeReplace(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to manage uploader service: %w", err)
	}

	return []pulumi.Resource{
		config,
		status,
	}, nil
}
