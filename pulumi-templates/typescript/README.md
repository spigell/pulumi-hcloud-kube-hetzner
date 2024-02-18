## Pulumi Hcloud Kube Hetzner

## Usage
### Prerequisites
Please install following tools:
- Pulumi CLI and required runtime for your language
- GNU Make
- packer (only for microos image creation. If you have existing image, you can skip this step)
- curl

It is recomended to export env variable `HCLOUD_TOKEN` since it is required for large amount of commands
However, you can provide it every time when you requested it

### TL;DR
```
$ export HCLOUD_TOKEN=<your token>
$ make microos
$ make pulumi-init-stack
$ yarn install (if typescript is the runtime)
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

#### 3. Create pulumi stack and generate configuration for it
```
make pulumi-init-stack [PULUMI_EXAMPLE_NAME=<name of the file in /examples directory without .yaml suffix>]
```
#### 4. Install dependencies (if required)
```
yarn instal
```
**That's it! Now you can use pulumi commands like `up` or `preview`.**

### Outputs
The program sends outputs via map called `phkh`. The one can get outputs using command `pulumi stack output --show-secrets -j pkhk`.

The YAML state file will be created as well. It is used by the program for internal purposes. If you use some VCS, like git, you should store this file along with your configuration.

### Configuration
Configuration can be made via editing Pulumi.<stack>.yaml file.

All valid conbinations between defauls/nodepools/nodes are considered to be supported, but some changes require cluster recreation (cluster recreation means `pulumi destroy` and `pulumi up`).
If you find any panic (due accessing to a null value or like that), please create an issue!

### Useful commands and snippets
#### Get ssh private key
```
pulumi stack output --show-secrets -j pkhk | jq .privatekey -r
```
#### Check ssh connectivity to nodes from local machine
```
make pulumi-ssh-check
```
#### SSH to node with make
```
make pulumi-ssh-to-node TARGET=<ID of node>
```


## Development
This directory contains files for bootstrap pulumi project.

It used by `Makefile` in the root of the repository as well

Includes:
- A packer template for microos image creation in Hetzner Cloud;
- A pulumi stack template for creating a cluster with required files;

*Note*: Please do not change `${PROJECT}` in `package.json`, `go.mod`, and `Pulumi.yaml` files if changes are required there.
