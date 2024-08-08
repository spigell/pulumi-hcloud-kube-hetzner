// *** WARNING: this file was generated by Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as inputs from "../types/input";
import * as outputs from "../types/output";

export namespace cluster {
    export interface AddonsConfigArgs {
        /**
         * CCM defines configuration [hetzner-cloud-controller-manager](https://github.com/hetznercloud/hcloud-cloud-controller-manager). 
         */
        CCM?: pulumi.Input<inputs.cluster.CcmConfigArgs>;
        /**
         * K3SSystemUpgrader defines configuration for [system-upgrade-controller](https://github.com/rancher/system-upgrade-controller). 
         */
        K3SSystemUpgrader?: pulumi.Input<inputs.cluster.K3supgraderConfigArgs>;
    }

    export interface AuditAuditLogConfigArgs {
        /**
         * AuditLogMaxAge defines the maximum number of days to retain old audit log files. Default is 10. 
         */
        AuditLogMaxAge?: pulumi.Input<number>;
        /**
         * AuditLogMaxBackup specifies the maximum number of audit log files to retain. Default is 30. 
         */
        AuditLogMaxBackup?: pulumi.Input<number>;
        /**
         * AuditLogMaxSize specifies the maximum size in megabytes of the audit log file before it gets rotated. Default is 100. 
         */
        AuditLogMaxSize?: pulumi.Input<number>;
        /**
         * Enabled specifies if the audit log is enabled. If nil, it might default to a cluster-level setting. Default is true. 
         */
        Enabled?: pulumi.Input<boolean>;
        /**
         * PolicyFilePath is the path to the local file that defines the audit policy configuration. 
         */
        PolicyFilePath?: pulumi.Input<string>;
    }

    export interface CcmConfigArgs {
        /**
         * Enabled is a flag to enable or disable hcloud CCM. 
         */
        Enabled?: pulumi.Input<boolean>;
        Helm?: pulumi.Input<inputs.cluster.HelmConfigArgs>;
        /**
         * DefaultloadbalancerLocation is a default location for the loadbancers. 
         */
        LoadbalancersDefaultLocation?: pulumi.Input<string>;
        /**
         * LoadbalancersEnabled is a flag to enable or disable loadbalancers management. Note: internal loadbalancer for k3s will be disabled. 
         */
        LoadbalancersEnabled?: pulumi.Input<boolean>;
        /**
         * Token is a hcloud token to access hcloud API for CCM. 
         */
        Token?: pulumi.Input<string>;
    }

    export interface ConfigConfigArgs {
        /**
         * Defaults is a map with default settings for agents and servers. Global values for all nodes can be set here as well. Default is not specified. 
         */
        Defaults?: pulumi.Input<inputs.cluster.ConfigDefaultConfigArgs>;
        /**
         * K8S defines a distribution-agnostic cluster configuration. Default is not specified. 
         */
        K8S?: pulumi.Input<inputs.cluster.K8sconfigConfigArgs>;
        /**
         * Network defines network configuration for cluster. Default is not specified. 
         */
        Network?: pulumi.Input<inputs.cluster.ConfigNetworkConfigArgs>;
        /**
         * Nodepools is a map with agents and servers defined. Required for at least one server node. Default is not specified. 
         */
        Nodepools?: pulumi.Input<inputs.cluster.ConfigNodepoolsConfigArgs>;
    }

    export interface ConfigDefaultConfigArgs {
        /**
         * Agents holds configuration settings specific to agent nodes, overriding Global settings where specified. 
         */
        Agents?: pulumi.Input<inputs.cluster.ConfigNodeConfigArgs>;
        /**
         * Global provides configuration settings that are applied to all nodes, unless overridden by specific roles. 
         */
        Global?: pulumi.Input<inputs.cluster.ConfigNodeConfigArgs>;
        /**
         * Servers holds configuration settings specific to server nodes, overriding Global settings where specified. 
         */
        Servers?: pulumi.Input<inputs.cluster.ConfigNodeConfigArgs>;
    }

    export interface ConfigFirewallConfigArgs {
        /**
         * Hetzner specify firewall configuration for cloud firewall. 
         */
        Hetzner?: pulumi.Input<inputs.cluster.FirewallConfigArgs>;
    }

    export interface ConfigNetworkConfigArgs {
        /**
         * Hetzner specifies network configuration for private networking. 
         */
        Hetzner?: pulumi.Input<inputs.cluster.NetworkConfigArgs>;
    }

    export interface ConfigNodeConfigArgs {
        /**
         * K3S is the configuration of a k3s cluster. 
         */
        K3s?: pulumi.Input<inputs.cluster.K3sConfigArgs>;
        /**
         * K8S is common configuration for nodes. 
         */
        K8S?: pulumi.Input<inputs.cluster.K8sconfigNodeConfigArgs>;
        /**
         * Leader specifies the leader of a multi-master cluster. Required if the number of masters is more than 1. Default is not specified. 
         */
        Leader?: pulumi.Input<boolean>;
        /**
         * NodeID is the id of a server. It is used throughout the entire program as a key. Required. Default is not specified. 
         */
        NodeID?: pulumi.Input<string>;
        /**
         * OS defines configuration for operating system. 
         */
        OS?: pulumi.Input<inputs.cluster.OsconfigOSConfigArgs>;
        /**
         * Server is the configuration of a Hetzner server. 
         */
        Server?: pulumi.Input<inputs.cluster.ConfigServerConfigArgs>;
    }

    export interface ConfigNodepoolConfigArgs {
        /**
         * Config is the default node configuration for the group. 
         */
        Config?: pulumi.Input<inputs.cluster.ConfigNodeConfigArgs>;
        /**
         * Nodes is a list of nodes inside of the group. 
         */
        Nodes?: pulumi.Input<pulumi.Input<inputs.cluster.ConfigNodeConfigArgs>[]>;
        /**
         * PoolID is id of group of servers. It is used through the entire program as key for the group. Required. Default is not specified. 
         */
        PoolID?: pulumi.Input<string>;
    }

    export interface ConfigNodepoolsConfigArgs {
        /**
         * Agents is a list of NodepoolConfig objects, each representing a configuration for a pool of agent nodes. 
         */
        Agents?: pulumi.Input<pulumi.Input<inputs.cluster.ConfigNodepoolConfigArgs>[]>;
        /**
         * Servers is a list of NodepoolConfig objects, each representing a configuration for a pool of server nodes. 
         */
        Servers?: pulumi.Input<pulumi.Input<inputs.cluster.ConfigNodepoolConfigArgs>[]>;
    }

    export interface ConfigServerConfigArgs {
        /**
         * AdditionalSSHKeys contains a list of additional public SSH keys to install in the server's user account. 
         */
        AdditionalSSHKeys?: pulumi.Input<pulumi.Input<string>[]>;
        /**
         * Firewall points to an optional configuration for a firewall to be associated with the server. 
         */
        Firewall?: pulumi.Input<inputs.cluster.ConfigFirewallConfigArgs>;
        /**
         * Hostname is the desired hostname to assign to the server. Default is `phkh-${name-of-stack}-${name-of-cluster}-${id-of-node}`. 
         */
        Hostname?: pulumi.Input<string>;
        /**
         * Image specifies the operating system image to use for the server (e.g., "ubuntu-20.04" or id of private image). Default is autodiscovered. 
         */
        Image?: pulumi.Input<string>;
        /**
         * Location specifies the physical location or data center where the server will be hosted (e.g., "fsn1"). Default is hel1. 
         */
        Location?: pulumi.Input<string>;
        /**
         * ServerType specifies the type of server to be provisioned (e.g., "cx11", "cx21"). Default is cx21. 
         */
        ServerType?: pulumi.Input<string>;
        /**
         * UserName is the primary user account name that will be created on the server. Default is rancher. 
         */
        UserName?: pulumi.Input<string>;
        /**
         * UserPasswd is the password for the primary user account on the server. 
         */
        UserPasswd?: pulumi.Input<string>;
    }

    export interface FirewallConfigArgs {
        /**
         * AdditionalRules is a list of additional rules to be applied. 
         */
        AdditionalRules?: pulumi.Input<pulumi.Input<inputs.cluster.FirewallRuleConfigArgs>[]>;
        /**
         * AllowICMP indicates whether ICMP traffic is allowed. Default is false. 
         */
        AllowICMP?: pulumi.Input<boolean>;
        /**
         * Enabled specifies if the configuration is active. Default is false. 
         */
        Enabled?: pulumi.Input<boolean>;
        /**
         * SSH holds the SSH specific configurations. 
         */
        SSH?: pulumi.Input<inputs.cluster.FirewallSSHConfigArgs>;
    }

    export interface FirewallRuleConfigArgs {
        /**
         * Description provides a human-readable explanation of what the rule is intended to do. 
         */
        Description?: pulumi.Input<string>;
        /**
         * Port specifies the network port number or range applicable for the rule. Required. 
         */
        Port?: pulumi.Input<string>;
        /**
         * Protocol specifies the network protocol (e.g., TCP, UDP) applicable for the rule. Default is TCP. 
         */
        Protocol?: pulumi.Input<string>;
        /**
         * SourceIps lists IP addresses or subnets from which traffic is allowed or to which traffic is directed, based on the Direction. Required. 
         */
        SourceIps?: pulumi.Input<pulumi.Input<string>[]>;
    }

    export interface FirewallSSHConfigArgs {
        /**
         * Allow indicates whether SSH access is permitted. Default is false. 
         */
        Allow?: pulumi.Input<boolean>;
        /**
         * AllowedIps lists specific IP addresses that are permitted to access via SSH. 
         */
        AllowedIps?: pulumi.Input<pulumi.Input<string>[]>;
        /**
         * DisallowOwnIP specifies whether SSH access from the deployer's own IP address is disallowed. Default is false. 
         */
        DisallowOwnIP?: pulumi.Input<boolean>;
    }

    export interface HelmConfigArgs {
        /**
         * ValuesFilePaths is a list of path/to/file to values files. See https://www.pulumi.com/registry/packages/kubernetes/api-docs/helm/v3/release/#valueyamlfiles_nodejs for details. 
         */
        ValuesFilePath?: pulumi.Input<pulumi.Input<string>[]>;
        /**
         * Version is version of helm chart. Default is taken from default-helm-versions.yaml in template's versions directory. 
         */
        Version?: pulumi.Input<string>;
    }

    export interface JournaldConfigArgs {
        /**
         * GatherAuditD indicates whether auditd logs should be gathered. Default is true. 
         */
        GatherAuditD?: pulumi.Input<boolean>;
        /**
         * GatherToLeader indicates whether journald logs should be sent to the leader node. Default is true. 
         */
        GatherToLeader?: pulumi.Input<boolean>;
    }

    export interface K3sConfigArgs {
        /**
         * [Experimental] clean-data-on-upgrade is used to delete all data while upgrade. This is based on the script https://docs.k3s.io/upgrades/killall 
         */
        CleanDataOnUpgrade?: pulumi.Input<boolean>;
        /**
         * The real config of k3s service. 
         */
        K3S?: pulumi.Input<inputs.cluster.K3sK3sConfigArgs>;
        /**
         * Version is used to determine if k3s should be upgraded if auto-upgrade is disabled. If the version is changed, k3s will be upgraded. 
         */
        Version?: pulumi.Input<string>;
    }

    export interface K3sK3sConfigArgs {
        /**
         * ClusterCidr defines the IP range from which pod IPs shall be allocated. Default is 10.141.0.0/16. 
         */
        ClusterCidr?: pulumi.Input<string>;
        /**
         * ClusterDNS specifies the IP address of the DNS service within the cluster. Default is autopicked. 
         */
        ClusterDNS?: pulumi.Input<string>;
        /**
         * ClusterDomain specifies the domain name of the cluster. 
         */
        ClusterDomain?: pulumi.Input<string>;
        /**
         * Disable lists components or features to disable. 
         */
        Disable?: pulumi.Input<pulumi.Input<string>[]>;
        /**
         * DisableCloudController determines whether to disable the integrated cloud controller manager. Default is false, but will be true if ccm is enabled. 
         */
        DisableCloudController?: pulumi.Input<boolean>;
        /**
         * DisableNetworkPolicy determines whether to disable network policies. 
         */
        DisableNetworkPolicy?: pulumi.Input<boolean>;
        /**
         * FlannelBackend determines the type of backend used for Flannel, a networking solution. 
         */
        FlannelBackend?: pulumi.Input<string>;
        /**
         * KubeAPIServerArgs allows passing additional arguments to the Kubernetes API server. 
         */
        KubeAPIServerArgs?: pulumi.Input<pulumi.Input<string>[]>;
        /**
         * KubeCloudControllerManagerArgs allows passing additional arguments to the Kubernetes cloud controller manager. 
         */
        KubeCloudControllerManagerArgs?: pulumi.Input<pulumi.Input<string>[]>;
        /**
         * KubeControllerManagerArgs allows passing additional arguments to the Kubernetes controller manager. 
         */
        KubeControllerManagerArgs?: pulumi.Input<pulumi.Input<string>[]>;
        /**
         * KubeletArgs allows passing additional arguments to the kubelet service. 
         */
        KubeletArgs?: pulumi.Input<pulumi.Input<string>[]>;
        /**
         * NodeLabels set labels on registration. 
         */
        NodeLabels?: pulumi.Input<pulumi.Input<string>[]>;
        /**
         * NodeTaints are used to taint the node with key=value:effect. By default, server node is tainted with a couple of taints if number of agents nodes more than 0. 
         */
        NodeTaints?: pulumi.Input<pulumi.Input<string>[]>;
        /**
         * ServiceCidr defines the IP range from which service cluster IPs are allocated. Default is 10.140.0.0/16. 
         */
        ServiceCidr?: pulumi.Input<string>;
    }

    export interface K3supgraderConfigArgs {
        /**
         * ConfigEnv is a map of environment variables to pass to the controller. 
         */
        ConfigEnv?: pulumi.Input<pulumi.Input<string>[]>;
        Enabled?: pulumi.Input<boolean>;
        Helm?: pulumi.Input<inputs.cluster.HelmConfigArgs>;
        /**
         * Channel is a channel to use for the upgrade. Conflicts with Version. 
         */
        TargetChannel?: pulumi.Input<string>;
        /**
         * Version is a version to use for the upgrade. Conflicts with Channel. 
         */
        TargetVersion?: pulumi.Input<string>;
    }

    export interface K8sconfigBasicFirewallConfigArgs {
        /**
         * HetznerPublic is used to describe firewall attached to public k8s api endpoint. 
         */
        HetznerPublic?: pulumi.Input<inputs.cluster.K8sconfigHetnzerBasicFirewallConfigArgs>;
    }

    export interface K8sconfigConfigArgs {
        Addons?: pulumi.Input<inputs.cluster.AddonsConfigArgs>;
        AuditLog?: pulumi.Input<inputs.cluster.AuditAuditLogConfigArgs>;
        KubeAPIEndpoint?: pulumi.Input<inputs.cluster.K8sconfigK8SEndpointConfigArgs>;
    }

    export interface K8sconfigHetnzerBasicFirewallConfigArgs {
        /**
         * AllowedIps specifies a list of IP addresses that are permitted to access the k8s api endpoint. Only traffic from these IPs will be allowed if this list is configured. Default is 0.0.0.0/0 (all ipv4 addresses). 
         */
        AllowedIps?: pulumi.Input<pulumi.Input<string>[]>;
        /**
         * DisallowOwnIP is a security setting that, when enabled, prevents access to the server from deployer own public IP address. 
         */
        DisallowOwnIP?: pulumi.Input<boolean>;
    }

    export interface K8sconfigK8SEndpointConfigArgs {
        /**
         * Firewall defines configuration for the firewall attached to api access. This is used only for public type since private network considered to be secure. 
         */
        Firewall?: pulumi.Input<inputs.cluster.K8sconfigBasicFirewallConfigArgs>;
        /**
         * Type of k8s endpoint: public or private. Default is public. 
         */
        Type?: pulumi.Input<string>;
    }

    export interface K8sconfigNodeConfigArgs {
        /**
         * NodeLabels are used to label the node with key=value. 
         */
        NodeLabels?: pulumi.Input<pulumi.Input<string>[]>;
        /**
         * NodeTaints configures taint node manager. 
         */
        NodeTaints?: pulumi.Input<inputs.cluster.K8sconfigTaintConfigArgs>;
    }

    export interface K8sconfigTaintConfigArgs {
        /**
         * Do not add default taints to the server node. Default is false. 
         */
        DisableDefaultsTaints?: pulumi.Input<boolean>;
        /**
         * Enable or disable taint management. Default is false. 
         */
        Enabled?: pulumi.Input<boolean>;
        /**
         * Taints are used to taint the node with key=value:effect. Default is server node is tainted with a couple of taints if number of agents nodes more than 0. But only if disable-default-taints set to false. 
         */
        Taints?: pulumi.Input<pulumi.Input<string>[]>;
    }

    export interface NetworkConfigArgs {
        /**
         * CIDR of private network. Default is 10.20.0.0/16 
         */
        CIDR?: pulumi.Input<string>;
        /**
         * Enabled of not. Default is false. 
         */
        Enabled?: pulumi.Input<boolean>;
        /**
         * Network zone. Default is eu-central. 
         */
        Zone?: pulumi.Input<string>;
    }

    export interface OsconfigOSConfigArgs {
        JournalD?: pulumi.Input<inputs.cluster.JournaldConfigArgs>;
    }

}
