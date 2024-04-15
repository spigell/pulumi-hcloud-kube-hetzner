## config.Config

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| nodepools | [config.*NodepoolsConfig](#confignodepoolsconfig) |  |  |
| defaults | [config.*DefaultConfig](#configdefaultconfig) |  |  |
| network | [config.*NetworkConfig](#confignetworkconfig) |  |  |
| k8s | [*k8sconfig.Config](#k8sconfigconfig) |  |  |

## config.DefaultConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| global | [config.*NodeConfig](#confignodeconfig) |  |  |
| servers | [config.*NodeConfig](#confignodeconfig) |  |  |
| agents | [config.*NodeConfig](#confignodeconfig) |  |  |

## config.NodepoolsConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| servers | [config.[]*NodepoolConfig](#confignodepoolconfig) |  |  |
| agents | [config.[]*NodepoolConfig](#confignodepoolconfig) |  |  |

## config.NodepoolConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| id | string |  |  |
| config | [config.*NodeConfig](#confignodeconfig) |  |  |
| nodes | [config.[]*NodeConfig](#confignodeconfig) |  |  |

## config.NetworkConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| hetzner | [*network.Params](#networkparams) |  |  |

## config.NodeConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| id | string |  |  |
| leader | bool |  |  |
| server | [config.*ServerConfig](#configserverconfig) |  |  |
| k3s | [*k3s.Config](#k3sconfig) |  |  |
| k8s | [*k8sconfig.NodeConfig](#k8sconfignodeconfig) |  |  |
| role | string |  |  |

## config.ServerConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| server-type | string | ServerType specifies the type of server to be provisioned (e.g., "cx11", "cx21"). Default is cx21.  | cx21 |
| hostname | string | Hostname is the desired hostname to assign to the server. Default is 'phkh-<name-of-stack>-<id-of-node>'.  | 'phkh-<name-of-stack>-<id-of-node>' |
| firewall | [config.*FirewallConfig](#configfirewallconfig) | Firewall points to an optional configuration for a firewall to be associated with the server.  |  |
| location | string | Location specifies the physical location or data center where the server will be hosted (e.g., "fsn1"). Default is hel1.  | hel1 |
| additional-ssh-keys | []string | AdditionalSSHKeys contains a list of additional public SSH keys to install in the server's user account. Default is [].  | [] |
| user-name | string | UserName is the primary user account name that will be created on the server. Default is rancher.  | rancher |
| user-password | string | UserPasswd is the password for the primary user account on the server. Default is not configured.  | not configured |
| image | string | Image specifies the operating system image to use for the server (e.g., "ubuntu-20.04" or id of private image). Default is autodiscovered.  | autodiscovered |

## config.FirewallConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| hetzner | [*firewall.Config](#firewallconfig) |  |  |

## firewall.Config

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| enabled | bool |  |  |
| allow-icmp | bool |  |  |
| ssh | [firewall.*SSHConfig](#firewallsshconfig) |  |  |
| additional-rules | [firewall.[]*RuleConfig](#firewallruleconfig) |  |  |

## firewall.SSHConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| allow | bool |  |  |
| disallow-own-ip | bool |  |  |
| allowed-ips | []string |  |  |

## firewall.RuleConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| protocol | string |  |  |
| port | string |  |  |
| source-ips | []string |  |  |
| direction | string |  |  |
| description | string |  |  |

## addons.Config

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| ccm | [*ccm.Config](#ccmconfig) |  |  |
| k3s-upgrade-controller | [*k3supgrader.Config](#k3supgraderconfig) |  |  |

## ccm.Config

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| enabled | bool | Enabled is a flag to enable or disable hcloud CCM.  |  |
| helm | [*helm.Config](#helmconfig) |  |  |
| loadbalancers-enabled | bool | LoadbalancersEnabled is a flag to enable or disable loadbalancers management. Note: internal loadbalancer for k3s will be disabled.  |  |
| loadbalancers-default-location | string | DefaultloadbalancerLocation is a default location for the loadbancers.  |  |
| token | string | Token is a hcloud token to access hcloud API for CCM.  |  |

## upgrader.Config

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| enabled | bool |  |  |
| helm | [*helm.Config](#helmconfig) |  |  |
| target-version | string | Version is a version to use for the upgrade. Conflicts with Channel.  |  |
| target-channel | string | Channel is a channel to use for the upgrade. Conflicts with Version.  |  |
| config-env | []string | ConfigEnv is a map of environment variables to pass to the controller.  |  |

## audit.AuditLogConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| enabled | [audit.*bool](#auditbool) |  |  |
| policy-file-path | string |  |  |
| audit-log-maxage | int |  |  |
| audit-log-maxbackup | int |  |  |
| audit-log-maxsize | int |  |  |

## helm.Config

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| values-files | []string |  |  |
| version | string |  |  |

## k8sconfig.Config

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| kube-api-endpoint | [k8sconfig.*K8SEndpointConfig](#k8sconfigk8sendpointconfig) |  |  |
| audit-log | [*audit.AuditLogConfig](#auditauditlogconfig) |  |  |
| addons | [*addons.Config](#addonsconfig) |  |  |

## k8sconfig.NodeConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| node-taint | []string | NodeTaints are used to taint the node with key=value:effect. By default, server node is tainted with a couple of taints if number of agents nodes more than 0.  |  |
| node-label | []string |  |  |

## k8sconfig.K8SEndpointConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| type | string |  |  |
| firewall | [k8sconfig.*BasicFirewallConfig](#k8sconfigbasicfirewallconfig) |  |  |

## k8sconfig.BasicFirewallConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| hetzner-public | [k8sconfig.*HetnzerBasicFirewallConfig](#k8sconfighetnzerbasicfirewallconfig) |  |  |

## k8sconfig.HetnzerBasicFirewallConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| disallow-own-ip | bool |  |  |
| allowed-ips | []string |  |  |

## k3s.Config

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| version | string | version is used to determine if k3s should be upgraded if auto-upgrade is disabled. If the version is changed, k3s will be upgraded.  |  |
| clean-data-on-upgrade | bool | [Experimental] clean-data-on-upgrade is used to delete all data while upgrade. This is based on the script https://docs.k3s.io/upgrades/killall  |  |
| disable-default-taints | bool | Do not add default taints to the server node.  |  |
| config | [k3s.*K3sConfig](#k3sk3sconfig) | The real config of k3s service.  |  |

## k3s.K3sConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| token (computed). Not possible to configure! | string | Token used for nodes to join the cluster, generated automatically.  |  |
| server (computed). Not possible to configure! | string | Server specifies the address of the main server node (leader) in the cluster, generated automatically.  |  |
| flannel-iface (computed). Not possible to configure! | string | FlannelIface specifies the network interface that Flannel should use.  |  |
| write-kubeconfig-mode (computed). Not possible to configure! | string | WriteKubeconfigMode defines the file permission mode for the kubeconfig file on disk.  |  |
| advertise-address (computed). Not possible to configure! | string | AdvertiseAddr specifies the IP address that the server uses to advertise to members of the cluster.  |  |
| node-ip (computed). Not possible to configure! | string | NodeIP specifies the IP address to advertise for this node.  |  |
| bind-address (computed). Not possible to configure! | string | BindAddress is the IP address that the server should bind to for API server traffic.  |  |
| cluster-init (computed). Not possible to configure! | bool | ClusterInit indicates whether this node should initialize a new cluster.  |  |
| node-external-ip (computed). Not possible to configure! | string | ExternalNodeIP specifies the external IP address of the node.  |  |
| tls-san-security (computed). Not possible to configure! | bool | TLSSanSecurity enables or disables the addition of TLS SANs (Subject Alternative Names).  |  |
| tls-san (computed). Not possible to configure! | string | TLSSan adds specific TLS SANs for securing communication to the K3s server.  |  |
| node-name (computed). Not possible to configure! | string | NodeName specifies the name of the node within the cluster.  |  |
| cluster-cidr | string | ClusterCidr defines the IP range from which pod IPs shall be allocated.  |  |
| service-cidr | string | ServiceCidr defines the IP range from which service cluster IPs are allocated.  |  |
| cluster-domain | string | ClusterDomain specifies the domain name of the cluster.  |  |
| cluster-dns | string | ClusterDNS specifies the IP address of the DNS service within the cluster.  |  |
| flannel-backend | string | FlannelBackend determines the type of backend used for Flannel, a networking solution.  |  |
| disable-network-policy | bool | DisableNetworkPolicy determines whether to disable network policies.  |  |
| kubelet-arg | []string | KubeletArgs allows passing additional arguments to the kubelet service.  |  |
| kube-controller-manager-arg | []string | KubeControllerManagerArgs allows passing additional arguments to the Kubernetes controller manager.  |  |
| kube-cloud-controller-manager-arg | []string | KubeCloudControllerManagerArgs allows passing additional arguments to the Kubernetes cloud controller manager.  |  |
| kube-apiserver-arg | []string | KubeAPIServerArgs allows passing additional arguments to the Kubernetes API server.  |  |
| disable-cloud-controller | bool | DisableCloudController determines whether to disable the integrated cloud controller manager.  |  |
| disable | []string | Disable lists components or features to disable.  |  |


