## Limitations
Right now it is not possible to create nodes without ipv4 and ipv6 public addresses.

## Network modes
PHKH supports several types of network modes (communication between nodes in cluster):
- using only public IP (network.enabled: false);
- using private hetnzer network (network.enabled: true);
- using wireguard built on top of public ips (wireguard.enabled: true and network.enabled: false);
- using wireguard based on private ips (wireguard.enabled: true and network.enabled: true);

Switching between modes on the fly is supported except switching from scenarios where `network.enabled: false -> network.enabled: true`.
Since NetworkManager is configured on the stage of cluster creation, it is not possible to reconfigure it on the fly right now. Changing `network.enabled: false -> network.enabled: true` will lead to an unstable and unreachable cluster.
You should recreate the cluster instead.

## Kubernetes API Server external access
KubeAPI Server listens on port `6443`. By default, kube api endpoint considered as `public`. It can be changed by specifying `k8s.endpoint.type` property in the cluster configuration:
```yaml
config:
  <project>:k8s:
    kube-api-endpoint:
      type: internal
```
The following values are supported: ['public', 'internal', 'wireguard'].

With `public` type, the apiserver will be accessible from the public network. You can restric access to the apiserver from the public network by specifying the following configuration:
```yaml
config:
  <project>:k8s:
    kube-api-endpoint:
      type: public
      firewall:
        hetzner-public: # <-- This only works for the public endpoint.
          allowed-ips:  # <-- Allow access to the k8s api from the this cidr!
            - '102.0.0.0/0'
```

It is recommended to switch to `internal` or `wireguard` mode if you want to restrict access to the apiserver from the public network after the 1st deploy of the cluster. It will remove a rule for public access entierly and change endpoint IP address in kubeconfig output.

For `internal` type you should have access to private network. 
For `wireguard` type you can use master connection for wireguard cluster.


## Firewall
Since the firewall property belongs to `Node` structure, the hetzner firewall can be enabled or disabled on all layers of configuration. The count of firewalls depends on several factors:
- firewalls for the role;
- firewalls for nodepool;
- firewalls for nodes;

For example, if you have 2 nodepools (control-plane, worker) with 2 nodes in each without any overrides on any level, you will have 2 firewalls: for `server` role and `agent` role.

For every override on `nodepool` and `node` level you will have additional firewall. If there is no need for `role` firewall (all nodepools and nodes have their own overrides) the `role` firewalls will not be created.

Also, the additional firewall will be created for internal communication between servers via a public network.

For every role (server, worker) firewall will be created If you want to disable the firewall for a specific node, you can set `firewall.enabled: false` for this node.

### Wireguard
By default, a hetzner firewall rule is added to allow all traffic to **51822** port for every traffic if wireguard enabled for in-cluster communication method. Restriction can be applied by specifying the following configuration:
```yaml
<project>:network:
    wireguard:
      enabled: true
      firewall:
        hetzner:
          allowed-ips:
            - '102.0.0.0/8'
```

### K8S APIServer access
By default, a hetzner firewall rule allows all traffic to **6443** port if `k8s.endpoint.type` specified as `public` (this is a default value). If you want to restrict access to the apiserver from the public network, you can use the following configuration:
```yaml
    endpoint:
      type: public
      firewall:
        # This only works for the public endpoint.
        hetzner-public:
          allowed-ips:
            - '102.0.0.0/8' # <--- Allow access to the k8s api from the this cidr!
```
Internal networks and wireguard networks are considered as *secured*. So, no rules will be applied for them.
