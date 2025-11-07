# Read-only lookups to verify connectivity; no resources created.

data "hcloud_server_type" "cpx11" {
  name = "cpx11"
}

data "hcloud_image" "ubuntu_2404" {
  name = "ubuntu-24.04"
}

data "hcloud_location" "nbg1" {
  name = "nbg1"
}

resource "hcloud_network" "main" {
  name     = "mono-net"
  ip_range = "10.0.0.0/16"
}

resource "hcloud_network_subnet" "main" {
  network_id   = hcloud_network.main.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.0.0/24"
}

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
}

resource "hcloud_ssh_key" "workstation" {
  name       = "workstation"
  public_key = file("${path.module}/ssh_key.pub")
}

resource "hcloud_server" "vm1" {
  name        = "vm1"
  image       = data.hcloud_image.ubuntu_2404.id
  server_type = data.hcloud_server_type.cpx11.id
  location    = data.hcloud_location.nbg1.id
  ssh_keys    = [hcloud_ssh_key.workstation.id]
}

resource "hcloud_server_network" "vm1" {
  server_id  = hcloud_server.vm1.id
  network_id = hcloud_network.main.id
}

resource "hcloud_firewall_attachment" "vm1" {
  firewall_id = hcloud_firewall.default.id
  server_ids  = [hcloud_server.vm1.id]
}


