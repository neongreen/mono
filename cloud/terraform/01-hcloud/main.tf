# Data sources: look up Hetzner Cloud resources (no creation)
data "hcloud_server_type" "cpx22" {
  name = "cpx22"
}

data "hcloud_image" "ubuntu_2404" {
  name = "ubuntu-24.04"
}

data "hcloud_location" "nbg1" {
  name = "nbg1"
}

# Network: private network for VMs
resource "hcloud_network" "main" {
  name     = "mono-net"
  ip_range = "10.0.0.0/16"
}

# Subnet: subnet within the network
resource "hcloud_network_subnet" "main" {
  network_id   = hcloud_network.main.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.0.0/24"
}

# Firewall: rules for inbound traffic
resource "hcloud_firewall" "default" {
  name = "mono-fw"

  rule {
    direction  = "in"
    protocol   = "icmp"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "22"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "80"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "443"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  # k3s API server
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "6443"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  # kubelet read-only metrics / node API (for remote admin)
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "10250"
    source_ips = ["0.0.0.0/0", "::/0"]
  }
}

# SSH key: your public key for server access
resource "hcloud_ssh_key" "workstation" {
  name       = "workstation"
  public_key = file("${path.module}/ssh_key.pub")
}

# Server: VM running Ubuntu with k3s installed
resource "hcloud_server" "vm1" {
  name        = "vm1"
  image       = data.hcloud_image.ubuntu_2404.name
  server_type = data.hcloud_server_type.cpx22.name
  location    = data.hcloud_location.nbg1.name
  ssh_keys    = [hcloud_ssh_key.workstation.id]

  # Cloud-init: runs on first boot to configure the server
  user_data = <<-CLOUDCFG
  #cloud-config
  hostname: vm1
  package_update: true
  package_upgrade: true

  users:
    - name: emily
      groups: sudo
      shell: /bin/bash
      sudo: ['ALL=(ALL) NOPASSWD:ALL']
      ssh_authorized_keys:
        - ${file("${path.module}/ssh_key.pub")}

  ssh_pwauth: false

  runcmd:
    # Install gVisor (runsc runtime for enhanced container isolation)
    - |
      set -e
      ARCH=$(uname -m)
      URL=https://storage.googleapis.com/gvisor/releases/release/latest/$${ARCH}
      wget $${URL}/runsc $${URL}/runsc.sha512
      sha512sum -c runsc.sha512
      rm -f runsc.sha512
      chmod a+rx runsc
      mv runsc /usr/local/bin/
    
    # Install k3s (lightweight Kubernetes)
    - curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--write-kubeconfig-mode 0644 --disable traefik" sh -s -
    
    # Configure containerd for gVisor (install creates the shim)
    - |
      set -e
      /usr/local/bin/runsc install
      mkdir -p /var/lib/rancher/k3s/agent/etc/containerd
      cat > /var/lib/rancher/k3s/agent/etc/containerd/config.toml.tmpl << 'EOF'
      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
        runtime_type = "io.containerd.runc.v2"
      
      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runsc]
        runtime_type = "io.containerd.runsc.v1"
      EOF
      systemctl restart k3s
  CLOUDCFG
}

# Network attachment: connect server to private network
resource "hcloud_server_network" "vm1" {
  server_id  = hcloud_server.vm1.id
  network_id = hcloud_network.main.id
}

# Firewall attachment: apply firewall rules to server
resource "hcloud_firewall_attachment" "vm1" {
  firewall_id = hcloud_firewall.default.id
  server_ids  = [hcloud_server.vm1.id]
}


