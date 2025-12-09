---
title: Core Concepts
description: Fundamental concepts and data models
tags: [architecture, concepts, models]
---

# Core Concepts

## Compute

Physical or virtual machines (baremetal, vps, vm) with provider, region, and state.

Attributes:
- **Name**: Unique identifier
- **Type**: baremetal, vps, vm
- **Provider**: Cloud or hosting provider
- **Region**: Geographic location
- **State**: active, inactive, maintenance
- **Tags**: Key-value metadata for placement rules
- **Billing**: Monthly cost, annual cost, contract end date, renewal date (optional)

## Component

Hardware parts (CPU, RAM, GPU, storage, NIC, PSU) with manufacturer, model, and specs.

Component types:
- `cpu`: Processors
- `ram`: Memory modules
- `gpu`: Graphics cards
- `storage`: Disks (HDD, SSD, NVMe)
- `nic`: Network cards
- `psu`: Power supplies

Specs are stored as JSON and vary by component type:
- CPU: `cores`, `threads`, `ghz`
- RAM: `capacity_gb`, `speed_mhz`
- Storage: `capacity_gb`, `type`
- GPU: `vram_gb`, `cuda_cores`

## Component Assignment

Maps components to computes with quantity, slot, serial number, and RAID configuration.

Attributes:
- **Quantity**: Number of identical components
- **Slot**: Physical location (e.g., CPU1, DIMM0-3)
- **Serial Number**: For tracking individual units
- **RAID Level**: For storage (raid0, raid1, raid5, raid6, raid10)
- **RAID Group**: Components with same group form RAID array

## Service

Workloads with min/max resource requirements and placement rules.

Resource keys:
- `cores`: CPU cores
- `memory`: RAM in MB
- `vram`: GPU memory in MB
- `nvme`: Storage in GB
- `gpu`: Number of GPUs

Placement rules:
- **Affinity**: Must match tags
- **Anti-affinity**: Must NOT match tags
- **Spread Max**: Max instances per compute

## Assignment

Allocates services to computes with resource tracking.

Tracks allocated resources per service on each compute. Used to calculate available capacity.

## IP Address

Network addresses assigned to compute resources with CIDR, gateway, and DNS configuration.

Attributes:
- **Address**: IP address (IPv4 or IPv6)
- **Type**: public or private
- **CIDR**: Network CIDR notation (e.g., 192.168.1.0/24)
- **Gateway**: Gateway address
- **DNS Servers**: Comma-separated DNS servers
- **Provider**: Network provider
- **Region**: Geographic location
- **State**: available, assigned, reserved
- **Notes**: Additional information

IP addresses support upsert (create or update by address).

## IP Assignment

Maps IP addresses to compute resources with primary designation.

Attributes:
- **Compute ID**: Target compute
- **IP ID**: Assigned IP address
- **Primary**: Whether this is the primary IP for the compute

Each compute can have multiple IP addresses, but only one primary IP.

## DNS Record

DNS records (A, AAAA, CNAME, PTR) with optional IP linkage.

Attributes:
- **Name**: DNS record name (e.g., www.example.com)
- **Type**: A, AAAA, CNAME, PTR
- **Value**: IP address or hostname
- **Zone**: DNS zone (e.g., example.com)
- **TTL**: Time to live in seconds (default: 3600)
- **IP ID**: Optional link to IP address
- **Notes**: Additional information

DNS records support upsert (create or update by name+type+zone).

## Port Assignment

Maps external ports on IP addresses to internal service ports.

Attributes:
- **Assignment ID**: Links to service-to-compute assignment
- **IP ID**: IP address for the port
- **Port**: External port number
- **Protocol**: tcp, udp, icmp, all
- **Service Port**: Internal service port
- **Description**: Optional description

Port assignments support upsert (create or update by ip+port+protocol).

Example: External 203.0.113.45:8080/tcp → Internal service port 80

## Firewall Rule

Network access policies that can be assigned to computes.

Attributes:
- **Name**: Unique rule identifier
- **Action**: ALLOW or DENY
- **Protocol**: tcp, udp, icmp, all
- **Source**: Source CIDR, IP, or "any"
- **Destination**: Destination CIDR, IP, or "any"
- **Port Start**: Starting port (0 for any)
- **Port End**: Ending port (0 for single port)
- **Priority**: Lower values = higher priority (default: 100)
- **Description**: Optional description

Firewall rules support upsert (create or update by name).

## Firewall Assignment

Maps firewall rules to computes with enable/disable control.

Attributes:
- **Compute ID**: Target compute
- **Rule ID**: Firewall rule
- **Enabled**: Whether rule is active (default: true)

Multiple rules can be assigned to a single compute, evaluated by priority.

## Journal

Audit log per compute for maintenance, incidents, deployments.

Categories:
- `maintenance`: Scheduled maintenance
- `incident`: Outages, failures
- `deployment`: Software deployments
- `hardware`: Hardware changes
- `network`: Network changes
- `other`: Miscellaneous

Each entry includes:
- Timestamp
- Category
- Content (plain text or markdown)
- Created by (API key name)

## Planning

Finds suitable computes for service placement based on available resources.

Process:
1. Filter computes by placement rules
2. Calculate available resources (total - allocated)
3. Check if service fits (between min and max spec)
4. Score candidates by utilization (target 60-70%)
5. Return ranked candidates or purchase recommendations
