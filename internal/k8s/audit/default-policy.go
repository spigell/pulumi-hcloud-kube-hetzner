package audit

const defaultAuditPoilcy = `apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: None
  verbs: ["get", "watch", "list"]

- level: None
  resources:
  - group: "" # core
    resources: ["events"]

# This is trusted users
- level: None
  users:
  - "system:kube-scheduler"
  - "system:kube-proxy"
  - "system:apiserver"
  - "system:kube-controller-manager"

- level: None
  userGroups: ["system:nodes"]
  verbs: ["get", "list"]

- level: RequestResponse
`
