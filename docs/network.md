## Network modes
PHKH supports several types of network modes (communication between nodes in cluster):
- using only public IP (network.enabled: false);
- using private Hetzner network (network.enabled: true);

Switching between modes on the fly is supported only for `network.enabled: true -> network.enabled: false`.
Since NetworkManager is configured on the stage of cluster creation, it is not possible to reconfigure it on the fly right now. Changing `network.enabled: false -> network.enabled: true` will lead to an unstable and unreachable cluster.
You should recreate the cluster instead.

*Note*: Although the network mode can be changed on the fly, most Kubernetes clusters can't survive such change if `node-ip` for kubelet is changed. For k3s, you should recreate the cluster. The reason is that etcd stores the node ip in the cluster state and it can't be changed automatically.

## SSH access
The program creates a key pair for ssh access to the servers. The private key is stored in the pulumi state and can be retrieved by `pulumi stack output --show-secrets -j phkh | jq .privatekey` command. The reason for this is that many people use weak and unsupported keys and the program cannot move further.

In the future, there will be a possibility to add your own public key.

## Firewall
Since the firewall property belongs to `Node` structure, the Hetzner firewall can be enabled or disabled at all levels of configuration. The count of firewalls depends on several factors:
- firewalls for the role;
- firewalls for nodepool;
- firewalls for nodes;

For example, if you have 2 nodepools (control-plane, worker) with 2 nodes in each without any overrides at any level, you will have 2 firewalls: for `server` role and `agent` role.

For every override on `nodepool` and `node` level you will have an additional firewall. If there is no need for a `role` firewall (all nodepools and nodes have their own overrides) the `role` firewalls will not be created.

Also, an additional firewall will be created for internal communication between servers via a public network.

For every role (server, worker) a firewall will be created. If you want to disable the firewall for a specific node, you can set `firewall.enabled: false` for this node.

*Note*: By default, your external IP (IPv4) address is added to the firewall rules. If you want to disable this behavior, you can set `disallow-own-ip: true` for `ssh` and `kube-api-endpoint` firewall rules.

## Limitations
Right now it is not possible to create nodes without an IPv4 address. Main reasons:
- The Pulumi Kubernetes provider uses public addresses as cluster endpoint;
- There is no way to use an existing internal network yet;

