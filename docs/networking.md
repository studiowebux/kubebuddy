---
title: Networking
description: IP addresses, DNS, ports, and firewall management
tags: [networking, ip, dns, firewall, ports]
---

# Networking

KubeBuddy networking features allow you to manage IP addresses and their assignments to compute resources.

## IP Address Management

### IP Address Types

**Public IPs**: Internet-routable addresses
**Private IPs**: Internal network addresses

### IP States

- **available**: Ready for assignment
- **assigned**: Currently assigned to compute
- **reserved**: Reserved for future use

### Creating IP Addresses

Basic IP creation with upsert support (updates if address exists):

```bash
kubebuddy ip create \
  --address "192.168.1.10" \
  --type private \
  --cidr "192.168.1.0/24" \
  --provider "datacenter" \
  --region "us-east"
```

With gateway and DNS servers:

```bash
kubebuddy ip create \
  --address "10.0.1.100" \
  --type private \
  --cidr "10.0.1.0/24" \
  --gateway "10.0.1.1" \
  --dns "8.8.8.8,8.8.4.4" \
  --provider "aws" \
  --region "us-east-1"
```

Public IP with notes:

```bash
kubebuddy ip create \
  --address "203.0.113.45" \
  --type public \
  --cidr "203.0.113.0/24" \
  --provider "aws" \
  --region "us-east-1" \
  --notes "Production web server IP"
```

### Listing IP Addresses

List all IPs:

```bash
kubebuddy ip list
```

Filter by type:

```bash
kubebuddy ip list --type public
kubebuddy ip list --type private
```

Filter by provider and region:

```bash
kubebuddy ip list --provider aws --region us-east-1
```

Filter by state:

```bash
kubebuddy ip list --state available
kubebuddy ip list --state assigned
```

### Get IP Details

```bash
kubebuddy ip get <ip-id>
```

### Delete IP Address

```bash
kubebuddy ip delete <ip-id>
```

## IP Assignment

### Assign IP to Compute

Assign secondary IP:

```bash
kubebuddy ip assign \
  --compute <compute-id> \
  --ip <ip-id>
```

Assign as primary IP:

```bash
kubebuddy ip assign \
  --compute <compute-id> \
  --ip <ip-id> \
  --primary
```

### Unassign IP

```bash
kubebuddy ip unassign <assignment-id>
```

### List Assignments

List IPs for a compute:

```bash
kubebuddy ip list-assignments --compute <compute-id>
```

List computes using an IP:

```bash
kubebuddy ip list-assignments --ip <ip-id>
```

## Common Workflows

### Setup New Compute Network

```bash
# Create public IP
kubebuddy ip create \
  --address "203.0.113.45" \
  --type public \
  --cidr "203.0.113.0/24" \
  --provider "datacenter" \
  --region "us-east"

# Create private IP
kubebuddy ip create \
  --address "10.0.1.100" \
  --type private \
  --cidr "10.0.1.0/24" \
  --gateway "10.0.1.1" \
  --dns "8.8.8.8,8.8.4.4" \
  --provider "datacenter" \
  --region "us-east"

# Assign private IP as primary
kubebuddy ip assign \
  --compute <compute-id> \
  --ip <private-ip-id> \
  --primary

# Assign public IP as secondary
kubebuddy ip assign \
  --compute <compute-id> \
  --ip <public-ip-id>
```

### Migrate IP Between Computes

```bash
# List assignments for the IP
kubebuddy ip list-assignments --ip <ip-id>

# Unassign from old compute
kubebuddy ip unassign <assignment-id>

# Assign to new compute
kubebuddy ip assign \
  --compute <new-compute-id> \
  --ip <ip-id>
```

### IP Pool Management

Create an IP pool for a region:

```bash
for i in {10..20}; do
  kubebuddy ip create \
    --address "192.168.1.$i" \
    --type private \
    --cidr "192.168.1.0/24" \
    --gateway "192.168.1.1" \
    --provider "datacenter" \
    --region "us-east"
done
```

List available IPs:

```bash
kubebuddy ip list --state available
```

## DNS Management

### DNS Record Types

- **A**: IPv4 address record
- **AAAA**: IPv6 address record
- **CNAME**: Canonical name (alias)
- **PTR**: Reverse DNS pointer

### Creating DNS Records

Basic A record with upsert support (updates if name+type+zone exists):

```bash
kubebuddy dns create \
  --name "www.example.com" \
  --type A \
  --value "203.0.113.45" \
  --zone "example.com"
```

CNAME record:

```bash
kubebuddy dns create \
  --name "blog.example.com" \
  --type CNAME \
  --value "www.example.com" \
  --zone "example.com"
```

With custom TTL and IP link:

```bash
kubebuddy dns create \
  --name "api.example.com" \
  --type A \
  --value "203.0.113.50" \
  --zone "example.com" \
  --ttl 1800 \
  --ip <ip-id>
```

PTR record for reverse DNS:

```bash
kubebuddy dns create \
  --name "45.113.0.203.in-addr.arpa" \
  --type PTR \
  --value "www.example.com" \
  --zone "113.0.203.in-addr.arpa"
```

### Listing DNS Records

List all records:

```bash
kubebuddy dns list
```

Filter by type:

```bash
kubebuddy dns list --type A
kubebuddy dns list --type CNAME
```

Filter by zone:

```bash
kubebuddy dns list --zone example.com
```

Filter by name (partial match):

```bash
kubebuddy dns list --name "www"
```

Filter by linked IP:

```bash
kubebuddy dns list --ip <ip-id>
```

### Get DNS Details

```bash
kubebuddy dns get <record-id>
```

### Delete DNS Record

```bash
kubebuddy dns delete <record-id>
```

## Port Assignment Management

Port assignments map external ports on IP addresses to internal service ports.

### Port Assignment Structure

- **AssignmentID**: Links to service-to-compute assignment
- **IPID**: IP address for the port
- **Port**: External port number
- **Protocol**: tcp, udp, icmp, all
- **ServicePort**: Internal service port
- **Description**: Optional description

### Creating Port Assignments

Basic port mapping with upsert support (updates if ip+port+protocol exists):

```bash
kubebuddy port create \
  --assignment <assignment-id> \
  --ip <ip-id> \
  --port 8080 \
  --protocol tcp \
  --service-port 80 \
  --description "HTTP traffic"
```

HTTPS port mapping:

```bash
kubebuddy port create \
  --assignment <assignment-id> \
  --ip <ip-id> \
  --port 443 \
  --protocol tcp \
  --service-port 8443
```

UDP port for DNS:

```bash
kubebuddy port create \
  --assignment <assignment-id> \
  --ip <ip-id> \
  --port 53 \
  --protocol udp \
  --service-port 53 \
  --description "DNS"
```

### Listing Port Assignments

All port assignments:

```bash
kubebuddy port list
```

Filter by service assignment:

```bash
kubebuddy port list --assignment <assignment-id>
```

Filter by IP address:

```bash
kubebuddy port list --ip <ip-id>
```

Filter by protocol:

```bash
kubebuddy port list --protocol tcp
```

### Get Port Assignment Details

```bash
kubebuddy port get <port-assignment-id>
```

### Delete Port Assignment

```bash
kubebuddy port delete <port-assignment-id>
```

## Firewall Rule Management

Firewall rules define network access policies that can be assigned to computes.

### Firewall Rule Structure

- **Name**: Unique rule identifier
- **Action**: ALLOW or DENY
- **Protocol**: tcp, udp, icmp, all
- **Source**: Source CIDR, IP, or "any"
- **Destination**: Destination CIDR, IP, or "any"
- **PortStart/PortEnd**: Port range (optional)
- **Priority**: Lower values = higher priority (default: 100)
- **Description**: Optional description

### Creating Firewall Rules

Basic allow rule with upsert support (updates if name exists):

```bash
kubebuddy firewall create \
  --name "allow-http" \
  --action ALLOW \
  --protocol tcp \
  --source "any" \
  --destination "any" \
  --port-start 80 \
  --description "Allow HTTP traffic"
```

SSH access from specific network:

```bash
kubebuddy firewall create \
  --name "allow-ssh" \
  --action ALLOW \
  --protocol tcp \
  --source "192.168.1.0/24" \
  --destination "any" \
  --port-start 22 \
  --priority 50
```

Deny rule for specific port range:

```bash
kubebuddy firewall create \
  --name "deny-high-ports" \
  --action DENY \
  --protocol tcp \
  --source "any" \
  --destination "any" \
  --port-start 8000 \
  --port-end 9000 \
  --priority 200
```

Allow all from trusted network:

```bash
kubebuddy firewall create \
  --name "allow-internal" \
  --action ALLOW \
  --protocol all \
  --source "10.0.0.0/8" \
  --destination "any" \
  --priority 10
```

### Listing Firewall Rules

All rules (sorted by priority):

```bash
kubebuddy firewall list
```

Filter by action:

```bash
kubebuddy firewall list --action ALLOW
```

Filter by protocol:

```bash
kubebuddy firewall list --protocol tcp
```

### Get Firewall Rule Details

```bash
kubebuddy firewall get <rule-id>
```

### Delete Firewall Rule

```bash
kubebuddy firewall delete <rule-id>
```

### Assigning Firewall Rules to Computes

Assign rule to compute (enabled by default):

```bash
kubebuddy firewall assign \
  --compute <compute-id> \
  --rule <rule-id>
```

Assign but keep disabled:

```bash
kubebuddy firewall assign \
  --compute <compute-id> \
  --rule <rule-id> \
  --enabled=false
```

### List Firewall Assignments

By compute:

```bash
kubebuddy firewall list-assignments --compute <compute-id>
```

By rule:

```bash
kubebuddy firewall list-assignments --rule <rule-id>
```

### Unassign Firewall Rule

```bash
kubebuddy firewall unassign <assignment-id>
```

## Common DNS Workflows

### Setup Domain DNS

```bash
# A record for root domain
kubebuddy dns create \
  --name "example.com" \
  --type A \
  --value "203.0.113.45" \
  --zone "example.com"

# WWW CNAME
kubebuddy dns create \
  --name "www.example.com" \
  --type CNAME \
  --value "example.com" \
  --zone "example.com"

# Mail server
kubebuddy dns create \
  --name "mail.example.com" \
  --type A \
  --value "203.0.113.46" \
  --zone "example.com"

# API subdomain
kubebuddy dns create \
  --name "api.example.com" \
  --type A \
  --value "203.0.113.47" \
  --zone "example.com"
```

### Link DNS to IP

```bash
# Create IP first
IP_ID=$(kubebuddy ip create \
  --address "203.0.113.45" \
  --type public \
  --cidr "203.0.113.0/24" \
  --provider "datacenter" \
  --region "us-east" | jq -r '.id')

# Create DNS record linked to IP
kubebuddy dns create \
  --name "www.example.com" \
  --type A \
  --value "203.0.113.45" \
  --zone "example.com" \
  --ip "$IP_ID"
```

### Update DNS (Upsert)

Re-running create with same name+type+zone updates the record:

```bash
# Initial creation
kubebuddy dns create \
  --name "www.example.com" \
  --type A \
  --value "203.0.113.45" \
  --zone "example.com"

# Update value (upsert)
kubebuddy dns create \
  --name "www.example.com" \
  --type A \
  --value "203.0.113.50" \
  --zone "example.com"
```

## Auto-completion

IP, DNS, Port, and Firewall commands support shell completion for:
- IP types (public, private)
- IP states (available, assigned, reserved)
- IP IDs (shows address and provider/region)
- DNS record types (A, AAAA, CNAME, PTR)
- DNS record IDs (shows name, type, and zone)
- Port protocols (tcp, udp, icmp, all)
- Port assignment IDs
- Firewall actions (ALLOW, DENY)
- Firewall protocols (tcp, udp, icmp, all)
- Firewall rule IDs (shows name, action, and protocol)
- Service assignment IDs
- Compute IDs
