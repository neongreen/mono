# Data sources: look up Hetzner Cloud resources (no creation)
data "hcloud_server_type" "cx53" {
  name = "cx53"
}

data "hcloud_image" "ubuntu_2404" {
  name = "ubuntu-24.04"
}

data "hcloud_location" "hel1" {
  name = "hel1"
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

# =============================================================================
# SSH KEYS - DO NOT MODIFY
# =============================================================================
# WARNING: Changing ssh_keys on hcloud_server DESTROYS AND RECREATES THE SERVER.
# This will destroy all data, k3s cluster, and everything on the server.
#
# To add new SSH keys to an existing server, SSH in and edit ~/.ssh/authorized_keys
# manually. Do NOT add keys here.
#
# SSH keys are managed outside terraform (created manually in Hetzner console).
# We only look them up here for use in server creation.
# =============================================================================
data "hcloud_ssh_key" "workstation" {
  name = "workstation"
}

data "hcloud_ssh_key" "laptop" {
  name = "laptop"
}

# Single server: CX53 in Helsinki with k3s
resource "hcloud_server" "mono" {
  name        = "mono"
  image       = data.hcloud_image.ubuntu_2404.name
  server_type = data.hcloud_server_type.cx53.name
  location    = data.hcloud_location.hel1.name
  # DO NOT MODIFY ssh_keys - see warning at top of file
  ssh_keys = [data.hcloud_ssh_key.workstation.id, data.hcloud_ssh_key.laptop.id]

  user_data = <<-CLOUDCFG
  #cloud-config
  hostname: mono
  package_update: true
  package_upgrade: true

  users:
    - name: emily
      groups: sudo
      shell: /bin/bash
      sudo: ['ALL=(ALL) NOPASSWD:ALL']
      ssh_authorized_keys:
        - ${file("${path.module}/ssh_key.pub")}
        - ${file("${path.module}/ssh_key_laptop.pub")}

  ssh_pwauth: false

  runcmd:
    # Install gVisor
    - |
      set -e
      apt-get update
      apt-get install -y apt-transport-https ca-certificates curl gnupg
      curl -fsSL https://gvisor.dev/archive.key | gpg --dearmor -o /usr/share/keyrings/gvisor-archive-keyring.gpg
      echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/gvisor-archive-keyring.gpg] https://storage.googleapis.com/gvisor/releases release main" > /etc/apt/sources.list.d/gvisor.list
      apt-get update
      apt-get install -y runsc

    # Install k3s server (traefik disabled in favor of nginx-ingress deployed separately)
    - curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--write-kubeconfig-mode 0644 --disable traefik" sh -s -

    # Configure containerd for gVisor
    - |
      set -e
      mkdir -p /var/lib/rancher/k3s/agent/etc/containerd

      cat > /var/lib/rancher/k3s/agent/etc/containerd/config.toml.tmpl << 'EOF'
      {{ template "base" . }}

      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runsc]
        runtime_type = "io.containerd.runsc.v1"
      EOF

      cat > /var/lib/rancher/k3s/agent/etc/containerd/config-v3.toml.tmpl << 'EOF'
      {{ template "base" . }}

      [plugins."io.containerd.cri.v1.runtime".containerd.runtimes.runsc]
        runtime_type = "io.containerd.runsc.v1"
      EOF

      systemctl restart k3s
  CLOUDCFG
}

# Network attachment
resource "hcloud_server_network" "mono" {
  server_id  = hcloud_server.mono.id
  network_id = hcloud_network.main.id
}

# Firewall attachment
resource "hcloud_firewall_attachment" "mono" {
  firewall_id = hcloud_firewall.default.id
  server_ids  = [hcloud_server.mono.id]
}

