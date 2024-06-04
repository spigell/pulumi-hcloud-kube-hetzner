import * as hkh from "@spigell/hcloud-kube-hetzner"

const clusterName = 'simple'
// Create cluster
const cluster = new hkh.Cluster(clusterName, {config: {
    Nodepools: {
        Servers: [
        {
            PoolID: 'servers', 
            Nodes: [
                {NodeID: 'server-01'}
            ]
        }]
    }
}})

export const phkh = {
    [clusterName]: {
        kubeconfig: cluster.kubeconfig,
        servers: cluster.servers,
        privatekey: cluster.privatekey
    }
}
