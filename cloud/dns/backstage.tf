terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.36"
    }
  }
}

provider "digitalocean" {
  token = var.do_token
}

variable "do_token" {
  type      = string
  sensitive = true
}

variable "domain" {
  type    = string
  default = "artyom.me"
}

variable "hostname" {
  type    = string
  default = "backstage"
}

variable "address" {
  type    = string
  default = "91.98.238.65"
}

resource "digitalocean_record" "backstage" {
  domain = var.domain
  type   = "A"
  name   = var.hostname
  value  = var.address
  ttl    = 60
}
