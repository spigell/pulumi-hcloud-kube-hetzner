package audit

type AuditLog struct {
	content *string

	Enabled           *bool
	PolicyFile        string `json:"policy-file" yaml:"policy-file"`
	AuditLogMaxAge    int    `json:"audit-log-maxage" yaml:"audit-log-maxage"`
	AuditLogMaxBackup int    `json:"audit-log-maxbackup" yaml:"audit-log-maxbackup"`
	AuditLogMaxSize   int    `json:"audit-log-maxsize" yaml:"audit-log-maxsize"`
}

func (a *AuditLog) WithDefaults() *AuditLog {
	if a.Enabled == nil {
		t := true
		a.Enabled = &t
	}

	if a.AuditLogMaxAge == 0 {
		a.AuditLogMaxAge = 30
	}

	if a.AuditLogMaxBackup == 0 {
		a.AuditLogMaxBackup = 10
	}

	if a.AuditLogMaxSize == 0 {
		a.AuditLogMaxSize = 100
	}

	if a.PolicyFile == "" {
		a.SetPolicyContent(defaultAuditPoilcy)
	}

	return a
}

func (a *AuditLog) SetPolicyContent(content string) {
	a.content = &content
}

func (a *AuditLog) PolicyContent() *string {
	return a.content
}
