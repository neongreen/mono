# Data source to reference existing domain
data "digitalocean_domain" "artyom_me" {
  name = "artyom.me"
}

# Wildcard DNS record for cloud subdomain
# This allows *.cloud.artyom.me to resolve to the k3s cluster
resource "digitalocean_record" "cloud_wildcard" {
  domain = data.digitalocean_domain.artyom_me.id
  type   = "A"
  name   = "*.cloud"
  value  = "91.98.238.65" # Hetzner k3s cluster IP
  ttl    = 300
}

# Optional: Direct A record for cloud.artyom.me itself
resource "digitalocean_record" "cloud_root" {
  domain = data.digitalocean_domain.artyom_me.id
  type   = "A"
  name   = "cloud"
  value  = "91.98.238.65"
  ttl    = 300
}
