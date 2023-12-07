## K3S
### Version management
The program allows to manage k3s versions both of methods:
- exact version specified in the configuration;
- using [system-upgrade-controller](https://github.com/rancher/system-upgrade-controller) and crd `Plan` (version and channel are available);

Also the k3s well-known label `k3s-upgrade` will be added if system-upgrade-controller is enabled. But the user can disable it setting `k3s-upgrade=false` in k3s label section.

The following table describes error combinations:

| version (manual) |upgrader enabled (version or channel)|  `k3s-upgrade=false` label  | Error type                   |
|:----------------:|:------------------------------------|-----------------------------|------------------------------|
|         x        |                 x                   |                             | Either manual or auto        |
|                  |                 x                   |             x               | Version must be set manually |
