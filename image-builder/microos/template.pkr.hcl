packer {
  required_plugins {
    hcloud = {
      version = ">= 1.0.0"
      source  = "github.com/hashicorp/hcloud"
    }
  }
}

source "hcloud" "main" {
  image       = "ubuntu-20.04"
  location    = "hel1"
  rescue      = "linux64"
  server_type = "cpx11"
  snapshot_labels = {
    // Took from terraform-hcloud-kube-hetzner
    microos-snapshot = "yes"
    creator          = "phkh-builder"
  }
  snapshot_name = "microos-amd64-{{ isotime `2006-01-02` }}"
  ssh_username  = "root"
}


build {
  sources = ["source.hcloud.main"]

  provisioner "shell" {
    only        = [ "hcloud.main" ]
    inline = [
      "set -ex",
      "apt-get update",
      "apt-get install -y aria2 qemu-utils",
      "aria2c --follow-metalink=mem https://download.opensuse.org/tumbleweed/appliances/openSUSE-MicroOS.x86_64-OpenStack-Cloud.qcow2.meta4",
      "qemu-img convert -p -f qcow2 -O host_device $(ls -a | grep -ie '^opensuse.*microos.*qcow2$') /dev/sda"
    ]
  }
}
