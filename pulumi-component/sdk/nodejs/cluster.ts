// *** WARNING: this file was generated by Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as inputs from "./types/input";
import * as outputs from "./types/output";
import * as utilities from "./utilities";

/**
 * Component for creating a Hetzner Cloud Kubernetes cluster.
 */
export class Cluster extends pulumi.ComponentResource {
    /** @internal */
    public static readonly __pulumiType = 'hcloud-kube-hetzner:index:Cluster';

    /**
     * Returns true if the given object is an instance of Cluster.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    public static isInstance(obj: any): obj is Cluster {
        if (obj === undefined || obj === null) {
            return false;
        }
        return obj['__pulumiType'] === Cluster.__pulumiType;
    }

    /**
     * The kubeconfig for the cluster.
     */
    public /*out*/ readonly kubeconfig!: pulumi.Output<string | undefined>;
    /**
     * The private key for nodes.
     */
    public /*out*/ readonly privatekey!: pulumi.Output<string | undefined>;
    /**
     * Information about hetnzer servers.
     */
    public /*out*/ readonly servers!: pulumi.Output<outputs.cluster.Servers[] | undefined>;

    /**
     * Create a Cluster resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args?: ClusterArgs, opts?: pulumi.ComponentResourceOptions) {
        let resourceInputs: pulumi.Inputs = {};
        opts = opts || {};
        if (!opts.id) {
            resourceInputs["config"] = args ? args.config : undefined;
            resourceInputs["kubeconfig"] = undefined /*out*/;
            resourceInputs["privatekey"] = undefined /*out*/;
            resourceInputs["servers"] = undefined /*out*/;
        } else {
            resourceInputs["kubeconfig"] = undefined /*out*/;
            resourceInputs["privatekey"] = undefined /*out*/;
            resourceInputs["servers"] = undefined /*out*/;
        }
        opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts);
        super(Cluster.__pulumiType, name, resourceInputs, opts, true /*remote*/);
    }
}

/**
 * The set of arguments for constructing a Cluster resource.
 */
export interface ClusterArgs {
    /**
     * Configuration for the cluster. 
     * Can be Struct or pulumi.Map types. 
     * Despite of the fact that SDK can accept multiple types it is recommended to use strong typep struct if possible. 
     * Caution: Not all configuration options for k3s cluster are available. 
     * Additional information can be found at https://github.com/spigell/pulumi-hcloud-kube-hetzner/blob/main/docs/parameters.md
     */
    config?: pulumi.Input<inputs.cluster.ConfigConfigArgs | {[key: string]: pulumi.Input<string>}>;
}
