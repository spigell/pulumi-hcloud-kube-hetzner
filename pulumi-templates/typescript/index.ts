import * as hkh from "@spigell/hcloud-kube-hetzner"
import * as pulumi from '@pulumi/pulumi'

interface ClusterConfig {
    [key: string]: any;
}

const config = new pulumi.Config()
const clusters = config.requireObject<ClusterConfig>("clusters")

// This is an example for changing values for specific cluster configuration.
const clusterName = 'main'
const cfg = clusters[clusterName]

// Disabling internal network
// cfg['network']['hetzner']['enabled'] = false

// Create cluster
const main = new hkh.Cluster(clusterName, {configuration: cfg})

export const phkh = {
    [clusterName]: {
        kubeconfig: main.kubeconfig,
        servers: main.servers,
        privatekey: main.privatekey
    }
}

