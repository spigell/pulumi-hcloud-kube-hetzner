// *** WARNING: this file was generated by Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as inputs from "../types/input";
import * as outputs from "../types/output";

export namespace cluster {
    export interface AddonsConfig {
        /**
         * CCM defines configuration [hetzner-cloud-controller-manager](https://github.com/hetznercloud/hcloud-cloud-controller-manager). 
         */
        CCM?: outputs.cluster.CcmConfig;
        /**
         * K3SSystemUpgrader defines configuration for [system-upgrade-controller](https://github.com/rancher/system-upgrade-controller). 
         */
        K3SSystemUpgrader?: outputs.cluster.K3supgraderConfig;
    }

    export interface AuditAuditLogConfig {
        /**
         * AuditLogMaxAge defines the maximum number of days to retain old audit log files. Default is 10. 
         */
        AuditLogMaxAge?: number;
        /**
         * AuditLogMaxBackup specifies the maximum number of audit log files to retain. Default is 30. 
         */
        AuditLogMaxBackup?: number;
        /**
         * AuditLogMaxSize specifies the maximum size in megabytes of the audit log file before it gets rotated. Default is 100. 
         */
        AuditLogMaxSize?: number;
        /**
         * Enabled specifies if the audit log is enabled. If nil, it might default to a cluster-level setting. Default is true. 
         */
        Enabled?: boolean;
        /**
         * PolicyFilePath is the path to the local file that defines the audit policy configuration. 
         */
        PolicyFilePath?: string;
    }

    export interface CcmConfig {
        /**
         * Enabled is a flag to enable or disable hcloud CCM. 
         */
        Enabled?: boolean;
        Helm?: outputs.cluster.HelmConfig;
        /**
         * DefaultloadbalancerLocation is a default location for the loadbancers. 
         */
        LoadbalancersDefaultLocation?: string;
        /**
         * LoadbalancersEnabled is a flag to enable or disable loadbalancers management. Note: internal loadbalancer for k3s will be disabled. 
         */
        LoadbalancersEnabled?: boolean;
        /**
         * Token is a hcloud token to access hcloud API for CCM. 
         */
        Token?: string;
    }

    export interface ConfigConfig {
        /**
         * Defaults is a map with default settings for agents and servers. Global values for all nodes can be set here as well. Default is not specified. 
         */
        Defaults?: outputs.cluster.ConfigDefaultConfig;
        /**
         * K8S defines a distribution-agnostic cluster configuration. Default is not specified. 
         */
        K8S?: outputs.cluster.K8sconfigConfig;
        /**
         * Network defines network configuration for cluster. Default is not specified. 
         */
        Network?: outputs.cluster.ConfigNetworkConfig;
        /**
         * Nodepools is a map with agents and servers defined. Required for at least one server node. Default is not specified. 
         */
        Nodepools?: outputs.cluster.ConfigNodepoolsConfig;
    }

    export interface ConfigDefaultConfig {
        /**
         * Agents holds configuration settings specific to agent nodes, overriding Global settings where specified. 
         */
        Agents?: outputs.cluster.ConfigNodeConfig;
        /**
         * Global provides configuration settings that are applied to all nodes, unless overridden by specific roles. 
         */
        Global?: outputs.cluster.ConfigNodeConfig;
        /**
         * Servers holds configuration settings specific to server nodes, overriding Global settings where specified. 
         */
        Servers?: outputs.cluster.ConfigNodeConfig;
    }

    export interface ConfigFirewallConfig {
        /**
         * Hetzner specify firewall configuration for cloud firewall. 
         */
        Hetzner?: outputs.cluster.FirewallConfig;
    }

    export interface ConfigNetworkConfig {
        /**
         * Hetzner specifies network configuration for private networking. 
         */
        Hetzner?: outputs.cluster.NetworkConfig;
    }

    export interface ConfigNodeConfig {
        /**
         * K3S is the configuration of a k3s cluster. 
         */
        K3s?: outputs.cluster.K3sConfig;
        /**
         * K8S is common configuration for nodes. 
         */
        K8S?: outputs.cluster.K8sconfigNodeConfig;
        /**
         * Leader specifies the leader of a multi-master cluster. Required if the number of masters is more than 1. Default is not specified. 
         */
        Leader?: boolean;
        /**
         * NodeID is the id of a server. It is used throughout the entire program as a key. Required. Default is not specified. 
         */
        NodeID?: string;
        /**
         * OS defines configuration for operating system. 
         */
        OS?: outputs.cluster.OsconfigOSConfig;
        /**
         * Server is the configuration of a Hetzner server. 
         */
        Server?: outputs.cluster.ConfigServerConfig;
    }

    export interface ConfigNodepoolConfig {
        /**
         * Config is the default node configuration for the group. 
         */
        Config?: outputs.cluster.ConfigNodeConfig;
        /**
         * Nodes is a list of nodes inside of the group. 
         */
        Nodes?: outputs.cluster.ConfigNodeConfig[];
        /**
         * PoolID is id of group of servers. It is used through the entire program as key for the group. Required. Default is not specified. 
         */
        PoolID?: string;
    }

    export interface ConfigNodepoolsConfig {
        /**
         * Agents is a list of NodepoolConfig objects, each representing a configuration for a pool of agent nodes. 
         */
        Agents?: outputs.cluster.ConfigNodepoolConfig[];
        /**
         * Servers is a list of NodepoolConfig objects, each representing a configuration for a pool of server nodes. 
         */
        Servers?: outputs.cluster.ConfigNodepoolConfig[];
    }

    export interface ConfigServerConfig {
        /**
         * AdditionalSSHKeys contains a list of additional public SSH keys to install in the server's user account. 
         */
        AdditionalSSHKeys?: string[];
        /**
         * Firewall points to an optional configuration for a firewall to be associated with the server. 
         */
        Firewall?: outputs.cluster.ConfigFirewallConfig;
        /**
         * Hostname is the desired hostname to assign to the server. Default is `phkh-${name-of-stack}-${name-of-cluster}-${id-of-node}`. 
         */
        Hostname?: string;
        /**
         * Image specifies the operating system image to use for the server (e.g., "ubuntu-20.04" or id of private image). Default is autodiscovered. 
         */
        Image?: string;
        /**
         * Location specifies the physical location or data center where the server will be hosted (e.g., "fsn1"). Default is hel1. 
         */
        Location?: string;
        /**
         * ServerType specifies the type of server to be provisioned (e.g., "cx11", "cx21"). Default is cx21. 
         */
        ServerType?: string;
        /**
         * UserName is the primary user account name that will be created on the server. Default is rancher. 
         */
        UserName?: string;
        /**
         * UserPasswd is the password for the primary user account on the server. 
         */
        UserPasswd?: string;
    }

    export interface FirewallConfig {
        /**
         * AdditionalRules is a list of additional rules to be applied. 
         */
        AdditionalRules?: outputs.cluster.FirewallRuleConfig[];
        /**
         * AllowICMP indicates whether ICMP traffic is allowed. Default is false. 
         */
        AllowICMP?: boolean;
        /**
         * Enabled specifies if the configuration is active. Default is false. 
         */
        Enabled?: boolean;
        /**
         * SSH holds the SSH specific configurations. 
         */
        SSH?: outputs.cluster.FirewallSSHConfig;
    }

    export interface FirewallRuleConfig {
        /**
         * Description provides a human-readable explanation of what the rule is intended to do. 
         */
        Description?: string;
        /**
         * Port specifies the network port number or range applicable for the rule. Required. 
         */
        Port?: string;
        /**
         * Protocol specifies the network protocol (e.g., TCP, UDP) applicable for the rule. Default is TCP. 
         */
        Protocol?: string;
        /**
         * SourceIps lists IP addresses or subnets from which traffic is allowed or to which traffic is directed, based on the Direction. Required. 
         */
        SourceIps?: string[];
    }

    export interface FirewallSSHConfig {
        /**
         * Allow indicates whether SSH access is permitted. Default is false. 
         */
        Allow?: boolean;
        /**
         * AllowedIps lists specific IP addresses that are permitted to access via SSH. 
         */
        AllowedIps?: string[];
        /**
         * DisallowOwnIP specifies whether SSH access from the deployer's own IP address is disallowed. Default is false. 
         */
        DisallowOwnIP?: boolean;
    }

    export interface HelmConfig {
        /**
         * ValuesFilePaths is a list of path/to/file to values files. See https://www.pulumi.com/registry/packages/kubernetes/api-docs/helm/v3/release/#valueyamlfiles_nodejs for details. 
         */
        ValuesFilePath?: string[];
        /**
         * Version is version of helm chart. Default is taken from default-helm-versions.yaml in template's versions directory. 
         */
        Version?: string;
    }

    export interface JournaldConfig {
        /**
         * GatherAuditD indicates whether auditd logs should be gathered. Default is true. 
         */
        GatherAuditD?: boolean;
        /**
         * GatherToLeader indicates whether journald logs should be sent to the leader node. Default is true. 
         */
        GatherToLeader?: boolean;
    }

    export interface K3sConfig {
        /**
         * [Experimental] clean-data-on-upgrade is used to delete all data while upgrade. This is based on the script https://docs.k3s.io/upgrades/killall 
         */
        CleanDataOnUpgrade?: boolean;
        /**
         * The real config of k3s service. 
         */
        K3S?: outputs.cluster.K3sK3sConfig;
        /**
         * Version is used to determine if k3s should be upgraded if auto-upgrade is disabled. If the version is changed, k3s will be upgraded. 
         */
        Version?: string;
    }

    export interface K3sK3sConfig {
        /**
         * ClusterCidr defines the IP range from which pod IPs shall be allocated. Default is 10.141.0.0/16. 
         */
        ClusterCidr?: string;
        /**
         * ClusterDNS specifies the IP address of the DNS service within the cluster. Default is autopicked. 
         */
        ClusterDNS?: string;
        /**
         * ClusterDomain specifies the domain name of the cluster. 
         */
        ClusterDomain?: string;
        /**
         * Disable lists components or features to disable. 
         */
        Disable?: string[];
        /**
         * DisableCloudController determines whether to disable the integrated cloud controller manager. Default is false, but will be true if ccm is enabled. 
         */
        DisableCloudController?: boolean;
        /**
         * DisableNetworkPolicy determines whether to disable network policies. 
         */
        DisableNetworkPolicy?: boolean;
        /**
         * FlannelBackend determines the type of backend used for Flannel, a networking solution. 
         */
        FlannelBackend?: string;
        /**
         * KubeAPIServerArgs allows passing additional arguments to the Kubernetes API server. 
         */
        KubeAPIServerArgs?: string[];
        /**
         * KubeCloudControllerManagerArgs allows passing additional arguments to the Kubernetes cloud controller manager. 
         */
        KubeCloudControllerManagerArgs?: string[];
        /**
         * KubeControllerManagerArgs allows passing additional arguments to the Kubernetes controller manager. 
         */
        KubeControllerManagerArgs?: string[];
        /**
         * KubeletArgs allows passing additional arguments to the kubelet service. 
         */
        KubeletArgs?: string[];
        /**
         * NodeLabels set labels on registration. 
         */
        NodeLabels?: string[];
        /**
         * NodeTaints are used to taint the node with key=value:effect. By default, server node is tainted with a couple of taints if number of agents nodes more than 0. 
         */
        NodeTaints?: string[];
        /**
         * ServiceCidr defines the IP range from which service cluster IPs are allocated. Default is 10.140.0.0/16. 
         */
        ServiceCidr?: string;
    }

    export interface K3supgraderConfig {
        /**
         * ConfigEnv is a map of environment variables to pass to the controller. 
         */
        ConfigEnv?: string[];
        Enabled?: boolean;
        Helm?: outputs.cluster.HelmConfig;
        /**
         * Channel is a channel to use for the upgrade. Conflicts with Version. 
         */
        TargetChannel?: string;
        /**
         * Version is a version to use for the upgrade. Conflicts with Channel. 
         */
        TargetVersion?: string;
    }

    export interface K8sconfigBasicFirewallConfig {
        /**
         * HetznerPublic is used to describe firewall attached to public k8s api endpoint. 
         */
        HetznerPublic?: outputs.cluster.K8sconfigHetnzerBasicFirewallConfig;
    }

    export interface K8sconfigConfig {
        Addons?: outputs.cluster.AddonsConfig;
        AuditLog?: outputs.cluster.AuditAuditLogConfig;
        KubeAPIEndpoint?: outputs.cluster.K8sconfigK8SEndpointConfig;
    }

    export interface K8sconfigHetnzerBasicFirewallConfig {
        /**
         * AllowedIps specifies a list of IP addresses that are permitted to access the k8s api endpoint. Only traffic from these IPs will be allowed if this list is configured. Default is 0.0.0.0/0 (all ipv4 addresses). 
         */
        AllowedIps?: string[];
        /**
         * DisallowOwnIP is a security setting that, when enabled, prevents access to the server from deployer own public IP address. 
         */
        DisallowOwnIP?: boolean;
    }

    export interface K8sconfigK8SEndpointConfig {
        /**
         * Firewall defines configuration for the firewall attached to api access. This is used only for public type since private network considered to be secure. 
         */
        Firewall?: outputs.cluster.K8sconfigBasicFirewallConfig;
        /**
         * Type of k8s endpoint: public or private. Default is public. 
         */
        Type?: string;
    }

    export interface K8sconfigNodeConfig {
        /**
         * NodeLabels are used to label the node with key=value. 
         */
        NodeLabels?: string[];
        /**
         * NodeTaints configures taint node manager. 
         */
        NodeTaints?: outputs.cluster.K8sconfigTaintConfig;
    }

    export interface K8sconfigTaintConfig {
        /**
         * Do not add default taints to the server node. Default is false. 
         */
        DisableDefaultsTaints?: boolean;
        /**
         * Enable or disable taint management. Default is false. 
         */
        Enabled?: boolean;
        /**
         * Taints are used to taint the node with key=value:effect. Default is server node is tainted with a couple of taints if number of agents nodes more than 0. But only if disable-default-taints set to false. 
         */
        Taints?: string[];
    }

    export interface NetworkConfig {
        /**
         * CIDR of private network. Default is 10.20.0.0/16 
         */
        CIDR?: string;
        /**
         * Enabled of not. Default is false. 
         */
        Enabled?: boolean;
        /**
         * Network zone. Default is eu-central. 
         */
        Zone?: string;
    }

    export interface OsconfigOSConfig {
        JournalD?: outputs.cluster.JournaldConfig;
    }

    export interface Servers {
        internalIP?: string;
        ip?: string;
        name?: string;
        user?: string;
    }

}
