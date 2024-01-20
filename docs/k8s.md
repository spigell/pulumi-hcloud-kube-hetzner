## Kubernetes API Server
KubeAPI Server listens on port `6443` for **all** interfaces. By default, kube api endpoint is considered as `public`. It can be changed by specifying `k8s.endpoint.type` property in the cluster configuration:
```yaml
config:
  <project>:k8s:
    kube-api-endpoint:
      type: internal
```
The following values are supported: ['public', 'internal']

For firewall rules, there is a note in #k8s-apiserver-access

It is recommended to switch to `internal` or mode if you want to restrict access to the apiserver from the public network after the 1st deployment of the cluster. It will remove a rule for public access entierly.
For using kubeconfig with `internal` type you should have access to private network.

**Note**: The kubernetes pulumi provider uses the kubeconfig with **public** address of the cluster. Due the nature of how custom providers work in pulumi it is not an easy task to migrate existing resources (helm charts, manifests, etc) to another provider. Thus, the kube pulumi provider ignores change in endpoint type in kubeconfig right now.


### K8S APIServer external access
By default, a hetzner firewall rule allows all traffic to **6443** port if `k8s.endpoint.type` is specified as `public` (this is a default value). If you want to restrict access to the apiserver from the public network, you can use the following configuration:
```yaml
config:
  <project>:k8s:
    endpoint:
      type: public
      firewall:
        # This only works for the public endpoint.
        hetzner-public:
          allowed-ips:
            - '102.0.0.0/8' # <--- Allow access to the k8s api from the this cidr!
```
Internal network networks are considered as *secured*. So, no rules will be applied for them.

## Node Management
### Node Labels and Taints and K3S
Despike the fact that the labels and taints are using only at registration stage the program allows change them after the registration. It is done by cluster-manager that use nodePatch ServerSide Apply to manage labels and taints on the nodes after bootstraping.

## Addons
Most of the addons are installed using helm. So, you can specify `helm` property to configure some values:

- `version`: The version of helm chart. The default helm versions are specified in the (default-helm-versions.yaml)[../../pulumi-template/versions/default-helm-versions.yaml] file.
- `values-files`: A list of values files to be used with helm chart. It can be used to override unmanaged settings. Not all addons support this feature.

### Addons
Additional components can be installed to the cluster using `addons` property:
```yaml
config:
  <project>:k8s:
    addons:
      ccm:
        enabled: true
        default-loadbalancers-location: fsn1
        loadbalancers-enabled: true
        helm:
          version: v1.2.0
          values-files:
            - ./yaml/ccm/values.yaml
```

#### Hetzner CCM
Please note that Hetzner CCM is disabled by default. It is used to provision loadbalancers in the Hetzner cloud and other cool things. You can enabled it by setting `ccm.enabled` to `true`, but according to the [documentation](https://github.com/hetznercloud/hcloud-cloud-controller-manager/issues/80) you should recreate a cluster with enabled CCM to add the --cloud-manager=external to kubelet args.

#### K3S Upgrade Controller
K3S upgrade controller is used to upgrade k3s cluster to the specified `target-version` and/or `target-channel`. It is disabled by default and utilize the [system-upgrade-controller chart by nimbolus](https://github.com/nimbolus/helm-charts/blob/main/charts/system-upgrade-controller). It doesn't support `values-files` property. But settings of the upgrader can be configured using `config-env` property:
```yaml
config:
  <project>:k8s:
    addons:
      k3s-upgrade-controller:
        enabled: true
        target-channel: v1.28
        config-env:
          - "SYSTEM_UPGRADE_CONTROLLER_DEBUG=false"
```
Please see all available variables in the [chart default values](https://github.com/nimbolus/helm-charts/blob/main/charts/system-upgrade-controller/values.yaml).
