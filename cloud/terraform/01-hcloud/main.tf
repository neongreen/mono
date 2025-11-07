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


