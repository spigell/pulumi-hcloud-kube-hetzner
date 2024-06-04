## Pulumi-templates

Each directory here is a Pulumi template and can be used via the pulumi new command. The naming schema of the templates is `phkh-<language>-<type of configuration>`.

The types of configuration are:

- simple: Creates a very simple one-node cluster with defaults.
- cluster-files: All cluster configurations are stored in the clusters directory. There is a directory called cluster-examples with YAML files that can be dropped into the clusters directory. Please follow the documentation for parameters.
- pulumi-config: Used for testing purposes. It can create only one cluster.

### Development

This directory contains files for bootstrapping a Pulumi project. It is used by the Makefile in the root of the repository for syncing and managing dependencies:

- make sync-templates: Syncs all similar files between all directories. Executed on every MR push. Source is go-simple project.
- make up-template-versions: Updates all dependencies, including this project. Must be used after releasing a new version.

*Note*: Please do not change ${PROJECT} in package.json, go.mod, and Pulumi.yaml files if changes are required there.