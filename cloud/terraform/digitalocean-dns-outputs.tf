output "cloud_wildcard_fqdn" {
  description = "Wildcard DNS record for cloud subdomain"
  value       = digitalocean_record.cloud_wildcard.fqdn
}

output "cloud_root_fqdn" {
  description = "Root DNS record for cloud subdomain"
  value       = digitalocean_record.cloud_root.fqdn
}
