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

resource "hcloud_subnetwork" "main" {
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


