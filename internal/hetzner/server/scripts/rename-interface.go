package scripts

const RenameInterface = `
  # Based onhttps://github.com/kube-hetzner/terraform-hcloud-kube-hetzner/blob/eb779a2d24ed4e61ea679944912f92deba7283d2/locals.tf#L48
  # Added default gateway pinning
  #!/bin/bash
  set -euo pipefail -x

  # Wait for additional ip to be assigned
  until [[ $(ip link show | awk '/^3:/{print $2}' | sed 's/://g') ]]; do \
    sleep 1 ; \
  done
  INTERFACE=$(ip link show | awk '/^3:/{print $2}' | sed 's/://g')

  MAC=$(cat /sys/class/net/$INTERFACE/address)

  cat <<EOF > /etc/udev/rules.d/70-persistent-net.rules
  SUBSYSTEM=="net", ACTION=="add", DRIVERS=="?*", ATTR{address}=="$MAC", NAME="eth1"
  EOF

  ip link set $INTERFACE down
  ip link set $INTERFACE name eth1
  ip link set eth1 up

  eth0_connection=$(nmcli -g GENERAL.CONNECTION device show eth0)
  nmcli connection modify "$eth0_connection" \
    con-name eth0 \
    connection.interface-name eth0

  eth1_connection=$(nmcli -g GENERAL.CONNECTION device show eth1)
  nmcli connection modify "$eth1_connection" \
    con-name eth1 \
    connection.interface-name eth1
  nmcli connection modify eth1 \
    ipv4.never-default yes

  # Twice delete default route since there are two defaults routes
  ip r del default || true
  ip r del default || true

  # After restart the default route will be set
  systemctl restart NetworkManager

  ip r
`
