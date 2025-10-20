terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
  }
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

variable "cloudflare_api_token" {
  description = "The Cloudflare API Token"
  type        = string
  sensitive   = true
}

variable "zone_id" {
  description = "The zone ID"
  type        = string
  default     = "e2e3f4r5t6y7u8i9o0p1a2s3d4f5g6h7"
}

# Simple A record
resource "cloudflare_dns_record" "example_a" {
  zone_id = var.zone_id
  name    = "example"
  type    = "A"
  ttl     = 3600
  proxied = true
  comment = "Example A record"
  content = "192.0.2.1"
}

# AAAA record
resource "cloudflare_dns_record" "example_aaaa" {
  zone_id = var.zone_id
  name    = "ipv6"
  type    = "AAAA"
  ttl     = 1
  proxied = false
  content = "2001:db8::1"
}

# CNAME record
resource "cloudflare_dns_record" "example_cname" {
  zone_id = var.zone_id
  name    = "www"
  type    = "CNAME"
  proxied = true
  ttl     = 1
  content = "example.com"
}

# MX record with priority
resource "cloudflare_dns_record" "example_mx" {
  zone_id  = var.zone_id
  name     = "@"
  type     = "MX"
  priority = 10
  ttl      = 3600
  content  = "mail.example.com"
}

# TXT record
resource "cloudflare_dns_record" "example_txt" {
  zone_id = var.zone_id
  name    = "_dmarc"
  type    = "TXT"
  ttl     = 3600
  content = "v=DMARC1; p=reject; rua=mailto:dmarc@example.com"
}

# SRV record with data block (v4 style)
resource "cloudflare_dns_record" "example_srv" {
  zone_id = var.zone_id
  name    = "_sip._tcp"
  type    = "SRV"


  ttl      = 3600
  priority = 10
  data = {
    priority = 10
    weight   = 60
    port     = 5060
    target   = "sipserver.example.com"
    service  = "_sip"
    proto    = "_tcp"
    name     = "example.com"
  }
}

# CAA record with data block (v4 style)
resource "cloudflare_dns_record" "example_caa" {
  zone_id = var.zone_id
  name    = "@"
  type    = "CAA"


  ttl = 3600
  data = {
    flags = "0"
    tag   = "issue"
    value = "letsencrypt.org"
  }
}

# Multiple A records for load balancing
resource "cloudflare_dns_record" "lb_a_1" {
  zone_id = var.zone_id
  name    = "lb"
  type    = "A"
  ttl     = 60
  proxied = true
  content = "192.0.2.10"
}

resource "cloudflare_dns_record" "lb_a_2" {
  zone_id = var.zone_id
  name    = "lb"
  type    = "A"
  ttl     = 60
  proxied = true
  content = "192.0.2.11"
}

# NS record
resource "cloudflare_dns_record" "subdomain_ns" {
  zone_id = var.zone_id
  name    = "subdomain"
  type    = "NS"
  ttl     = 86400
  content = "ns1.example.com"
}

# PTR record
resource "cloudflare_dns_record" "example_ptr" {
  zone_id = var.zone_id
  name    = "1.2.0.192.in-addr.arpa"
  type    = "PTR"
  ttl     = 3600
  content = "host.example.com"
}

# Dynamic record creation
locals {
  subdomains = ["app1", "app2", "app3"]
}

resource "cloudflare_dns_record" "dynamic_records" {
  for_each = toset(local.subdomains)

  zone_id = var.zone_id
  name    = each.value
  type    = "A"
  ttl     = 3600
  proxied = true
  comment = "Dynamic record for ${each.value}"
  content = "192.0.2.50"
}

# Record with tags (if supported in v4)
resource "cloudflare_dns_record" "tagged_record" {
  zone_id = var.zone_id
  name    = "api"
  type    = "A"
  ttl     = 300
  proxied = false
  tags    = ["production", "api", "critical"]
  content = "192.0.2.100"
}