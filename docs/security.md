# Security

## Audit logs
Audit logs are enabled by default and can be disabled by setting the `k8s.audit-log.enabled ` to `false`. The audit logs are stored in `/var/lib/rancher/k3s/server/logs` directory on server nodes.

User can configure their own audit log policy by setting the `k8s.audit-log.policy-file-path` to the path of the policy file. The policy file is a YAML file that defines the audit log policy.

More information on audit logs can be found in the [official Kubernetes documentation](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/).


## Journald logs
This feature gathers all journald logs, providing comprehensive logging information from the system journal.

It utilizes the journald-remote and journald-upload programs. All connections are encrypted with mTLS. This schema utilizes approximately 3GB of disk space, so it can gather a limited amount of logs.

Auditd is enabled by default for supported OSes. However, the journald socket for auditd is disabled.

The `gather-auditd` option (default is true) can be used to disable gathering.
The `gather-to-leader` option can be used to disable sending logs to the leader node.

```yaml
defaults:
  global:
	os:
	  journald:
	    # By default is true
	    # Auditd is enabled by default but not gathered by journald.
	    gather-auditd: false
	    # By default is true
	    gather-to-leader: false
```

### Usage