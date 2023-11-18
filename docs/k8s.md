## Kubernetes API Server
KubeAPI Server listens on port `6443` for **all** interfaces. By default, kube api endpoint is considered as `public`. It can be changed by specifying `k8s.endpoint.type` property in the cluster configuration:
```yaml
config:
  <project>:k8s:
    kube-api-endpoint:
      type: internal
```
The following values are supported: ['public', 'internal', 'wireguard'].

For firewall rules, there is a note in #k8s-apiserver-access

It is recommended to switch to `internal` or `wireguard` mode if you want to restrict access to the apiserver from the public network after the 1st deployment of the cluster. It will remove a rule for public access entierly and change the endpoint IP address in the kubeconfig output.
For using kubeconfig with `internal` type you should have access to private network.
For `wireguard` type you can use master connection for wireguard cluster to establish a secure tunnel.

**Note**: The kubernetes pulumi provider uses the kubeconfig with **public** address of the cluster. Due the nature of how custom providers work in pulumi it is not an easy task to migrate existing resources (helm charts, manifests, etc) to another provider. So, the kube pulumi provider ignores change in endpoint type in kubeconfig right now.

**Note**: kubeconfig output is not updated automatically. You should run `pulumi up` to get updated kubeconfig output.


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
Internal networks and wireguard networks are considered as *secured*. So, no rules will be applied for them.

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
```
Most of the addons are installed using helm. So, you can specify `helm` property to override the default helm version for the addon. The default helm versions are specified in the (default-helm-versions.yaml)[../../pulumi-template/versions/default-helm-versions.yaml] file.

#### Hetzner CCM
Please note that Hetzner CCM is disabled by default. It is used to provision loadbalancers in the Hetzner cloud and other cool things. You can enabled it by setting `ccm.enabled` to `true`, but according to the [documentation](https://github.com/hetznercloud/hcloud-cloud-controller-manager/issues/80) you should recreate a cluster with enabled CCM to add the --cloud-manager=external to kubelet args.
