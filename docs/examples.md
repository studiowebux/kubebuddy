---
title: Workflow Examples
description: Common workflows and usage patterns
tags: [examples, workflows, tutorial]
---

# Workflow Examples

## Web Hosting Infrastructure

Build infrastructure with database and web servers.

### Start Server

```bash
export ADMIN_API_KEY=secure-key
kubebuddy server --db ~/kubebuddy.db --create-admin-key
```

### Create Components

```bash
# CPU
CPU_ID=$(kubebuddy component create \
  --name "Intel Xeon E5-2680v4" \
  --type cpu \
  --manufacturer Intel \
  --model "E5-2680v4" \
  --specs '{"threads":28}' | jq -r .id)

# RAM
RAM_ID=$(kubebuddy component create \
  --name "32GB DDR4" \
  --type ram \
  --manufacturer Samsung \
  --model "M393A4K40CB2" \
  --specs '{"capacity_gb":32}' | jq -r .id)

# Storage
NVME_ID=$(kubebuddy component create \
  --name "960GB NVMe" \
  --type storage \
  --manufacturer Samsung \
  --model "PM983" \
  --specs '{"capacity_gb":960}' | jq -r .id)
```

### Create Database Server

```bash
# Create compute with tags
DB_SERVER=$(kubebuddy compute create \
  --name "db-prod-01" \
  --type baremetal \
  --provider ovh \
  --region us-east \
  --tags "env=prod,tier=database" \
  --monthly-cost 299.99 \
  --contract-end 2026-06-30 | jq -r .id)

# Assign hardware
kubebuddy component assign --compute $DB_SERVER --component $CPU_ID --quantity 2
kubebuddy component assign --compute $DB_SERVER --component $RAM_ID --quantity 8
kubebuddy component assign --compute $DB_SERVER --component $NVME_ID --quantity 2 \
  --raid raid1 --raid-group db-storage
```

### Create Web Server

```bash
# Create compute with tags
WEB_SERVER=$(kubebuddy compute create \
  --name "web-prod-01" \
  --type baremetal \
  --provider ovh \
  --region us-east \
  --tags "env=prod,tier=web" | jq -r .id)

# Assign hardware
kubebuddy component assign --compute $WEB_SERVER --component $CPU_ID --quantity 1
kubebuddy component assign --compute $WEB_SERVER --component $RAM_ID --quantity 4
kubebuddy component assign --compute $WEB_SERVER --component $NVME_ID --quantity 1
```

### Define Services

```bash
# Database service with affinity for database tier
kubebuddy service create \
  --name "postgres-db" \
  --min-spec '{"cores":4,"memory":8192,"nvme":100}' \
  --max-spec '{"cores":8,"memory":16384,"nvme":200}' \
  --placement '{"affinity":[{"match_labels":{"tier":"database"}}]}'

# Web service with affinity for web tier
kubebuddy service create \
  --name "nginx-web" \
  --min-spec '{"cores":2,"memory":2048,"nvme":50}' \
  --max-spec '{"cores":4,"memory":4096,"nvme":100}' \
  --placement '{"affinity":[{"match_labels":{"tier":"web"}}]}'
```

### Plan and Assign

```bash
# Plan and assign automatically
kubebuddy plan postgres-db --assign
kubebuddy plan nginx-web --assign
```

### Add Journal Entries

```bash
kubebuddy journal add \
  --compute $DB_SERVER \
  --category deployment \
  --content "Deployed PostgreSQL 15.2 with RAID1 storage"
```

### Generate Reports

```bash
# All computes with journal details
kubebuddy report compute --journal

# Specific compute
kubebuddy report compute $DB_SERVER --journal
```

## Planning Workflow

### Find Candidates

```bash
kubebuddy plan postgres-db
```

Output:
```
Service: postgres-db

✓ Feasible on 1 compute(s)

Candidates:
1. db-prod-01 (score: 95.5)
   - Type: baremetal
   - Provider: ovh
   - Region: us-east
   - Utilization: 65%
   - Available: cores=48, memory=245760, nvme=860
```

### Plan for Specific Compute

```bash
kubebuddy plan postgres-db --compute $DB_SERVER
```

### Auto-Assign to Best Candidate

```bash
kubebuddy plan postgres-db --assign
```

### Force Assignment

```bash
kubebuddy plan postgres-db --compute $DB_SERVER --assign --force
```

## Component Management

### Create and Assign GPU

```bash
# Create GPU component
GPU_ID=$(kubebuddy component create \
  --name "NVIDIA RTX 4090" \
  --type gpu \
  --manufacturer NVIDIA \
  --model "RTX 4090" \
  --specs '{"vram_gb":24,"cuda_cores":16384}' | jq -r .id)

# Assign to compute
kubebuddy component assign \
  --compute $ML_SERVER \
  --component $GPU_ID \
  --quantity 4 \
  --slot "PCIE1,PCIE2,PCIE3,PCIE4"
```

### Storage with RAID5

```bash
# Create storage component
SSD_ID=$(kubebuddy component create \
  --name "Samsung 2TB SSD" \
  --type storage \
  --manufacturer Samsung \
  --model "870 EVO" \
  --specs '{"capacity_gb":2000}' | jq -r .id)

# Assign 4 drives in RAID5
kubebuddy component assign \
  --compute $SERVER \
  --component $SSD_ID \
  --quantity 4 \
  --raid raid5 \
  --raid-group storage-array-1

# Effective capacity: (4-1) * 2000 = 6000 GB
```

## Service with Placement Rules

### Affinity (Must Match)

```bash
kubebuddy service create \
  --name "prod-api" \
  --min-spec '{"cores":2,"memory":4096}' \
  --max-spec '{"cores":4,"memory":8192}' \
  --placement '{"affinity":[{"match_labels":{"env":"prod"}}]}'
```

### Anti-Affinity (Must NOT Match)

```bash
kubebuddy service create \
  --name "critical-service" \
  --min-spec '{"cores":2,"memory":4096}' \
  --max-spec '{"cores":4,"memory":8192}' \
  --placement '{"anti_affinity":[{"match_labels":{"env":"dev"}}]}'
```

### Spread Constraint

```bash
kubebuddy service create \
  --name "redis-cache" \
  --min-spec '{"cores":2,"memory":8192}' \
  --max-spec '{"cores":4,"memory":16384}' \
  --placement '{"spread_max":1}'
```

Only one instance per compute for high availability.
