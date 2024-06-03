import * as hkh from "@spigell/hcloud-kube-hetzner"
import * as pulumi from '@pulumi/pulumi'

interface ClusterConfig {
    [key: string]: any;
}

const config = new pulumi.Config()
const cfg = config.requireObject<ClusterConfig>("cluster")


// Create cluster
const main = new hkh.Cluster('test', {config: cfg})

export const phkh = {
    kubeconfig: main.kubeconfig,
    servers: main.servers,
    privatekey: main.privatekey
}
