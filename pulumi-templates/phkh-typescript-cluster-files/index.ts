import * as hkh from "@spigell/hcloud-kube-hetzner"
import * as pulumi from "@pulumi/pulumi"
import * as yaml from 'yaml'
import * as fs from 'fs'
import * as path from 'path'

const dir = './clusters'
const files = fs.readdirSync(dir);
const yamlFiles = files.filter(file => file.endsWith('.yaml') || file.endsWith('.yml'));

export const outputs = new Map<string, any>()

yamlFiles.forEach(file => {
    const name = path.parse(file).name;
    const filePath = path.join(dir, file);
    const content = fs.readFileSync(filePath, 'utf8');
    const parsedContent = yaml.parse(content, {merge: true});

    const cluster = new hkh.Cluster(name, {config: pulumi.output(parsedContent)})

    outputs.set(name, {kubeconfig: cluster.kubeconfig})
});
