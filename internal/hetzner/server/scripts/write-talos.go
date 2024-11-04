package scripts

const WriteTalos = `
  #!/bin/bash
  # set -euo pipefail -x
  talos_version=v1.8.1
  arch=amd64
  TALOS_IMAGE="https://factory.talos.dev/image/1da3394e6229e507d4e3d166b718cacff86435a61c4765feedd66b43ac237558/v1.8.2/hcloud-amd64.raw.xz"
  WGET="wget --timeout=5 --waitretry=5 --tries=5 --retry-connrefused --inet4-only"

  apt-get install -y wget
  $WGET -O /tmp/talos.raw.xz ${TALOS_IMAGE}
  xz -d -c /tmp/talos.raw.xz | dd of=/dev/sda && sync
  # Reboot
  echo b > /proc/sysrq-trigger
`
