## Pulumi Hcloud Kube Hetzner
This project is a golang library for creating Kubernetes clusters in Hetzner Cloud with Pulumi. It is inspired by [terraform-hcloud-kube-hetzner](https://github.com/kube-hetzner/terraform-hcloud-kube-hetzner). It is available only for go projects since there is no `component` for such things in pulumi.

## Getting Started
### Prerequisites
Please install following tools:
- pulumi cli
- go (version 1.21+)
- GNU Make
- packer (only for microos image creation. If you have existed image, you can skip this step)

You need to have a Hetzner Cloud account. You can sign up for free [here](https://hetzner.com/cloud/).

### Usage
#### TL;DR
```
$ export HCLOUD_TOKEN=<your token>
$ pulumi new -g https://github.com/spigell/pulumi-hcloud-kube-hetzner/tree/main/pulumi-template pulumi-hcloud-kube-hetzner
$ cd pulumi-hcloud-kube-hetzner
$ make microos
$ make pulumi-config
$ pulumi up -yf
```

#### Step by step
It is recomended to export env variable `HCLOUD_TOKEN` since it is required for large amount of commands

However, you can provide it every time when you requested it

### Create microos image
```
make microos
```
*Note*: right now only x86 architecture is supported. If you need arm64, please create an issue.

### Create pulumi stack and generate configuration for it
```
make pulumi-config PULUMI_CONFIG_SOURCE=/path/to/file
```
`PULUMI_CONFIG_SOURCE` is the path for config. It can be any yaml config file. You can browse examples in example directory. Most of this examples are tested via Actions and considered as supported.

## Supported scenarios
All valid conbinations between defauls{agents/servers}/nodepools.config/nodes are considered to be supported and changeable on the fly without cluster recreation (cluster recreation means `pulumi destroy` and `pulumi up`).
If you find any panic (due accessing to a null value or like that), please create an issue!

### Nodepools and Nodes
Adding or Deleting nodepools/nodes are supported with several limitation.

Due the nature of non-statefull ip allocation for **internal** Hetzner network, we must ensure to keep order of all nodepools and nodes. All nodes and nodepools are sorted alphabetical in `compilation` stage. Thus, changing order in configuration file does not affect on cluster. However, adding or deleting nodepools/nodes can change order. So, when planning new cluster, please consider naming convention for nodes and nodepools. For example, you can use digit prefix like `01-control-plane-nodepool`. For deleting node, it is recomended to add property `deleted: true` for nodepool and node instead of removing them from configuration file. Remember, this only affects internal network. Wireguard network and public Hetzner ips are statefull and do not depend on order.

### Network changes
PHKH supports several types of communication between nodes of cluster:
- using only public ip (network.enabled: false);
- using private hetnzer network (network.enabled: true);
- using wireguard built on top of public ips (wireguard.enabled: true and network.enabled: false);
- using wireguard on private ips (wireguard.enabled: true and network.enabled: true);

Switching between modes on the fly is supported except switching from scenarios where `network.enabled: false -> network.enabled: true`.
Since NetworkManager configured on stage of cluster creation, it is not possible to switch between these scenarios. You should recreate cluster.

#### Useful commands and snippets
### Get ssh keys
```
pulumi stack output --show-secrets   -j ssh:keypair | jq .PrivateKey -r
```

### Get wg master key
```
pulumi stack output --show-secrets   wireguard:connection > ~/wg-dev.conf && wg-quick down ~/wg-dev.conf ; wg-quick up ~/wg-dev.conf
```

## Development
```
$ make test-project
```


# RoadMap
## Documentation
- [ ] Add doc generation from structs
- [ ] Describe network modes
- [ ] Spelling

## Code
### High (pre-release)
- [x] Rewrite ssh checker
- [x] Error checking for systemctl services
- [x] Set timeouts for Command resources
- [ ] Add more validation rules
- [ ] K3s token generation
- [ ] Add fw rules for the public network mode
- [ ] Add basic k8s apps (VM, metrics-server, etc, hetzner MCC, upgrader, kured)

### Bugs
- [x] Fix taints for master node
- [ ] Use external ip for master wireguard connection always.

### Non-high
- [ ] Rewrite wireguard stage
- [x] Add reasonable defaults for variables
- [ ] Add arm64 support
- [ ] Allow change config from code
- [ ] Implement non-parallel provisioning (useful while upgrading in manual mode). All nodes waits for leader now.
- [ ] Package stage: reboot if changes detected only
- [x] Restart k3s if wireguard restarted (!)

## CI
- [ ] Add linter run for every PR
- [ ] Add go test run for every PR
- [ ] Use pulumi cli instead of actions for up and preview. Collect logs.

## Tests
- [ ] Add idempotent tests for all runs
- [ ] Add tests for wireguard run (check master connection)
- [ ] Test with multiple servers
- [ ] Test with single node cluster (without leader tag)
