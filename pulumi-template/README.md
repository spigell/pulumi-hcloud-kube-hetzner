# PHKH Pulumi Template
This directory contains files for bootstrap pulumi project.

It used by `Makefile` in the root of the repository.

Includes:
- A packer template for microos image creation in Hetzner Cloud;
- main.go with example of usage the library;
- scripts and tools;

*Note*: Please do not change `${PROJECT}` variable in go.mod and `Pulumi.yaml` files if changes are required.
