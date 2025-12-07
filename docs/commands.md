---
title: Command Reference
description: Complete CLI command reference
tags: [cli, reference, commands]
---

# Command Reference

## server

Start API server.

```bash
kubebuddy server [flags]
```

**Flags:**

- `--db`: Database file path (default: ./kubebuddy.db)
- `--port`: Server port (default: 8080)
- `--create-admin-key`: Create admin API key from ADMIN_API_KEY env var
- `--seed`: Seed database with sample data

**Example:**

```bash
kubebuddy server --db /data/kubebuddy.db --port 9000 --create-admin-key
```

## compute

Manage compute resources.

### list

List all computes.

```bash
kubebuddy compute list
```

### get

Get compute by ID.

```bash
kubebuddy compute get <id>
```

### create

Create compute.

```bash
kubebuddy compute create \
  --name "prod-server-01" \
  --type baremetal \
  --provider ovh \
  --region us-east \
  --tags "env=prod,zone=us-east-1"
```

**Flags:**

- `--name`: Compute name (required)
- `--type`: baremetal, vps, vm (default: baremetal)
- `--provider`: Provider name (required)
- `--region`: Region (required)
- `--tags`: Tags as key=value pairs, comma-separated

### update

Update compute.

```bash
kubebuddy compute update <id> --tags "env=prod,zone=us-west"
kubebuddy compute update <id> --state maintenance
```

**Flags:**

- `--name`: Compute name
- `--type`: baremetal, vps, vm
- `--provider`: Provider name
- `--region`: Region
- `--tags`: Tags as key=value pairs, comma-separated
- `--state`: State (active, inactive, maintenance)

### delete

Delete compute.

```bash
kubebuddy compute delete <id>
```

## component

Manage hardware components.

### list

List components.

```bash
kubebuddy component list
kubebuddy component list --type cpu --manufacturer Intel
```

**Flags:**

- `--type`: Filter by type (cpu, ram, storage, gpu, nic, psu)
- `--manufacturer`: Filter by manufacturer

### get

Get component details.

```bash
kubebuddy component get <id>
```

### create

Create component.

```bash
kubebuddy component create \
  --name "Intel Xeon E5-2680v4" \
  --type cpu \
  --manufacturer Intel \
  --model "E5-2680v4" \
  --specs '{"cores":14,"threads":28,"ghz":2.4}'

kubebuddy component create \
  --name "Samsung 32GB DDR4" \
  --type ram \
  --manufacturer Samsung \
  --model "M393A4K40CB2" \
  --specs '{"capacity_gb":32,"speed_mhz":2400}'

kubebuddy component create \
  --name "Samsung 960GB NVMe" \
  --type storage \
  --manufacturer Samsung \
  --model "PM983" \
  --specs '{"capacity_gb":960,"type":"nvme"}'
```

**Flags:**

- `--name`: Component name (required)
- `--type`: Component type (required)
- `--manufacturer`: Manufacturer (required)
- `--model`: Model (required)
- `--specs`: JSON specs (e.g., `{"cores":8,"ghz":3.5}`)
- `--notes`: Notes

### delete

Delete component.

```bash
kubebuddy component delete <id>
```

### assign

Assign component to compute.

```bash
kubebuddy component assign \
  --compute server-01 \
  --component cpu-123 \
  --quantity 2 \
  --slot "CPU1,CPU2"

# Storage with RAID1
kubebuddy component assign \
  --compute server-01 \
  --component nvme-500gb \
  --quantity 2 \
  --raid raid1 \
  --raid-group storage-group-1
```

**Flags:**

- `--compute`: Compute ID (required)
- `--component`: Component ID (required)
- `--quantity`: Quantity (default: 1)
- `--slot`: Physical slot (e.g., CPU1, DIMM0-3)
- `--serial`: Serial number
- `--notes`: Notes
- `--raid`: RAID level (raid0, raid1, raid5, raid6, raid10)
- `--raid-group`: RAID group ID (components with same group form array)

### unassign

Unassign component.

```bash
kubebuddy component unassign <id>
```

### list-assignments

List component assignments.

```bash
kubebuddy component list-assignments
kubebuddy component list-assignments --compute server-01
kubebuddy component list-assignments --component cpu-123
```

**Flags:**

- `--compute`: Filter by compute ID
- `--component`: Filter by component ID

## ip

Manage IP addresses and assignments.

### list

List IP addresses.

```bash
kubebuddy ip list
kubebuddy ip list --type public
kubebuddy ip list --provider aws --region us-east-1
kubebuddy ip list --state available
```

**Flags:**

- `--type`: Filter by type (public, private)
- `--provider`: Filter by provider
- `--region`: Filter by region
- `--state`: Filter by state (available, assigned, reserved)

### get

Get IP address details.

```bash
kubebuddy ip get <id>
```

### create

Create IP address (upserts by address).

```bash
kubebuddy ip create \
  --address "192.168.1.10" \
  --type private \
  --cidr "192.168.1.0/24" \
  --provider "datacenter" \
  --region "us-east"

kubebuddy ip create \
  --address "10.0.1.100" \
  --type private \
  --cidr "10.0.1.0/24" \
  --gateway "10.0.1.1" \
  --dns "8.8.8.8,8.8.4.4" \
  --provider "aws" \
  --region "us-east-1"
```

**Flags:**

- `--address`: IP address (required)
- `--type`: IP type - public or private (required)
- `--cidr`: CIDR notation (e.g., 192.168.1.0/24) (required)
- `--gateway`: Gateway address
- `--dns`: DNS servers (comma-separated)
- `--provider`: Provider (required)
- `--region`: Region (required)
- `--notes`: Notes
- `--state`: State (available, assigned, reserved) (default: available)

### delete

Delete IP address.

```bash
kubebuddy ip delete <id>
```

### assign

Assign IP to compute.

```bash
kubebuddy ip assign \
  --compute server-01 \
  --ip ip-123

kubebuddy ip assign \
  --compute server-01 \
  --ip ip-456 \
  --primary
```

**Flags:**

- `--compute`: Compute ID (required)
- `--ip`: IP address ID (required)
- `--primary`: Set as primary IP

### unassign

Unassign IP from compute.

```bash
kubebuddy ip unassign <assignment-id>
```

### list-assignments

List IP assignments.

```bash
kubebuddy ip list-assignments
kubebuddy ip list-assignments --compute server-01
kubebuddy ip list-assignments --ip ip-123
```

**Flags:**

- `--compute`: Filter by compute ID
- `--ip`: Filter by IP address ID

## dns

Manage DNS records (A, AAAA, CNAME, PTR).

### list

List DNS records.

```bash
kubebuddy dns list
kubebuddy dns list --type A
kubebuddy dns list --zone example.com
kubebuddy dns list --name "www"
kubebuddy dns list --ip <ip-id>
```

**Flags:**

- `--type`: Filter by record type (A, AAAA, CNAME, PTR)
- `--zone`: Filter by DNS zone
- `--ip`: Filter by linked IP address ID
- `--name`: Filter by name (partial match)

### get

Get DNS record details.

```bash
kubebuddy dns get <id>
```

### create

Create DNS record (upserts by name+type+zone).

```bash
kubebuddy dns create \
  --name "www.example.com" \
  --type A \
  --value "203.0.113.45" \
  --zone "example.com"

kubebuddy dns create \
  --name "blog.example.com" \
  --type CNAME \
  --value "www.example.com" \
  --zone "example.com"

kubebuddy dns create \
  --name "api.example.com" \
  --type A \
  --value "203.0.113.50" \
  --zone "example.com" \
  --ttl 1800 \
  --ip <ip-id>
```

**Flags:**

- `--name`: DNS record name (e.g., www.example.com) (required)
- `--type`: Record type - A, AAAA, CNAME, PTR (required)
- `--value`: Record value (IP or hostname) (required)
- `--zone`: DNS zone (e.g., example.com) (required)
- `--ttl`: TTL in seconds (default: 3600)
- `--ip`: Link to IP address ID (optional)
- `--notes`: Notes

### delete

Delete DNS record.

```bash
kubebuddy dns delete <id>
```

## port

Manage port assignments (external to service port mappings).

### list

List port assignments.

```bash
kubebuddy port list
kubebuddy port list --assignment <assignment-id>
kubebuddy port list --ip <ip-id>
kubebuddy port list --protocol tcp
```

**Flags:**

- `--assignment`: Filter by service assignment ID
- `--ip`: Filter by IP address ID
- `--protocol`: Filter by protocol (tcp, udp, icmp, all)

### get

Get port assignment details.

```bash
kubebuddy port get <id>
```

### create

Create port assignment (upserts by ip+port+protocol).

```bash
kubebuddy port create \
  --assignment <assignment-id> \
  --ip <ip-id> \
  --port 8080 \
  --protocol tcp \
  --service-port 80 \
  --description "HTTP traffic"
```

**Flags:**

- `--assignment`: Service assignment ID (required)
- `--ip`: IP address ID (required)
- `--port`: External port number (required)
- `--protocol`: Protocol - tcp, udp, icmp, all (default: tcp)
- `--service-port`: Internal service port (required)
- `--description`: Description

### delete

Delete port assignment.

```bash
kubebuddy port delete <id>
```

## firewall

Manage firewall rules and assignments to computes.

### list

List firewall rules.

```bash
kubebuddy firewall list
kubebuddy firewall list --action ALLOW
kubebuddy firewall list --protocol tcp
```

**Flags:**

- `--action`: Filter by action (ALLOW, DENY)
- `--protocol`: Filter by protocol (tcp, udp, icmp, all)

### get

Get firewall rule details.

```bash
kubebuddy firewall get <id>
```

### create

Create firewall rule (upserts by name).

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

**Flags:**

- `--name`: Rule name (required, unique)
- `--action`: Action - ALLOW or DENY (required)
- `--protocol`: Protocol - tcp, udp, icmp, all (required)
- `--source`: Source CIDR, IP, or "any" (required)
- `--destination`: Destination CIDR, IP, or "any" (required)
- `--port-start`: Port start (0 for any)
- `--port-end`: Port end (0 for single port)
- `--description`: Description
- `--priority`: Priority (default: 100, lower = higher priority)

### delete

Delete firewall rule.

```bash
kubebuddy firewall delete <id>
```

### assign

Assign firewall rule to compute.

```bash
kubebuddy firewall assign \
  --compute <compute-id> \
  --rule <rule-id>
```

**Flags:**

- `--compute`: Compute ID (required)
- `--rule`: Firewall rule ID (required)
- `--enabled`: Enable rule (default: true)

### unassign

Unassign firewall rule from compute.

```bash
kubebuddy firewall unassign <assignment-id>
```

### list-assignments

List firewall rule assignments.

```bash
kubebuddy firewall list-assignments --compute <compute-id>
kubebuddy firewall list-assignments --rule <rule-id>
```

**Flags:**

- `--compute`: Filter by compute ID
- `--rule`: Filter by firewall rule ID

## service

Manage services.

### list

List all services.

```bash
kubebuddy service list
```

### get

Get service by ID.

```bash
kubebuddy service get <id>
```

### create

Create service (upserts by name).

```bash
kubebuddy service create \
  --name "postgres-db" \
  --min-spec '{"cores":2,"memory":4096,"nvme":100}' \
  --max-spec '{"cores":4,"memory":8192,"nvme":200}'

kubebuddy service create \
  --name "web-server" \
  --min-spec '{"cores":1,"memory":2048}' \
  --max-spec '{"cores":2,"memory":4096}'
```

**Flags:**

- `--name`: Service name (required)
- `--min-spec`: Minimum resources JSON (e.g., `{"cores":2,"memory":4096}`)
- `--max-spec`: Maximum resources JSON
- `--placement`: Placement rules JSON

**Resource keys**: cores, memory (MB), vram (MB), nvme (GB), gpu (count)

### delete

Delete service.

```bash
kubebuddy service delete <id>
```

## assignment

Manage service-to-compute assignments.

### list

List assignments.

```bash
kubebuddy assignment list
kubebuddy assignment list --compute server-01
kubebuddy assignment list --service postgres-db
```

**Flags:**

- `--compute`: Filter by compute ID
- `--service`: Filter by service ID

### create

Create assignment.

```bash
kubebuddy assignment create \
  --service postgres-db \
  --compute server-01

kubebuddy assignment create \
  --service web-server \
  --compute server-02 \
  --force
```

**Flags:**

- `--service`: Service ID (required)
- `--compute`: Compute ID (required)
- `--force`: Force assignment even if resources insufficient

### delete

Delete assignment.

```bash
kubebuddy assignment delete <id>
```

## plan

Find suitable computes for service placement.

```bash
kubebuddy plan <service-id> [flags]
```

**Flags:**

- `--json`: Output as JSON
- `--compute`: Plan for specific compute
- `--assign`: Create assignment on best candidate
- `--force`: Force assignment with --assign even if resources insufficient

**Example:**

```bash
# Find candidates
kubebuddy plan postgres-db

# Plan for specific compute
kubebuddy plan postgres-db --compute server-01

# Plan and assign to best candidate
kubebuddy plan postgres-db --assign

# Force assign to specific compute
kubebuddy plan postgres-db --compute server-01 --assign --force
```

Output shows:

- Feasibility status
- Candidate computes with scores and utilization
- Hardware resources (total vs allocated)
- Recommendations if not feasible

## journal

Manage per-compute journal entries.

### list

List journal entries.

```bash
kubebuddy journal list
kubebuddy journal list --compute server-01
```

**Flags:**

- `--compute`: Filter by compute ID

### add

Add journal entry.

```bash
kubebuddy journal add \
  --compute server-01 \
  --category maintenance \
  --content "Replaced failed PSU"

kubebuddy journal add \
  --compute server-01 \
  --category deployment \
  --content "Deployed postgres v15.2"
```

**Flags:**

- `--compute`: Compute ID (required)
- `--category`: maintenance, incident, deployment, hardware, network, other (default: other)
- `--content`: Entry content (required)

## report

Generate markdown reports.

### compute

Generate compute report.

```bash
# All computes
kubebuddy report compute

# Specific compute
kubebuddy report compute server-01

# With detailed journal entries
kubebuddy report compute server-01 --journal
```

**Flags:**

- `--journal`: Show detailed journal entries with full content

Output includes:

- Compute metadata (type, provider, region, state, tags)
- Hardware components with specs
- RAID configuration
- Network configuration (IP addresses, DNS records, firewall rules)
- Assigned services with port assignments
- Resource summary (total vs allocated)
- Storage breakdown with RAID arrays
- Journal entries table

**Networking details:**
- IP addresses (type, address, CIDR, gateway, DNS servers, primary status)
- DNS records (only shown when value matches the IP address)
- Port assignments (shown inline with each service: external IP:port → service port)
- Firewall rules assigned to the compute

## apikey

Manage API keys (admin scope required).

### list

List all API keys.

```bash
kubebuddy apikey list
```

### create

Create API key.

```bash
kubebuddy apikey create \
  --name "dev-key" \
  --scope readwrite \
  --description "Development access"
```

**Flags:**

- `--name`: API key name (required)
- `--scope`: admin, readwrite, readonly (default: readonly)
- `--description`: Description

### delete

Delete API key.

```bash
kubebuddy apikey delete <id>
```
