# Kube Buddy

Capacity planning system for managing compute resources (baremetal, VPS, VM) and services with intelligent placement rules.

## Features

- **Compute Management**: Register and track baremetal servers, VPS, and VMs with dynamic resource attributes
- **Component Management**: Hardware component catalog (CPU, RAM, Storage, GPU, NIC, PSU) with assign/unassign to computes
- **Service Management**: Define services with min/max resource specifications
- **Tag-Based Placement**: Kubernetes-style affinity, anti-affinity, and spread constraints
- **Capacity Planning**: Best-fit algorithm evaluates capacity and recommends purchases when needed
- **Journal System**: Per-compute logging with predefined and custom categories (supports markdown)
- **API Security**: Multi-scope API keys (admin, readwrite, readonly)
- **CLI + API**: 100% command-line driven with REST API

## Quick Start

### Build and Install

```bash
# Build
go build -o kubebuddy ./cmd/kubebuddy

# Install to PATH (optional but recommended for completion)
sudo cp kubebuddy /usr/local/bin/
```

### Start Server

**Production (admin key only):**
```bash
ADMIN_API_KEY=your-secret-key kubebuddy server --create-admin-key
```

**Development (admin key + sample data):**
```bash
ADMIN_API_KEY=test123 kubebuddy server --create-admin-key --seed
```

Server starts on `http://localhost:8080` (use `--port` to change)

Sample data includes:
- 3 compute resources (baremetal, VPS, VM)
- 3 services (nginx-ingress, postgres-db, ml-worker)
- 5 hardware components (Intel Xeon CPU, Samsung RAM, NVMe SSD, NVIDIA T4 GPU, Intel 10GbE NIC)
- 5 component assignments (components assigned to computes)
- 1 service assignment (nginx on baremetal)
- Sample journal entry

### Shell Completion (Powered by Cobra)

KubeBuddy uses Cobra for intelligent tab completion with **dynamic server data fetching**.

**Bash**:
```bash
# Enable for current session
source <(kubebuddy completion bash)

# Install permanently
kubebuddy completion bash | sudo tee /etc/bash_completion.d/kubebuddy
```

**Zsh**:
```bash
# Enable for current session
source <(kubebuddy completion zsh)

# Install permanently
kubebuddy completion zsh > ~/.zsh/completions/_kubebuddy
```

**Fish**:
```bash
kubebuddy completion fish > ~/.config/fish/completions/kubebuddy.fish
```

**What gets auto-completed**:
- All commands and subcommands
- **Compute IDs with names** (dynamically fetched from server)
- **Service IDs with names** (dynamically fetched from server)
- **Component IDs with names** (dynamically fetched from server)
- **Assignment IDs** (dynamically fetched from server)
- **API Key IDs with names** (dynamically fetched from server)
- Journal categories (maintenance, incident, deployment, hardware, network, other)
- Compute types (baremetal, vps, vm)
- Component types (cpu, ram, storage, gpu, nic, psu, other)
- API key scopes (admin, readwrite, readonly)
- All flag names and values

### CLI Usage

Export your API key:
```bash
export KUBEBUDDY_API_KEY=secret123
```

List computes:
```bash
kubebuddy compute list
```

List services:
```bash
kubebuddy service list
```

Plan capacity for a service (with tab completion):
```bash
kubebuddy plan <TAB><TAB>
# Shows: service-id:service-name for all services
```

List components:
```bash
kubebuddy component list
```

Assign component to compute:
```bash
kubebuddy component assign --compute <TAB><TAB> --component <TAB><TAB> --quantity 2
```

Add journal entry (with tab completion for computes and categories):
```bash
kubebuddy journal add --compute <TAB><TAB> --category <TAB><TAB> --content "Kernel upgrade completed"
```

## API Endpoints

### Computes
- `GET /api/v1/computes`
- `GET /api/v1/computes/:id`
- `POST /api/v1/computes`
- `PUT /api/v1/computes/:id`
- `DELETE /api/v1/computes/:id`

### Services
- `GET /api/v1/services`
- `GET /api/v1/services/:id`
- `POST /api/v1/services`
- `PUT /api/v1/services/:id`
- `DELETE /api/v1/services/:id`

### Assignments
- `GET /api/v1/assignments`
- `POST /api/v1/assignments`
- `DELETE /api/v1/assignments/:id`

### Capacity Planning
- `POST /api/v1/capacity/plan`
- `GET /api/v1/capacity/report`

### Journal
- `GET /api/v1/journal`
- `POST /api/v1/journal`

### Components
- `GET /api/v1/components`
- `GET /api/v1/components/:id`
- `POST /api/v1/components`
- `PUT /api/v1/components/:id`
- `DELETE /api/v1/components/:id`

### Component Assignments
- `GET /api/v1/component-assignments?compute_id=uuid`
- `GET /api/v1/component-assignments?component_id=uuid`
- `POST /api/v1/component-assignments`
- `DELETE /api/v1/component-assignments/:id`

### Admin (requires admin API key)
- `GET /api/v1/admin/apikeys`
- `POST /api/v1/admin/apikeys`
- `DELETE /api/v1/admin/apikeys/:id`

## Configuration

### Server

Flags:
- `--db`: Database file path (default: `kubebuddy.db`)
  - Supports `~` expansion (e.g., `~/.kubebuddy/kubebuddy.db`)
  - Auto-creates directories if they don't exist
  - If path ends with `/`, automatically appends `kubebuddy.db`
- `--port`: Server port (default: `8080`)
- `--create-admin-key`: Create admin API key from ADMIN_API_KEY env var
- `--seed`: Populate database with sample data

Environment variables:
- `ADMIN_API_KEY`: Admin API key (required when using --create-admin-key)

**Examples:**
```bash
# Database in home directory (auto-creates ~/.kubebuddy/)
ADMIN_API_KEY=secret kubebuddy server --create-admin-key --db ~/.kubebuddy/

# Custom directory (auto-creates /var/lib/kubebuddy/)
ADMIN_API_KEY=secret kubebuddy server --create-admin-key --db /var/lib/kubebuddy/

# Specific filename
ADMIN_API_KEY=secret kubebuddy server --create-admin-key --db ~/.kubebuddy/production.db
```

### CLI

Environment variables:
- `KUBEBUDDY_ENDPOINT`: API endpoint (default: `http://localhost:8080`)
- `KUBEBUDDY_API_KEY`: API key for authentication

## API Key Management

Create a new API key:
```bash
./kubebuddy apikey create --name developer --scope readwrite --description "Dev team key"
```

Scopes:
- `admin`: Can manage API keys
- `readwrite`: Can read and modify resources
- `readonly`: Can only read resources
