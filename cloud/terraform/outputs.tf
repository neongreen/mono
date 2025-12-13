output "server_type" {
  value = data.hcloud_server_type.cx53.id
}

output "image" {
  value = data.hcloud_image.ubuntu_2404.id
}

output "location" {
  value = data.hcloud_location.hel1.id
}

output "network_id" {
  value = hcloud_network.main.id
}

output "firewall_id" {
  value = hcloud_firewall.default.id
}

output "mono_server_id" {
  value = hcloud_server.mono.id
}

output "mono_server_ipv4" {
  value = hcloud_server.mono.ipv4_address
}
