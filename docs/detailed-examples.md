---
title: Detailed Examples
description: Comprehensive real-world examples covering all KubeBuddy features
tags: [examples, workflows, tutorial, advanced]
---

# Detailed Examples

Complete walkthroughs demonstrating KubeBuddy's features through realistic scenarios. Each example builds on concepts from getting-started.md and shows production-ready workflows.

## Prerequisites

Start the server with WebUI enabled:

```bash
export KUBEBUDDY_ADMIN_API_KEY="your-secure-key"
kubebuddy server --db ./kubebuddy.db --create-admin-key --webui
```

Access:
- API server: http://localhost:8080
- WebUI: http://localhost:8081

## Example 1: Building a Multi-Server Infrastructure

Complete workflow for provisioning a production environment with multiple servers, shared components, and OS tracking.

### Step 1: Create Compute Resources

```bash
# Production database servers
kubebuddy compute create \
  --name prod-db-01 \
  --type baremetal \
  --provider ovh \
  --region us-east \
  --tags "env=prod,role=database,zone=us-east-1"

kubebuddy compute create \
  --name prod-db-02 \
  --type baremetal \
  --provider ovh \
  --region us-east \
  --tags "env=prod,role=database,zone=us-east-2"

# Web application servers
kubebuddy compute create \
  --name prod-web-01 \
  --type baremetal \
  --provider ovh \
  --region us-east \
  --tags "env=prod,role=web,zone=us-east-1"

kubebuddy compute create \
  --name prod-web-02 \
  --type baremetal \
  --provider ovh \
  --region us-east \
  --tags "env=prod,role=web,zone=us-east-2"
```

### Step 2: Define Hardware Components

```bash
# CPU
kubebuddy component create \
  --name "Intel Xeon E5-2680v4" \
  --type cpu \
  --manufacturer Intel \
  --model E5-2680v4 \
  --specs '{"cores":14,"threads":28,"ghz":2.4,"tdp":120}'

# RAM
kubebuddy component create \
  --name "Samsung 32GB DDR4" \
  --type ram \
  --manufacturer Samsung \
  --model "M393A4K40BB1-CRC" \
  --specs '{"capacity_gb":32,"speed_mhz":2400,"type":"DDR4","ecc":true}'

# NVMe Storage
kubebuddy component create \
  --name "Samsung 960GB NVMe" \
  --type storage \
  --manufacturer Samsung \
  --model PM983 \
  --specs '{"capacity_gb":960,"interface":"nvme","write_iops":35000,"read_iops":180000}'

# SATA Storage
kubebuddy component create \
  --name "Seagate 4TB SATA" \
  --type storage \
  --manufacturer Seagate \
  --model ST4000NM0035 \
  --specs '{"capacity_gb":4000,"interface":"sata","rpm":7200}'

# Network Interface
kubebuddy component create \
  --name "Intel X710 10GbE" \
  --type nic \
  --manufacturer Intel \
  --model X710 \
  --specs '{"speed_gbps":10,"ports":2,"type":"fiber"}'

# Operating System
kubebuddy component create \
  --name "Ubuntu 22.04 LTS" \
  --type os \
  --manufacturer Canonical \
  --model "22.04" \
  --specs '{"distro":"ubuntu","version":"22.04","kernel":"6.5","support_until":"2027-04"}'
```

### Step 3: Assign Components (Multi-Machine)

Assign the same CPU to all servers:

```bash
kubebuddy component assign \
  --computes prod-db-01,prod-db-02,prod-web-01,prod-web-02 \
  --component "Intel Xeon E5-2680v4" \
  --quantity 2
```

Assign RAM to all servers:

```bash
kubebuddy component assign \
  --computes prod-db-01,prod-db-02,prod-web-01,prod-web-02 \
  --component "Samsung 32GB DDR4" \
  --quantity 8
```

Database servers: NVMe RAID1 boot + SATA RAID10 data:

```bash
# Boot drives (RAID1 using numeric format)
kubebuddy component assign \
  --computes prod-db-01,prod-db-02 \
  --component "Samsung 960GB NVMe" \
  --quantity 2 \
  --raid 1 \
  --raid-group boot-array \
  --notes "Boot drive - RAID1 mirror for reliability"

# Data drives (RAID10 using string format)
kubebuddy component assign \
  --computes prod-db-01,prod-db-02 \
  --component "Seagate 4TB SATA" \
  --quantity 4 \
  --raid raid10 \
  --raid-group data-array \
  --notes "Database storage - RAID10 for performance and redundancy"
```

Web servers: NVMe RAID1 only:

```bash
kubebuddy component assign \
  --computes prod-web-01,prod-web-02 \
  --component "Samsung 960GB NVMe" \
  --quantity 2 \
  --raid 1 \
  --raid-group system-array \
  --notes "System and application storage"
```

Network cards to all servers:

```bash
kubebuddy component assign \
  --computes prod-db-01,prod-db-02,prod-web-01,prod-web-02 \
  --component "Intel X710 10GbE" \
  --quantity 1
```

OS to all servers:

```bash
kubebuddy component assign \
  --computes prod-db-01,prod-db-02,prod-web-01,prod-web-02 \
  --component "Ubuntu 22.04 LTS" \
  --quantity 1
```

### Step 4: Verify Hardware Configuration

Check specific server:

```bash
kubebuddy component list-assignments --compute prod-db-01
```

Check all assignments for a component:

```bash
kubebuddy component list-assignments --component "Samsung 960GB NVMe"
```

Generate comprehensive report:

```bash
kubebuddy report compute prod-db-01
```

## Example 2: Service Planning with Placement Rules

Define services with resource requirements and affinity rules.

### Step 1: Create Services

PostgreSQL with production-only placement:

```bash
kubebuddy service create \
  --name postgres-primary \
  --min-spec '{
    "cores": 4,
    "memory": 16384,
    "nvme": 200,
    "sata": 1000
  }' \
  --max-spec '{
    "cores": 8,
    "memory": 32768,
    "nvme": 500,
    "sata": 4000
  }' \
  --placement '{
    "affinity": [
      {
        "matchExpressions": [
          {"key": "role", "operator": "In", "values": ["database"]},
          {"key": "env", "operator": "In", "values": ["prod"]}
        ]
      }
    ],
    "anti_affinity": [
      {
        "matchExpressions": [
          {"key": "env", "operator": "In", "values": ["dev", "staging"]}
        ]
      }
    ]
  }'
```

Web application with zone anti-affinity:

```bash
kubebuddy service create \
  --name web-app \
  --min-spec '{
    "cores": 2,
    "memory": 8192,
    "nvme": 100
  }' \
  --max-spec '{
    "cores": 4,
    "memory": 16384,
    "nvme": 200
  }' \
  --placement '{
    "affinity": [
      {
        "matchExpressions": [
          {"key": "role", "operator": "In", "values": ["web"]},
          {"key": "env", "operator": "In", "values": ["prod"]}
        ]
      }
    ]
  }'
```

### Step 2: Run Planning

Find suitable hosts for postgres-primary:

```bash
kubebuddy plan postgres-primary
```

The planner evaluates each compute resource:
- Checks affinity rules (must have role=database AND env=prod)
- Checks anti-affinity rules (must NOT have env=dev or env=staging)
- Verifies resource availability
- Returns ranked list of suitable hosts

### Step 3: Assign Services

Assign to best-fit compute:

```bash
kubebuddy assignment create \
  --compute prod-db-01 \
  --service postgres-primary \
  --allocated-spec '{
    "cores": 6,
    "memory": 24576,
    "nvme": 300,
    "sata": 2000
  }'
```

Assign web-app to both web servers:

```bash
kubebuddy assignment create \
  --compute prod-web-01 \
  --service web-app \
  --allocated-spec '{"cores": 2, "memory": 8192, "nvme": 150}'

kubebuddy assignment create \
  --compute prod-web-02 \
  --service web-app \
  --allocated-spec '{"cores": 2, "memory": 8192, "nvme": 150}'
```

### Step 4: Verify Assignments

List all assignments for a service:

```bash
kubebuddy assignment list --service postgres-primary
```

List all services on a compute:

```bash
kubebuddy assignment list --compute prod-db-01
```

## Example 3: Complete Network Configuration

Set up networking with IPs, DNS, firewall rules, and port mappings.

### Step 1: Create IP Addresses

Public IPs for web servers:

```bash
kubebuddy ip create \
  --address 203.0.113.10 \
  --type public \
  --cidr "203.0.113.0/24" \
  --gateway "203.0.113.1"

kubebuddy ip create \
  --address 203.0.113.11 \
  --type public \
  --cidr "203.0.113.0/24" \
  --gateway "203.0.113.1"
```

Private IPs for database servers:

```bash
kubebuddy ip create \
  --address 10.0.1.10 \
  --type private \
  --cidr "10.0.1.0/24" \
  --gateway "10.0.1.1"

kubebuddy ip create \
  --address 10.0.1.11 \
  --type private \
  --cidr "10.0.1.0/24" \
  --gateway "10.0.1.1"
```

### Step 2: Assign IPs to Computes

Using name resolution:

```bash
# Get IP IDs first
WEB1_IP=$(kubebuddy ip list | grep "203.0.113.10" | awk '{print $1}')
WEB2_IP=$(kubebuddy ip list | grep "203.0.113.11" | awk '{print $1}')
DB1_IP=$(kubebuddy ip list | grep "10.0.1.10" | awk '{print $1}')
DB2_IP=$(kubebuddy ip list | grep "10.0.1.11" | awk '{print $1}')

# Assign IPs
kubebuddy ip assign --compute prod-web-01 --ip $WEB1_IP
kubebuddy ip assign --compute prod-web-02 --ip $WEB2_IP
kubebuddy ip assign --compute prod-db-01 --ip $DB1_IP
kubebuddy ip assign --compute prod-db-02 --ip $DB2_IP
```

Verify IP assignments:

```bash
kubebuddy ip list-assignments --compute prod-web-01
```

### Step 3: Configure DNS Records

A records for web servers:

```bash
kubebuddy dns create \
  --zone example.com \
  --name www \
  --type A \
  --value 203.0.113.10 \
  --ttl 300

kubebuddy dns create \
  --zone example.com \
  --name www \
  --type A \
  --value 203.0.113.11 \
  --ttl 300
```

CNAME for alias:

```bash
kubebuddy dns create \
  --zone example.com \
  --name app \
  --type CNAME \
  --value www.example.com \
  --ttl 300
```

### Step 4: Create Firewall Rules

Allow SSH from management network:

```bash
kubebuddy firewall create \
  --name allow-ssh-mgmt \
  --action ALLOW \
  --protocol tcp \
  --port-start 22 \
  --port-end 22 \
  --source "10.0.0.0/24"
```

Allow HTTP/HTTPS to web servers:

```bash
kubebuddy firewall create \
  --name allow-http \
  --action ALLOW \
  --protocol tcp \
  --port-start 80 \
  --port-end 80 \
  --source "0.0.0.0/0"

kubebuddy firewall create \
  --name allow-https \
  --action ALLOW \
  --protocol tcp \
  --port-start 443 \
  --port-end 443 \
  --source "0.0.0.0/0"
```

Allow PostgreSQL between web and database servers:

```bash
kubebuddy firewall create \
  --name allow-postgres \
  --action ALLOW \
  --protocol tcp \
  --port-start 5432 \
  --port-end 5432 \
  --source "10.0.1.0/24"
```

Deny all other traffic:

```bash
kubebuddy firewall create \
  --name deny-all \
  --action DENY \
  --protocol tcp \
  --port-start 0 \
  --port-end 65535 \
  --source "0.0.0.0/0"
```

### Step 5: Assign Firewall Rules

Get rule IDs and assign to computes:

```bash
SSH_RULE=$(kubebuddy firewall list | grep "allow-ssh-mgmt" | awk '{print $1}')
HTTP_RULE=$(kubebuddy firewall list | grep "allow-http" | awk '{print $1}')
HTTPS_RULE=$(kubebuddy firewall list | grep "allow-https" | awk '{print $1}')
POSTGRES_RULE=$(kubebuddy firewall list | grep "allow-postgres" | awk '{print $1}')
DENY_RULE=$(kubebuddy firewall list | grep "deny-all" | awk '{print $1}')

# All servers get SSH access
kubebuddy firewall assign --compute prod-web-01 --rule $SSH_RULE
kubebuddy firewall assign --compute prod-web-02 --rule $SSH_RULE
kubebuddy firewall assign --compute prod-db-01 --rule $SSH_RULE
kubebuddy firewall assign --compute prod-db-02 --rule $SSH_RULE

# Web servers get HTTP/HTTPS
kubebuddy firewall assign --compute prod-web-01 --rule $HTTP_RULE
kubebuddy firewall assign --compute prod-web-01 --rule $HTTPS_RULE
kubebuddy firewall assign --compute prod-web-02 --rule $HTTP_RULE
kubebuddy firewall assign --compute prod-web-02 --rule $HTTPS_RULE

# Database servers get PostgreSQL
kubebuddy firewall assign --compute prod-db-01 --rule $POSTGRES_RULE
kubebuddy firewall assign --compute prod-db-02 --rule $POSTGRES_RULE
```

### Step 6: Create Port Mappings

Get service IDs:

```bash
WEB_SERVICE=$(kubebuddy service list | grep "web-app" | awk '{print $1}')
```

Map external ports to service:

```bash
kubebuddy port create \
  --service $WEB_SERVICE \
  --external-port 80 \
  --service-port 8080 \
  --protocol tcp

kubebuddy port create \
  --service $WEB_SERVICE \
  --external-port 443 \
  --service-port 8443 \
  --protocol tcp
```

## Example 4: Journal and Change Tracking

Document maintenance, deployments, and incidents.

### Step 1: Record Maintenance Activities

System update on database server:

```bash
kubebuddy journal add \
  --compute prod-db-01 \
  --category maintenance \
  --content "Applied kernel security updates (6.5.0-28). Rebooted at 02:00 UTC. All services resumed normally."
```

### Step 2: Track Deployments

Application deployment:

```bash
kubebuddy journal add \
  --compute prod-web-01 \
  --category deployment \
  --content "Deployed web-app v2.4.1. Updated configuration for new API endpoint. Zero downtime deployment completed."

kubebuddy journal add \
  --compute prod-web-02 \
  --category deployment \
  --content "Deployed web-app v2.4.1. Synchronized with prod-web-01."
```

### Step 3: Document Incidents

Database issue:

```bash
kubebuddy journal add \
  --compute prod-db-01 \
  --category incident \
  --content "HIGH: Disk latency spike detected at 14:35 UTC. Investigation revealed RAID rebuild in progress after drive replacement. Resolved at 16:20 UTC."
```

### Step 4: Record Configuration Changes

Hardware upgrade:

```bash
kubebuddy journal add \
  --compute prod-web-01 \
  --category change \
  --content "Added 2x Samsung 32GB DDR4 modules. Total RAM now 256GB. Validated with memtest86+."
```

### Step 5: Query Journal

View all entries for a compute:

```bash
kubebuddy journal list --compute prod-db-01
```

Filter by category:

```bash
kubebuddy journal list --compute prod-db-01 --category incident
```

## Example 5: WebUI Workflows

Complete workflows using the web interface.

### Scenario: Provision New Server via WebUI

1. Navigate to http://localhost:8081

2. **Create Compute** (Computes tab):
   - Click "+ Add Compute"
   - Name: staging-app-01
   - Type: baremetal
   - Provider: ovh
   - Region: us-west
   - Tags: env=staging,role=application
   - Click "Create"

3. **Assign Components** (Components tab):
   - Select existing components from dropdown
   - Or create new ones with "+ Add Component"

4. **View Report** (Reports tab):
   - Select "staging-app-01" from dropdown
   - Review comprehensive report including:
     - Compute information
     - Hardware components
     - Service assignments
     - IP assignments
     - Journal entries

5. **Add Journal Entry** (Journal tab):
   - Click "+ Add Journal Entry"
   - Compute: staging-app-01
   - Category: deployment
   - Content: "Initial server provisioning completed"
   - Click "Create"

### Scenario: Network Configuration via WebUI

1. **Create IPs** (IPs tab):
   - Click "+ Add IP"
   - Address: 192.168.10.100
   - Type: private
   - CIDR: 192.168.10.0/24
   - Gateway: 192.168.10.1
   - Click "Create"

2. **Create DNS Records** (DNS tab):
   - Click "+ Add DNS Record"
   - Zone: staging.internal
   - Name: app-01
   - Type: A
   - Value: 192.168.10.100
   - TTL: 300
   - Click "Create"

3. **Create Firewall Rules** (Firewall tab):
   - Click "+ Add Firewall Rule"
   - Name: allow-app-traffic
   - Action: ALLOW
   - Protocol: TCP
   - Port Start: 8080
   - Port End: 8080
   - Source: 192.168.0.0/16
   - Click "Create"

## Example 6: API Integration

Use KubeBuddy's REST API for automation.

### Authentication

All requests require API key:

```bash
export KUBEBUDDY_API_KEY="your-api-key"
```

### Create Compute via API

```bash
curl -X POST http://localhost:8080/api/computes \
  -H "Authorization: Bearer $KUBEBUDDY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "api-server-01",
    "compute_type": "baremetal",
    "provider": "ovh",
    "region": "eu-west",
    "tags": ["env=prod", "role=api"]
  }'
```

### List All Computes

```bash
curl http://localhost:8080/api/computes \
  -H "Authorization: Bearer $KUBEBUDDY_API_KEY"
```

### Get Compute Report

```bash
curl http://localhost:8080/api/reports/compute/api-server-01 \
  -H "Authorization: Bearer $KUBEBUDDY_API_KEY"
```

### Create Component Assignment

```bash
curl -X POST http://localhost:8080/api/components/assignments \
  -H "Authorization: Bearer $KUBEBUDDY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "compute_id": "compute-uuid-here",
    "component_id": "component-uuid-here",
    "quantity": 2,
    "raid_level": "raid1",
    "raid_group": "system",
    "notes": "Boot drives"
  }'
```

## Example 7: Advanced RAID Configurations

Different RAID levels for different workloads.

### RAID0 - Maximum Performance (Striping)

High-performance scratch space:

```bash
kubebuddy component assign \
  --computes ml-worker-01 \
  --component "Samsung 960GB NVMe" \
  --quantity 4 \
  --raid 0 \
  --raid-group scratch-array \
  --notes "Temporary ML training data - RAID0 for maximum throughput"
```

### RAID1 - Maximum Reliability (Mirroring)

Critical system drives:

```bash
kubebuddy component assign \
  --computes prod-db-01 \
  --component "Samsung 960GB NVMe" \
  --quantity 2 \
  --raid raid1 \
  --raid-group os-boot \
  --notes "Operating system and boot - RAID1 mirror"
```

### RAID5 - Balanced (Striping with Parity)

Large storage with single-drive failure tolerance:

```bash
kubebuddy component assign \
  --computes backup-server-01 \
  --component "Seagate 4TB SATA" \
  --quantity 5 \
  --raid 5 \
  --raid-group backup-pool \
  --notes "Backup storage - RAID5 for capacity and redundancy"
```

### RAID6 - Double Parity

Critical data with two-drive failure tolerance:

```bash
kubebuddy component assign \
  --computes archive-server-01 \
  --component "Seagate 4TB SATA" \
  --quantity 8 \
  --raid raid6 \
  --raid-group archive-pool \
  --notes "Long-term archive - RAID6 for maximum data protection"
```

### RAID10 - Performance + Reliability

Database storage requiring both speed and redundancy:

```bash
kubebuddy component assign \
  --computes prod-db-01,prod-db-02 \
  --component "Seagate 4TB SATA" \
  --quantity 4 \
  --raid 10 \
  --raid-group database-data \
  --notes "PostgreSQL data directory - RAID10 for performance and redundancy"
```

## Example 8: Multi-Role Server Configuration

Single server hosting multiple services with proper resource allocation.

### Create Multi-Role Compute

```bash
kubebuddy compute create \
  --name all-in-one-01 \
  --type baremetal \
  --provider ovh \
  --region us-central \
  --tags "env=dev,roles=web+db+cache"
```

### Assign Generous Hardware

```bash
# 2x CPU
kubebuddy component assign \
  --computes all-in-one-01 \
  --component "Intel Xeon E5-2680v4" \
  --quantity 2

# 512GB RAM
kubebuddy component assign \
  --computes all-in-one-01 \
  --component "Samsung 32GB DDR4" \
  --quantity 16

# Multiple storage tiers
kubebuddy component assign \
  --computes all-in-one-01 \
  --component "Samsung 960GB NVMe" \
  --quantity 2 \
  --raid 1 \
  --raid-group system

kubebuddy component assign \
  --computes all-in-one-01 \
  --component "Seagate 4TB SATA" \
  --quantity 4 \
  --raid 10 \
  --raid-group data
```

### Create Services

```bash
# Web service
kubebuddy service create \
  --name nginx-web \
  --min-spec '{"cores":2,"memory":4096,"nvme":50}'

# Database
kubebuddy service create \
  --name postgres-db \
  --min-spec '{"cores":4,"memory":16384,"nvme":100,"sata":500}'

# Cache
kubebuddy service create \
  --name redis-cache \
  --min-spec '{"cores":2,"memory":8192,"nvme":100}'
```

### Assign Services with Resource Partitioning

```bash
kubebuddy assignment create \
  --compute all-in-one-01 \
  --service nginx-web \
  --allocated-spec '{"cores":4,"memory":8192,"nvme":100}'

kubebuddy assignment create \
  --compute all-in-one-01 \
  --service postgres-db \
  --allocated-spec '{"cores":8,"memory":32768,"nvme":200,"sata":2000}'

kubebuddy assignment create \
  --compute all-in-one-01 \
  --service redis-cache \
  --allocated-spec '{"cores":4,"memory":16384,"nvme":150}'
```

### Verify Resource Allocation

```bash
kubebuddy report compute all-in-one-01
```

Check that total allocated resources don't exceed hardware capacity.

## Troubleshooting Common Scenarios

### Scenario: Service Won't Plan

**Problem**: `kubebuddy plan my-service` returns no candidates

**Checks**:

1. Verify placement rules match compute tags:
   ```bash
   kubebuddy service list | grep my-service
   kubebuddy compute list | grep -A5 "Tags"
   ```

2. Check resource availability:
   ```bash
   kubebuddy report compute candidate-server
   ```

3. Test without placement rules to isolate issue

### Scenario: Component Assignment Fails

**Problem**: Multi-machine component assignment shows errors

**Resolution**:

The command returns JSON with successes and errors arrays. Example:

```json
{
  "successes": [
    {"compute": "server1", "status": "assigned"},
    {"compute": "server2", "status": "assigned"}
  ],
  "errors": [
    {"compute": "server3", "error": "compute not found"}
  ]
}
```

Fix the failed computes individually.

### Scenario: WebUI Shows 404

**Problem**: WebUI requests return 404 errors

**Checks**:

1. Verify server started with `--webui` flag:
   ```bash
   ps aux | grep kubebuddy
   ```

2. Check correct ports:
   - API: 8080
   - WebUI: 8081

3. Verify API key is set:
   ```bash
   echo $KUBEBUDDY_ADMIN_API_KEY
   ```

4. Check server logs for errors

## Performance Testing

### Stress Test: Large Scale Provisioning

Create 100 computes:

```bash
for i in {1..100}; do
  kubebuddy compute create \
    --name "server-$(printf %03d $i)" \
    --type baremetal \
    --provider ovh \
    --region us-east \
    --tags "env=test,batch=stress-test"
done
```

Assign components to all 100:

```bash
COMPUTE_LIST=$(kubebuddy compute list | grep "server-" | awk '{print $2}' | tr '\n' ',' | sed 's/,$//')

kubebuddy component assign \
  --computes $COMPUTE_LIST \
  --component "Intel Xeon E5-2680v4" \
  --quantity 2
```

Test WebUI responsiveness with large datasets.

## Integration Examples

### Terraform Integration

Use KubeBuddy to track Terraform-managed infrastructure:

```bash
# After terraform apply
INSTANCE_NAME=$(terraform output instance_name)
INSTANCE_IP=$(terraform output instance_ip)

kubebuddy compute create \
  --name $INSTANCE_NAME \
  --type cloud \
  --provider aws \
  --region us-east-1

kubebuddy journal add \
  --compute $INSTANCE_NAME \
  --category deployment \
  --content "Terraform managed instance created. Instance ID: $(terraform output instance_id)"
```

### Ansible Integration

Track Ansible playbook runs:

```yaml
- name: Record deployment in KubeBuddy
  shell: |
    kubebuddy journal add \
      --compute {{ inventory_hostname }} \
      --category deployment \
      --content "Ansible playbook {{ playbook_name }} executed successfully"
  delegate_to: localhost
```

### Monitoring Integration

Export metrics for monitoring:

```bash
# Custom script to export capacity metrics
kubebuddy report compute prod-db-01 | \
  jq '.service_assignments[] | .allocated_spec' | \
  # Process and send to monitoring system
```

## Summary

These examples demonstrate:

- Multi-server provisioning with shared components
- RAID configuration (numeric and string formats)
- Service planning with placement rules
- Complete network setup (IPs, DNS, firewall, ports)
- Journal tracking for changes and incidents
- WebUI workflows for all operations
- REST API integration for automation
- Advanced RAID configurations for different workloads
- Multi-role server resource partitioning
- Troubleshooting common issues
- Performance testing scenarios
- Integration with infrastructure tools

For additional examples, see:
- examples.md - Basic workflow examples
- multi-role-example.md - Detailed multi-role server scenarios
- commands.md - Complete command reference
- networking.md - Advanced networking configuration
- raid.md - RAID configuration guide
