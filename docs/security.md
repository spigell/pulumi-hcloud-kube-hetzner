# Security

## Audit logs
Audit logs are enabled by default and can be disabled by setting the `k8s.audit-log.enabled ` to `false`. The audit logs are stored in `/var/lib/rancher/k3s/server/logs` directory on server nodes.

User can configure their own audit log policy by setting the `k8s.audit-log.policy-file-path` to the path of the policy file. The policy file is a YAML file that defines the audit log policy.

More information on audit logs can be found in the [official Kubernetes documentation](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/).
