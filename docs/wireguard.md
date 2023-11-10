## Wireguard
### Firewall
By default, a hetzner firewall rule is added to allow all traffic to **51822** port for every traffic if wireguard enabled for in-cluster communication method. Restriction can be applied by specifying the following configuration:
```yaml
<project>:network:
    wireguard:
      enabled: true
      firewall:
        hetzner:
          allowed-ips:
            - '102.0.0.0/8'
```
