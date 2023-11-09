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


### K8S APIServer access
By default, a hetzner firewall rule allows all traffic to 6443 port if `k8s.endpoint.type` specified as `public` (this is a default value). If you want to restrict access to the apiserver from the public network, you can use the following configuration:
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
