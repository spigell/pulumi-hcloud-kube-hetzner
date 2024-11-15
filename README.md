# THIS PROJECT IS SWITCHING TO TALOS (https://github.com/spigell/pulumi-hcloud-kube-hetzner/issues/196)

## Pulumi Hcloud Kube Hetzner
This is a [Pulumi component](https://www.pulumi.com/docs/concepts/resources/components) (only GO and Typescript/JS are supported now) for creating Kubernetes clusters in Hetzner Cloud. It is inspired by [terraform-hcloud-kube-hetzner](https://github.com/kube-hetzner/terraform-hcloud-kube-hetzner).

*Note: This project is in active development, and not everything is completed. However, it DOES work and is usable right now. I definitely appreciate feedback and will help with any issues*

### Goal
The goal of this project is to enable anybody, regardless of their "DevOps tools" expertise, to efficiently deploy and manage Kubernetes clusters on Hetzner's cost-effective cloud infrastructure. Leveraging existing features such as autotests and YAML out-of-box configuration (thanks to Pulumi), along with Hetzner's cheapest cloud offerings, the project aims to minimize operational costs and automate complex deployments. To prioritize security, additional efforts will be made to enhance security measures, ensuring the deployment and running are highly secured.

### Killer Features
- Ability to manage labels and taints directly!
- Adding and removing nodepools/nodes without changing internal IP addresses is possible.
- Most of the examples are tested via Github Actions and maintained.

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
$ pulumi new https://github.com/spigell/pulumi-hcloud-kube-hetzner/tree/main/pulumi-templates/phkh-typescript-simple
$ make microos (optional)
$ pulumi up -yf
```

## Development
### GO
*Note: It is recommended to use Pulumi local storage for development, as using the Pulumi service or self-hosted S3 storage can impact the speed of deployments.*
```
$ make test-go-project [DEV_TEMPLATE=<path/to/template-in-pulumi-template>]
$ cd test-project
```

For component building:
```
$ cd ./pulumi-component
$ make build && make install_provider # It generates all SDKs and build providers
$ export PATH=$PATH:~/go/bin
```

That's it. Now you can use all pulumi commands like `up` or `pre` with your own version of the project.

After making changes, create a PR to the `preview` branch.

## Roadmap
The roadmap is located in [roadmap.md](./docs/roadmap.md)
