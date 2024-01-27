## Pulumi Hcloud Kube Hetzner

## Usage
It is recomended to export env variable `HCLOUD_TOKEN` since it is required for large amount of commands
However, you can provide it every time when you requested it

TL;DR
```
$ export HCLOUD_TOKEN=<your token>
$ make microos
$ make pulumi-init-stack
$ yarn install (if typescript is the runtime)
$ pulumi up -yf
```

### More detailed
#### 1. Get your token
You can get your token [here](https://console.hetzner.cloud/projects)

#### 2. Create microos image
```
make microos
```
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

### Useful commands and snippets
#### Get ssh private key
```
pulumi stack output --show-secrets -j --path 'pkhk.privatekey' | jq . -r
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
