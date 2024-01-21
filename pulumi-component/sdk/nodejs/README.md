## Pulumi Hcloud Kube Hetzner
This is a pulumi component plugin for creating Kubernetes clusters in Hetzner Cloud with Pulumi. It is inspired by [terraform-hcloud-kube-hetzner](https://github.com/kube-hetzner/terraform-hcloud-kube-hetzner). It can be used as a golang library as well, tho :)

### Features
- Ability to manage labels and taints!
- Most of examples are tested via Github Actions and maintained.

## Getting Started
### Prerequisites
Please install following tools:
- pulumi cli
- go (version 1.21+)
- GNU Make
- packer (only for microos image creation. If you have existing image, you can skip this step)

You need to have a Hetzner Cloud account. You can sign up for free [here](https://hetzner.com/cloud/).

### Usage
#### TL;DR (Typescript)
```
$ export HCLOUD_TOKEN=<your token>
$ make pulumi-hcloud-kube-hetzner
$ cd pulumi-hcloud-kube-hetzner
$ pulumi new -g https://github.com/spigell/pulumi-hcloud-kube-hetzner/tree/main/pulumi-template/typescript
$ make microos
$ make pulumi-generate-config
$ yarn install
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

Due the nature of non-statefull ip allocation for **internal** Hetzner network, we must ensure to keep order of all nodepools and nodes. All nodes and nodepools are sorted alphabetical in `compilation` stage. Thus, changing order in configuration file does not affect on cluster. However, adding or deleting nodepools/nodes can change order. So, when planning new cluster, please consider naming convention for nodes and nodepools. For example, you can use digit prefix like `01-control-plane-nodepool`. For deleting node, it is recomended to add property `deleted: true` for nodepool and node instead of removing them from configuration file. Remember, this only affects internal network. Public Hetzner ips are statefull and do not depend on order.

## Development
```
$ make dev-project
```


# RoadMap
## Documentation
- [ ] Add doc generation from structs
- [ ] Describe network modes
- [ ] Describe project layout
- [ ] Spelling
- [ ] Add roadmap for autoscaling

## Code
### High (pre-release)
- [x] Rewrite ssh checker
- [x] Error checking for systemctl services
- [x] Set timeouts for Command resources
- [x] Expose kubeApiServer endpoint
- [x] Expose kubeconfig
- [x] Add a external ip of the program to FW rules
- [ ] Add more validation rules (size of the net, difference between servers flags)
- [ ] Add auto upgrade management for microos
- [x] Add dynamic version detection
- [x] Add an ability to run cluster without leader tag with single master
- [x] K3s token generation
- [x] Add fw rules for the public network mode
- [ ] Add the docker workbench
- [x] Mark all sensitive values as secrets
- [ ] Add basic k8s apps (VM, hetzner MCC, upgrader, kured)

### Bugs
- [x] Fix taints for master node

### Non-high
- [ ] Add autoscaling
- [x] Add reasonable defaults for variables
- [ ] Add arm64 support
- [ ] Allow change config from code
- [ ] Package stage: reboot if changes detected only

## CI
- [x] Add linter run for every branch
- [x] Add go test run for every branch
- [x] Use pulumi cli instead of actions for up and preview. Collect logs.

## Tests
- [x] Test with multiple servers
- [x] Test with single node cluster (without leader tag)
