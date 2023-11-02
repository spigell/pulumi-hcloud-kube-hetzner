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
## Code
- [ ] Rewrite wireguard part
- [ ] Rewrite ssh checker
- [ ] Add reasonable defaults for variables
- [ ] K3s token generation
- [ ] Add arm64 support

## Tests
- [ ] Add idempotent tests for all runs
- [ ] Add tests for wireguard run (check master connection)
- [ ] Test with multiple servers
- [ ] Test with single node cluster (without leader tag)
