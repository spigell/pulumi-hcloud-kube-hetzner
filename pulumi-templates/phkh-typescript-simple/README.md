## Pulumi Hcloud Kube Hetzner

## Usage
### Prerequisites
Please install following tools:
- Pulumi CLI and required runtime for your language
- GNU Make
- packer (only for microos image creation. If you have existing image, you can skip this step)
- curl

It is recomended to export env variable `HCLOUD_TOKEN` since it is required for large amount of commands
However, you can provide it every time when you requested it.

### TL;DR
```
$ export HCLOUD_TOKEN=<your token>
$ make microos
$ pulumi pre -yf
$ pulumi up -yf
```

### More detailed
#### 1. Get your token
You need to have a Hetzner Cloud account. You can sign up for free [here](https://hetzner.com/cloud/).

You can get your token on the project security page. Please see [here] (https://docs.hetzner.com/cloud/api/getting-started/generating-api-token/)

#### 2. Create microos image
```
make microos
```
It will create microos snapshot with name `microos-amd64-<timestamp>`. It uses packer and hcloud plugin for it.

*Note*: right now only x86 architecture is supported. If you need arm64, please create an issue.

**That's it! Now you can use pulumi commands like `up` or `preview`.**

### Outputs
The program sends outputs via map called `phkh`. The one can get outputs using command `pulumi stack output --show-secrets -j phkh`.

The YAML state file will be created as well in `states` directory. It is used by the program for internal purposes. If you use some VCS, like git, you should store this file along with your configuration.

### Configuration
Configuration can be made via `config` key in the `Cluster`.

All valid conbinations between defauls/nodepools/nodes are considered to be supported, but some changes require cluster recreation (cluster recreation means `pulumi destroy` and `pulumi up`).
If you find any panic (due accessing to a null value or like that), please create an issue!

### Useful commands and snippets
#### Get ssh private key
```
pulumi stack output --show-secrets -j phkh | jq .<name-of-cluster>.privatekey -r
```
#### Get kubeconfig
```
pulumi stack output --show-secrets -j phkh | jq .<name-of-cluster>.kubeconfig -r
```
