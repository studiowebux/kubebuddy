# KubeBuddy

Capacity planning system for compute resources and services with intelligent
placement rules.

## Features

- Compute resource management (baremetal, VPS, VM)
- Hardware component catalog with RAID support (numeric and string formats)
- Service definitions with resource specifications
- Tag-based placement constraints
- Network management (IP, DNS, ports, firewall)
- Capacity planning and reporting
- Per-compute journal system
- Multi-scope API key authentication
- Web UI with light/dark theme
- Name resolution (use names instead of UUIDs in CLI commands)
- Multi-machine component assignment
- OS component type tracking
- Resource summary with utilization percentages
- RAID capacity calculation

## Installation

```bash
go build -o kubebuddy ./cmd/kubebuddy
```

Optional: Install to PATH

```bash
sudo cp kubebuddy /usr/local/bin/
```

## Server

### Start Server

Basic:

```bash
KUBEBUDDY_ADMIN_API_KEY=your-secret-key kubebuddy server --create-admin-key
```

With WebUI:

```bash
KUBEBUDDY_ADMIN_API_KEY=your-secret-key kubebuddy server --create-admin-key --webui
```

Access:

- API server: http://localhost:8080
- WebUI: http://localhost:8081

### Flags

| Flag                 | Type   | Default        | Description                                             |
| -------------------- | ------ | -------------- | ------------------------------------------------------- |
| `--db`               | string | `kubebuddy.db` | Database file path (supports `~` expansion)             |
| `--port`             | string | `8080`         | Server port                                             |
| `--webui`            | bool   | `false`        | Enable WebUI server                                     |
| `--webui-port`       | string | `8081`         | WebUI port                                              |
| `--create-admin-key` | bool   | `false`        | Create admin API key from `KUBEBUDDY_ADMIN_API_KEY` env |
| `--seed`             | bool   | `false`        | Populate with sample data                               |

### Environment Variables

| Variable                     | Required                        | Description                            |
| ---------------------------- | ------------------------------- | -------------------------------------- |
| `KUBEBUDDY_ADMIN_API_KEY`    | When using `--create-admin-key` | Admin API key value                    |
| `KUBEBUDDY_PORT`             | No                              | Server port (overridden by `--port`)   |
| `KUBEBUDDY_DB`               | No                              | Database path (overridden by `--db`)   |
| `KUBEBUDDY_CREATE_ADMIN_KEY` | No                              | Set to `true` to create admin key      |
| `KUBEBUDDY_SEED`             | No                              | Set to `true` to seed database on boot |

### Examples

Development with sample data:

```bash
KUBEBUDDY_ADMIN_API_KEY=test123 kubebuddy server --create-admin-key --seed
```

Custom database location:

```bash
KUBEBUDDY_ADMIN_API_KEY=secret kubebuddy server --create-admin-key --db ~/.kubebuddy/
```

Using environment variables:

```bash
export KUBEBUDDY_PORT=3000
export KUBEBUDDY_DB=~/data/kubebuddy.db
export KUBEBUDDY_CREATE_ADMIN_KEY=true
export KUBEBUDDY_SEED=true
export KUBEBUDDY_ADMIN_API_KEY=secret
kubebuddy server
```

## CLI

### Global Flags

| Flag              | Type   | Default                 | Description                            |
| ----------------- | ------ | ----------------------- | -------------------------------------- |
| `--endpoint`      | string | `http://localhost:8080` | API endpoint (or `KUBEBUDDY_ENDPOINT`) |
| `--api-key`       | string |                         | API key (or `KUBEBUDDY_API_KEY`)       |
| `--version`, `-v` | bool   |                         | Show version                           |

### Environment Variables

| Variable             | Default                 | Description                |
| -------------------- | ----------------------- | -------------------------- |
| `KUBEBUDDY_ENDPOINT` | `http://localhost:8080` | API endpoint               |
| `KUBEBUDDY_API_KEY`  |                         | API key for authentication |

### Commands

#### Compute Management

```bash
kubebuddy compute list
kubebuddy compute get <id>
kubebuddy compute create --name server-01 --type baremetal --provider ovh --region eu
kubebuddy compute update <id> --tags "env=prod,tier=app"
kubebuddy compute delete <id>
```

#### Service Management

```bash
kubebuddy service list
kubebuddy service get <id>
kubebuddy service create --name nginx --min-cpu 1 --min-memory 512 --max-cpu 2 --max-memory 1024
kubebuddy service delete <id>
```

#### Assignments

```bash
kubebuddy assignment list
kubebuddy assignment create --service <service-id> --compute <compute-id>
kubebuddy assignment delete <id>
```

#### Component Management

```bash
kubebuddy component list
kubebuddy component create --name "Intel Xeon" --type cpu --manufacturer Intel --model "E5-2680v4" --specs '{"cores":14,"threads":28}'

# Single machine assignment
kubebuddy component assign --computes server-01 --component "Intel Xeon" --quantity 2 --raid raid5 --raid-group rg1

# Multi-machine assignment (assign to multiple servers at once)
kubebuddy component assign --computes server-01,server-02,server-03 --component "32GB RAM" --quantity 8

# RAID supports both numeric and string formats
kubebuddy component assign --computes server-01 --component "Samsung NVMe" --quantity 2 --raid 1 --raid-group boot
kubebuddy component assign --computes server-01 --component "Seagate SATA" --quantity 4 --raid raid10 --raid-group data

# With installation notes
kubebuddy component assign --computes server-01 --component "Samsung NVMe" --notes "Boot drive - RAID1 mirror"

kubebuddy component unassign <assignment-id>
```

Supported component types: `cpu`, `ram`, `storage`, `gpu`, `nic`, `psu`, `os`,
`other`

#### Network Management

IP addresses:

```bash
kubebuddy ip list
kubebuddy ip create --address 192.168.1.100 --type private --cidr 192.168.1.0/24 --gateway 192.168.1.1
kubebuddy ip assign --compute <compute-id> --ip <ip-id> --primary
kubebuddy ip unassign <assignment-id>
```

DNS records:

```bash
kubebuddy dns list --zone example.com
kubebuddy dns create --name example.com --type A --value 1.2.3.4 --zone example.com --ttl 3600
kubebuddy dns delete <id>
```

Port assignments:

```bash
kubebuddy port list --assignment <assignment-id>
kubebuddy port create --assignment <assignment-id> --ip <ip-id> --port 8080 --protocol tcp --service-port 80
kubebuddy port delete <id>
```

Firewall rules:

```bash
kubebuddy firewall list
kubebuddy firewall create --name allow-ssh --action ALLOW --protocol tcp --source any --destination any --port-start 22 --port-end 22
kubebuddy firewall assign --compute <compute-id> --rule <rule-id>
kubebuddy firewall unassign <assignment-id>
```

#### Capacity Planning

```bash
kubebuddy plan <service-id>
```

#### Reports

```bash
# Using name or ID
kubebuddy report compute server-01
kubebuddy report compute <compute-id>

# All computes
kubebuddy report compute

# With detailed journal entries
kubebuddy report compute server-01 --journal
```

Reports include:

- Compute information
- Resource summary (cores, RAM, VRAM, storage with utilization %)
- Storage configuration breakdown (RAID groups with effective capacity)
- Hardware components with specs
- Service assignments
- Network configuration (IPs, DNS, firewall rules, port mappings)
- Journal entries

#### Journal

```bash
kubebuddy journal list --compute <compute-id>
kubebuddy journal add --compute <compute-id> --category maintenance --content "Kernel upgrade"
```

#### API Keys

```bash
kubebuddy apikey list
kubebuddy apikey create --name developer --scope readwrite --description "Dev team key"
kubebuddy apikey delete <id>
```

**Scopes**: `admin`, `readwrite`, `readonly`

## WebUI

Access the web interface at http://localhost:8081 when server is started with
`--webui` flag.

### Features

- Light/dark theme with persistence
- Full CRUD operations for all resources:
  - Computes (create, list, delete)
  - Components (create, list, delete)
  - Services (create, list, delete)
  - Assignments (list with resolved names)
  - IPs (create, list, delete)
  - DNS Records (create, list, delete)
  - Firewall Rules (create, list, delete)
  - Port Mappings (create, list, delete)
  - Journal Entries (create, list)
  - API Keys (create, list, delete)
- Comprehensive Reports:
  - Select compute from dropdown
  - Resource summary with totals and utilization percentages
  - Storage configuration with RAID breakdown
  - Hardware components with specs
  - Service assignments
  - IP assignments (resolved addresses)
  - Journal entries with proper date formatting
- Name resolution (shows names instead of UUIDs)
- Responsive tables with horizontal scrolling
- No build step required (vanilla JavaScript)
- Single binary deployment

### Example

```bash
# Start server with WebUI
export KUBEBUDDY_ADMIN_API_KEY="your-secret-key"
kubebuddy server --db ./kubebuddy.db --create-admin-key --webui

# Access WebUI at http://localhost:8081
# API server at http://localhost:8080
```

## Shell Completion

### Bash

```bash
source <(kubebuddy completion bash)
```

Install permanently:

```bash
kubebuddy completion bash | sudo tee /etc/bash_completion.d/kubebuddy
```

### Zsh

```bash
source <(kubebuddy completion zsh)
```

Install permanently:

```bash
kubebuddy completion zsh > ~/.zsh/completions/_kubebuddy
```

### Fish

```bash
kubebuddy completion fish > ~/.config/fish/completions/kubebuddy.fish
```

## API Endpoints

### Computes

| Method | Endpoint               | Description    |
| ------ | ---------------------- | -------------- |
| GET    | `/api/v1/computes`     | List computes  |
| GET    | `/api/v1/computes/:id` | Get compute    |
| POST   | `/api/v1/computes`     | Create compute |
| PUT    | `/api/v1/computes/:id` | Update compute |
| DELETE | `/api/v1/computes/:id` | Delete compute |

### Services

| Method | Endpoint               | Description    |
| ------ | ---------------------- | -------------- |
| GET    | `/api/v1/services`     | List services  |
| GET    | `/api/v1/services/:id` | Get service    |
| POST   | `/api/v1/services`     | Create service |
| PUT    | `/api/v1/services/:id` | Update service |
| DELETE | `/api/v1/services/:id` | Delete service |

### Assignments

| Method | Endpoint                  | Description       |
| ------ | ------------------------- | ----------------- |
| GET    | `/api/v1/assignments`     | List assignments  |
| POST   | `/api/v1/assignments`     | Create assignment |
| DELETE | `/api/v1/assignments/:id` | Delete assignment |

### Components

| Method | Endpoint                 | Description      |
| ------ | ------------------------ | ---------------- |
| GET    | `/api/v1/components`     | List components  |
| GET    | `/api/v1/components/:id` | Get component    |
| POST   | `/api/v1/components`     | Create component |
| PUT    | `/api/v1/components/:id` | Update component |
| DELETE | `/api/v1/components/:id` | Delete component |

### Component Assignments

| Method | Endpoint                                          | Description        |
| ------ | ------------------------------------------------- | ------------------ |
| GET    | `/api/v1/component-assignments?compute_id=uuid`   | List by compute    |
| GET    | `/api/v1/component-assignments?component_id=uuid` | List by component  |
| POST   | `/api/v1/component-assignments`                   | Assign component   |
| DELETE | `/api/v1/component-assignments/:id`               | Unassign component |

### IP Addresses

| Method | Endpoint          | Description |
| ------ | ----------------- | ----------- |
| GET    | `/api/v1/ips`     | List IPs    |
| GET    | `/api/v1/ips/:id` | Get IP      |
| POST   | `/api/v1/ips`     | Create IP   |
| PUT    | `/api/v1/ips/:id` | Update IP   |
| DELETE | `/api/v1/ips/:id` | Delete IP   |

### IP Assignments

| Method | Endpoint                              | Description     |
| ------ | ------------------------------------- | --------------- |
| GET    | `/api/v1/compute-ips?compute_id=uuid` | List by compute |
| GET    | `/api/v1/compute-ips?ip_id=uuid`      | List by IP      |
| POST   | `/api/v1/compute-ips`                 | Assign IP       |
| DELETE | `/api/v1/compute-ips/:id`             | Unassign IP     |

### DNS Records

| Method | Endpoint          | Description       |
| ------ | ----------------- | ----------------- |
| GET    | `/api/v1/dns`     | List DNS records  |
| GET    | `/api/v1/dns/:id` | Get DNS record    |
| POST   | `/api/v1/dns`     | Create DNS record |
| PUT    | `/api/v1/dns/:id` | Update DNS record |
| DELETE | `/api/v1/dns/:id` | Delete DNS record |

### Port Assignments

| Method | Endpoint            | Description            |
| ------ | ------------------- | ---------------------- |
| GET    | `/api/v1/ports`     | List port assignments  |
| GET    | `/api/v1/ports/:id` | Get port assignment    |
| POST   | `/api/v1/ports`     | Create port assignment |
| PUT    | `/api/v1/ports/:id` | Update port assignment |
| DELETE | `/api/v1/ports/:id` | Delete port assignment |

### Firewall Rules

| Method | Endpoint               | Description          |
| ------ | ---------------------- | -------------------- |
| GET    | `/api/v1/firewall`     | List firewall rules  |
| GET    | `/api/v1/firewall/:id` | Get firewall rule    |
| POST   | `/api/v1/firewall`     | Create firewall rule |
| PUT    | `/api/v1/firewall/:id` | Update firewall rule |
| DELETE | `/api/v1/firewall/:id` | Delete firewall rule |

### Firewall Assignments

| Method | Endpoint                                   | Description            |
| ------ | ------------------------------------------ | ---------------------- |
| GET    | `/api/v1/compute-firewall?compute_id=uuid` | List by compute        |
| GET    | `/api/v1/compute-firewall?rule_id=uuid`    | List by rule           |
| POST   | `/api/v1/compute-firewall`                 | Assign firewall rule   |
| DELETE | `/api/v1/compute-firewall/:id`             | Unassign firewall rule |

### Journal

| Method | Endpoint          | Description          |
| ------ | ----------------- | -------------------- |
| GET    | `/api/v1/journal` | List journal entries |
| POST   | `/api/v1/journal` | Create journal entry |

### Capacity Planning

| Method | Endpoint                  | Description               |
| ------ | ------------------------- | ------------------------- |
| POST   | `/api/v1/capacity/plan`   | Plan capacity for service |
| GET    | `/api/v1/capacity/report` | Get capacity report       |

### Admin (requires admin scope)

| Method | Endpoint                    | Description    |
| ------ | --------------------------- | -------------- |
| GET    | `/api/v1/admin/apikeys`     | List API keys  |
| POST   | `/api/v1/admin/apikeys`     | Create API key |
| DELETE | `/api/v1/admin/apikeys/:id` | Delete API key |
