package audit

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AuditLogConfig struct { //nolint: revive
	Enabled           *bool
	PolicyFilePath    string `json:"policy-file-path" yaml:"policy-file-path"`
	AuditLogMaxAge    int    `json:"audit-log-maxage" yaml:"audit-log-maxage"`
	AuditLogMaxBackup int    `json:"audit-log-maxbackup" yaml:"audit-log-maxbackup"`
	AuditLogMaxSize   int    `json:"audit-log-maxsize" yaml:"audit-log-maxsize"`
}

type AuditLog struct { //nolint: revive
	content           *string
	enabled           *bool
	policyFilePath    string `yaml:"policy-file-path"`
	auditLogMaxAge    int    `yaml:"audit-log-maxage"`
	auditLogMaxBackup int    `yaml:"audit-log-maxbackup"`
	auditLogMaxSize   int    `yaml:"audit-log-maxsize"`
}

func NewAuditLog(config *AuditLogConfig) *AuditLog {
	a := &AuditLog{
		enabled:           config.Enabled,
		policyFilePath:    config.PolicyFilePath,
		auditLogMaxAge:    config.AuditLogMaxAge,
		auditLogMaxBackup: config.AuditLogMaxBackup,
		auditLogMaxSize:   config.AuditLogMaxSize,
	}

	a = a.withDefaults()

	if a.policyFilePath != "" {
		file, _ := os.ReadFile(a.policyFilePath)
		a.SetPolicyContent(string(file))
	}

	return a
}

func (a *AuditLog) withDefaults() *AuditLog {
	a.SetPolicyContent(defaultAuditPoilcy)

	if a.enabled == nil {
		t := true
		a.enabled = &t
	}

	if a.auditLogMaxAge == 0 {
		a.auditLogMaxAge = 30
	}

	if a.auditLogMaxBackup == 0 {
		a.auditLogMaxBackup = 10
	}

	if a.auditLogMaxSize == 0 {
		a.auditLogMaxSize = 100
	}

	return a
}

func (a *AuditLog) Validate() error {
	m := make(map[string]interface{})

	if a.policyFilePath != "" {
		file, err := os.ReadFile(a.policyFilePath)
		if err != nil {
			return fmt.Errorf("unable to read policy file: %w", err)
		}
		if err := yaml.Unmarshal(file, m); err != nil {
			return fmt.Errorf("unable to parse default versions file: %w", err)
		}
	}

	return nil
}

func (a *AuditLog) SetPolicyContent(content string) {
	a.content = &content
}

func (a *AuditLog) PolicyContent() *string {
	return a.content
}

func (a *AuditLog) Enabled() bool {
	return *a.enabled
}

func (a *AuditLog) AuditLogMaxAge() int {
	return a.auditLogMaxAge
}

func (a *AuditLog) AuditLogMaxBackup() int {
	return a.auditLogMaxBackup
}

func (a *AuditLog) AuditLogMaxSize() int {
	return a.auditLogMaxSize
}