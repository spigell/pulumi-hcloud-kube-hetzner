## Pulumi Hcloud Kube Hetzner
This is a [Pulumi component](https://www.pulumi.com/docs/concepts/resources/components) (only GO and Typescript/JS are supported now) for creating Kubernetes clusters in Hetzner Cloud. It is inspired by [terraform-hcloud-kube-hetzner](https://github.com/kube-hetzner/terraform-hcloud-kube-hetzner). It can be used as a golang library/module as well, tho :)

*Note: This project is in active development, and not everything is completed. However, it DOES work and is usable right now. I definitely appreciate feedback and will help with any issues*

### Features
- Ability to manage labels and taints!
- Adding and removing nodepools/nodes without changing internal IP addresses is possible.
- Most of examples are tested via Github Actions and maintained.

## Getting Started
### Prerequisites
Please install following tools:
- Pulumi CLI and required runtime for your language
- GNU Make
- packer (only for microos image creation. If you have existing image, you can skip this step)
- curl

You need to have a Hetzner Cloud account. You can sign up for free [here](https://hetzner.com/cloud/).

### Usage
#### TL;DR (Typescript)
```
$ export HCLOUD_TOKEN=<your token>
$ mkdir pulumi-hcloud-kube-hetzner
$ cd pulumi-hcloud-kube-hetzner
$ pulumi new -g https://github.com/spigell/pulumi-hcloud-kube-hetzner/tree/main/pulumi-templates/typescript
$ make microos
$ make pulumi-init-stack
$ yarn install
$ pulumi up -yf
```

## Development
### GO
```
$ make test-go-project [TEMPLATE=go/library|go/component]
$ cd test-component
$ make pulumi-generate-config [PULUMI_CONFIG_SOURCE=../examples/<EXAMPLE>.yaml]
```

For component building:
```
$ cd ./pulumi-component
$ make build && make install_provider # It generates all SDKs and build providers
$ export PATH=$PATH:~/go/bin
```

That it. Now you can use all pulumi command like `up` or `pre` with own version of the project.

After changes create a PR to the `preview` branch.

## Roadmap
The roadmap is located in [roadmap.md](./docs/roadmap.md)
