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
- Assigned services
- Resource summary (total vs allocated)
- Storage breakdown with RAID arrays
- Journal entries table

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
