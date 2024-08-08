package journald

var uploadTemplate = `

[Upload]
URL=https://%s:19532

TrustedCertificateFile=%s
ServerCertificateFile=%s
ServerKeyFile=%s

`

var receiverServiceTemplate = `

[Unit]
Description=Journal Remote Sink Service
Documentation=man:systemd-journal-remote(8) man:journal-remote.conf(5)
Requires=systemd-journal-remote.socket

[Service]
ExecStart=/usr/lib/systemd/systemd-journal-remote --listen-https=-3 --output=/var/log/journal/remote/
LockPersonality=yes
LogsDirectory=journal/remote
MemoryDenyWriteExecute=yes
NoNewPrivileges=yes
PrivateDevices=yes
PrivateNetwork=yes
PrivateTmp=yes
ProtectProc=invisible
ProtectClock=yes
ProtectControlGroups=yes
ProtectHome=yes
ProtectHostname=yes
ProtectKernelLogs=yes
ProtectKernelModules=yes
ProtectKernelTunables=yes
ProtectSystem=strict
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6
RestrictNamespaces=yes
RestrictRealtime=yes
RestrictSUIDSGID=yes
SystemCallArchitectures=native
WatchdogSec=3min

[Install]
Also=systemd-journal-remote.socket

`

var receiverConfigTemplate = `

[Remote]

SplitMode=host
TrustedCertificateFile=%s
ServerCertificateFile=%s
ServerKeyFile=%s

MaxUse=3G

`
