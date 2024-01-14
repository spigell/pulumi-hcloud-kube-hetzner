# PHKH Pulumi Template
This directory contains files for bootstrap pulumi project.

It used by `Makefile` in the root of the repository.

Includes:
- A packer template for microos image creation in Hetzner Cloud;
- main.go with example of usage the library;
- scripts and tools;

*Note*: Please do not change `${PROJECT}` variable in go.mod and `Pulumi.yaml` files if changes are required.

## Usage
#### Useful commands and snippets
### Get ssh keys
```
pulumi stack output --show-secrets   -j ssh:keypair | jq .PrivateKey -r
```

### Get wg master key
```
pulumi stack output --show-secrets   wireguard:connection > ~/wg-dev.conf && wg-quick down ~/wg-dev.conf ; wg-quick up ~/wg-dev.conf
```
