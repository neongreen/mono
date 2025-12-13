# Data sources: look up Hetzner Cloud resources (no creation)
data "hcloud_server_type" "cpx22" {
  name = "cpx22"
}

data "hcloud_server_type" "cpx51" {
  name = "cpx51"
}

data "hcloud_server_type" "cx53" {
  name = "cx53"
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
      apt-get update
      apt-get install -y apt-transport-https ca-certificates curl gnupg
      curl -fsSL https://gvisor.dev/archive.key | gpg --dearmor -o /usr/share/keyrings/gvisor-archive-keyring.gpg
      echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/gvisor-archive-keyring.gpg] https://storage.googleapis.com/gvisor/releases release main" > /etc/apt/sources.list.d/gvisor.list
      apt-get update
      apt-get install -y runsc

    # Install k3s (lightweight Kubernetes)
    - curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--write-kubeconfig-mode 0644 --disable traefik" sh -s -

    # Configure containerd for gVisor with k3s template syntax
    - |
      set -e
      mkdir -p /var/lib/rancher/k3s/agent/etc/containerd

      # Create config for containerd 1.7 (older k3s versions)
      cat > /var/lib/rancher/k3s/agent/etc/containerd/config.toml.tmpl << 'EOF'
      {{ template "base" . }}

      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runsc]
        runtime_type = "io.containerd.runsc.v1"
      EOF

      # Create config for containerd 2.0 (newer k3s versions - v1.31.6+)
      cat > /var/lib/rancher/k3s/agent/etc/containerd/config-v3.toml.tmpl << 'EOF'
      {{ template "base" . }}

      [plugins."io.containerd.cri.v1.runtime".containerd.runtimes.runsc]
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

# Server: VM2 for Onyx workloads (k3s agent node)
resource "hcloud_server" "vm2" {
  name        = "vm2-onyx"
  image       = data.hcloud_image.ubuntu_2404.name
  server_type = data.hcloud_server_type.cpx51.name
  location    = data.hcloud_location.nbg1.name
  ssh_keys    = [hcloud_ssh_key.workstation.id]

  # Depends on vm1 being created first
  depends_on = [hcloud_server.vm1]

  # Cloud-init: joins the k3s cluster as an agent node
  user_data = <<-CLOUDCFG
  #cloud-config
  hostname: vm2-onyx
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
      apt-get update
      apt-get install -y apt-transport-https ca-certificates curl gnupg
      curl -fsSL https://gvisor.dev/archive.key | gpg --dearmor -o /usr/share/keyrings/gvisor-archive-keyring.gpg
      echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/gvisor-archive-keyring.gpg] https://storage.googleapis.com/gvisor/releases release main" > /etc/apt/sources.list.d/gvisor.list
      apt-get update
      apt-get install -y runsc

    # Wait for vm1 to be ready and get k3s token
    - |
      set -e
      echo "Waiting for vm1 k3s server to be ready..."

      # Wait for vm1 to be accessible and k3s to be running
      VM1_IP="${hcloud_server.vm1.ipv4_address}"
      for i in {1..60}; do
        if timeout 5 bash -c "echo > /dev/tcp/$VM1_IP/6443" 2>/dev/null; then
          echo "vm1 k3s server is ready"
          break
        fi
        echo "Waiting for vm1... ($i/60)"
        sleep 10
      done

      # Get k3s token from vm1 via SSH
      echo "Fetching k3s token from vm1..."
      K3S_TOKEN=$(ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
        emily@$VM1_IP \
        'sudo cat /var/lib/rancher/k3s/server/node-token' 2>/dev/null || echo "")

      if [ -z "$K3S_TOKEN" ]; then
        echo "ERROR: Failed to get k3s token from vm1"
        exit 1
      fi

      # Join k3s cluster as an agent
      echo "Joining k3s cluster..."
      curl -sfL https://get.k3s.io | K3S_URL=https://$VM1_IP:6443 K3S_TOKEN=$K3S_TOKEN sh -s -

      # Configure containerd for gVisor
      mkdir -p /var/lib/rancher/k3s/agent/etc/containerd

      # Create config for containerd 1.7
      cat > /var/lib/rancher/k3s/agent/etc/containerd/config.toml.tmpl << 'EOF'
      {{ template "base" . }}

      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runsc]
        runtime_type = "io.containerd.runsc.v1"
      EOF

      # Create config for containerd 2.0
      cat > /var/lib/rancher/k3s/agent/etc/containerd/config-v3.toml.tmpl << 'EOF'
      {{ template "base" . }}

      [plugins."io.containerd.cri.v1.runtime".containerd.runtimes.runsc]
        runtime_type = "io.containerd.runsc.v1"
      EOF

      systemctl restart k3s-agent

      # Label this node for Onyx workloads
      # Note: This runs on vm1 since kubectl is there
      ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
        emily@$VM1_IP \
        "sudo kubectl label node vm2-onyx workload=onyx --overwrite" || echo "Failed to label node, will retry manually"

      echo "vm2-onyx successfully joined k3s cluster"
  CLOUDCFG
}

# Network attachment: connect vm2 to private network
resource "hcloud_server_network" "vm2" {
  server_id  = hcloud_server.vm2.id
  network_id = hcloud_network.main.id
}

# Firewall attachment: apply firewall rules to vm2
resource "hcloud_firewall_attachment" "vm2" {
  firewall_id = hcloud_firewall.default.id
  server_ids  = [hcloud_server.vm2.id]
}

# Server: VM3 for Coder workspaces (k3s agent node)
resource "hcloud_server" "vm3" {
  name        = "vm3-coder"
  image       = data.hcloud_image.ubuntu_2404.name
  server_type = data.hcloud_server_type.cx53.name
  location    = data.hcloud_location.nbg1.name
  ssh_keys    = [hcloud_ssh_key.workstation.id]

  depends_on = [hcloud_server.vm1]

  user_data = <<-CLOUDCFG
  #cloud-config
  hostname: vm3-coder
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
    # Install gVisor
    - |
      set -e
      apt-get update
      apt-get install -y apt-transport-https ca-certificates curl gnupg
      curl -fsSL https://gvisor.dev/archive.key | gpg --dearmor -o /usr/share/keyrings/gvisor-archive-keyring.gpg
      echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/gvisor-archive-keyring.gpg] https://storage.googleapis.com/gvisor/releases release main" > /etc/apt/sources.list.d/gvisor.list
      apt-get update
      apt-get install -y runsc

    # Join k3s cluster
    - |
      set -e
      echo "Waiting for vm1 k3s server to be ready..."

      VM1_IP="${hcloud_server.vm1.ipv4_address}"
      for i in {1..60}; do
        if timeout 5 bash -c "echo > /dev/tcp/$VM1_IP/6443" 2>/dev/null; then
          echo "vm1 k3s server is ready"
          break
        fi
        echo "Waiting for vm1... ($i/60)"
        sleep 10
      done

      echo "Fetching k3s token from vm1..."
      K3S_TOKEN=$(ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
        emily@$VM1_IP \
        'sudo cat /var/lib/rancher/k3s/server/node-token' 2>/dev/null || echo "")

      if [ -z "$K3S_TOKEN" ]; then
        echo "ERROR: Failed to get k3s token from vm1"
        exit 1
      fi

      echo "Joining k3s cluster..."
      curl -sfL https://get.k3s.io | K3S_URL=https://$VM1_IP:6443 K3S_TOKEN=$K3S_TOKEN sh -s -

      # Configure containerd for gVisor
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

      systemctl restart k3s-agent

      # Label node for Coder workloads
      ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
        emily@$VM1_IP \
        "sudo kubectl label node vm3-coder workload=coder --overwrite" || echo "Failed to label node, will retry manually"

      echo "vm3-coder successfully joined k3s cluster"
  CLOUDCFG
}

# Network attachment: connect vm3 to private network
resource "hcloud_server_network" "vm3" {
  server_id  = hcloud_server.vm3.id
  network_id = hcloud_network.main.id
}

# Firewall attachment: apply firewall rules to vm3
resource "hcloud_firewall_attachment" "vm3" {
  firewall_id = hcloud_firewall.default.id
  server_ids  = [hcloud_server.vm3.id]
}
