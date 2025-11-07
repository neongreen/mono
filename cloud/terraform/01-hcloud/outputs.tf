output "server_type" {
  value = data.hcloud_server_type.cpx11.id  # Smallest vCPU server
}

output "image" {
  value = data.hcloud_image.ubuntu_2404.id
}

output "location" {
  value = data.hcloud_location.nbg1.id  # German data center
}

output "network_id" {
  value = hcloud_network.main.id
}

output "firewall_id" {
  value = hcloud_firewall.default.id
}

output "server_id" {
  value = hcloud_server.vm1.id
}

output "server_ipv4" {
  value = hcloud_server.vm1.ipv4_address
}


