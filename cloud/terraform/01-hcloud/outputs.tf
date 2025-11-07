output "server_type" {
  value = data.hcloud_server_type.cpx11.id
}

output "image" {
  value = data.hcloud_image.ubuntu_2404.id
}

output "location" {
  value = data.hcloud_location.nbg1.id
}


