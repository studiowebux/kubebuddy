---
title: WebUI
description: Web interface for KubeBuddy
tags: [webui, interface, management]
---

# WebUI

Embedded web interface for managing KubeBuddy resources through a browser.

## Starting the WebUI

Start both API server and WebUI with **ONE command**:

```bash
# Set admin API key
export KUBEBUDDY_ADMIN_API_KEY="your-secure-key"

# Start server with WebUI enabled
kubebuddy server --db ./kubebuddy.db --create-admin-key --webui

# Access:
# - API server: http://localhost:8080
# - WebUI: http://localhost:8081
```

**Server Flags:**
- `--webui`: Enable WebUI server (requires KUBEBUDDY_ADMIN_API_KEY)
- `--webui-port`: WebUI port (default: 8081)
- `--port`: API server port (default: 8080)
- `--db`: Database file path
- `--create-admin-key`: Create admin API key from KUBEBUDDY_ADMIN_API_KEY env var
- `--seed`: Seed database with sample data

**Examples:**

```bash
# Basic startup
export KUBEBUDDY_ADMIN_API_KEY="mykey123"
kubebuddy server --create-admin-key --webui

# Custom ports
kubebuddy server --port 9000 --webui --webui-port 9001 --create-admin-key

# With database seeding
kubebuddy server --db /data/kubebuddy.db --create-admin-key --seed --webui
```

## Features

### Theme Support
- Light/dark theme toggle
- Theme preference persisted in localStorage
- Smooth transitions between themes

### Compute Management
- List all compute resources
- Create new computes with metadata
- Delete compute resources
- Tag management
- Responsive table with horizontal scrolling

### Component Management
- List hardware components
- Create components (CPU, RAM, Storage, GPU, NIC, PSU, OS, Other)
- Define component specs as JSON
- Delete components

### Service Management
- List services
- Create services with resource requirements
- Define min/max resource specifications
- Delete services

### Assignment Management
- View service-to-compute assignments
- Monitor allocated resources
- Track assignment history
- Resolved compute and service names (not just IDs)

### IP Management
- List IP addresses
- Create IP addresses (public/private)
- Define CIDR and gateway
- Delete IP addresses

### DNS Management
- List DNS records
- Create DNS records (A, AAAA, CNAME, MX, TXT, NS)
- Zone management
- TTL configuration
- Delete DNS records

### Firewall Management
- List firewall rules
- Create firewall rules (ALLOW/DENY)
- Protocol configuration (TCP, UDP, ICMP)
- Port range specification
- Source/destination configuration
- Delete firewall rules

### Port Mappings
- List port mappings
- Create port assignments
- External to service port mapping
- Protocol selection
- Delete port mappings

### Journal
- List journal entries
- Create journal entries per compute
- Category selection (maintenance, deployment, incident, change, other)
- Resolved compute names
- Timestamp tracking

### API Keys
- List API keys
- Create API keys with scope control (admin, readwrite, readonly)
- Description management
- Delete API keys

### Reports
- Select compute from dropdown
- View comprehensive compute report including:
  - Compute information (name, type, provider, region, state, tags)
  - Resource Summary - Total showing allocated resources with percentages:
    - Total Cores (with % allocated)
    - Total Memory GB (with % allocated)
    - Total VRAM GB (with % allocated)
    - Total Storage GB (with % allocated)
  - Resource Summary - Statistics showing Min/Max/Avg/Median:
    - Each resource type shows absolute values and percentages
    - Percentages calculated against total available resources
  - Storage Configuration breakdown:
    - RAID groups with effective capacity calculation
    - Components per RAID group
    - Non-RAID storage
  - Hardware components with RAID configuration
  - Service assignments with allocated resources
  - IP address assignments (resolved to actual addresses)
  - Journal entries (sorted by date with proper timestamp formatting)

### Capacity Summary
- View system-wide capacity report
- Overview showing:
  - Total computes in system
  - Active computes count
  - Total services defined
  - Total assignments
- Compute Resources table displaying:
  - Compute name
  - Allocated resources (formatted: cores, RAM, VRAM, NVMe, SATA)
  - Available resources (formatted)
  - Statistics per compute:
    - Min: Sum of all minimum specifications
    - Max: Sum of all maximum specifications
    - Avg: Average of maximum specifications
    - Median: Median of maximum specifications

### IP Assignments
- View all IP assignments across system
- Displays:
  - Compute name (resolved from ID)
  - IP address (resolved from ID)
  - Interface name
  - Primary IP indicator
  - Creation timestamp

## Architecture

### Tech Stack
- **Frontend**: Vanilla JavaScript, HTML5, CSS3
- **Backend**: Go HTTP server with embedded static files
- **Communication**: REST API over HTTP
- **Deployment**: Single binary (embedded assets via go:embed)

### API Endpoints

The WebUI uses these REST endpoints:

**Computes:**
- `GET /api/computes` - List all computes
- `POST /api/computes` - Create compute
- `GET /api/computes/:id` - Get compute by ID or name
- `PUT /api/computes/:id` - Update compute
- `DELETE /api/computes/:id` - Delete compute

**Components:**
- `GET /api/components` - List all components
- `POST /api/components` - Create component

**Services:**
- `GET /api/services` - List all services
- `POST /api/services` - Create service

**Assignments:**
- `GET /api/assignments` - List all assignments

## Development

### File Structure

```
cmd/kubebuddy/webui/
├── index.html    # Main HTML structure with theme toggle
├── style.css     # CSS variables for theming, responsive tables
└── app.js        # Application logic with all resource management
```

### Building

Static files are embedded in the binary using Go's embed directive:

```go
//go:embed webui
var webuiFS embed.FS
```

No build step required - just `go build` and the assets are included.

### CORS

CORS is enabled for development to allow API calls:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization`

## Limitations

Current implementation provides comprehensive CRUD operations. Advanced features not yet implemented:
- Planning interface in WebUI (use CLI for capacity planning)
- Real-time updates
- User authentication (uses API key from CLI)
- Advanced filtering and search
- Pagination for large datasets
- Bulk operations
- Edit functionality for most resources (create/delete only)

## Security

- WebUI inherits API key from CLI flags or environment
- No separate authentication mechanism
- Suitable for internal/trusted networks
- For production, consider:
  - Reverse proxy with TLS
  - Network isolation
  - API key rotation

## Examples

### Creating a Compute

1. Navigate to "Computes" tab
2. Click "+ Add Compute"
3. Fill in the form:
   - Name: prod-server-01
   - Type: baremetal
   - Provider: ovh
   - Region: us-east
   - Tags: env=prod,zone=us-east-1
4. Click "Create"

### Creating a Component

1. Navigate to "Components" tab
2. Click "+ Add Component"
3. Fill in the form:
   - Name: Intel Xeon E5-2680v4
   - Type: cpu
   - Manufacturer: Intel
   - Model: E5-2680v4
   - Specs: `{"cores":14,"threads":28,"ghz":2.4}`
4. Click "Create"

### Creating a Service

1. Navigate to "Services" tab
2. Click "+ Add Service"
3. Fill in the form:
   - Name: postgres-db
   - Min Spec: `{"cores":2,"memory":4096,"nvme":100}`
   - Max Spec: `{"cores":4,"memory":8192,"nvme":200}`
4. Click "Create"

## Troubleshooting

**WebUI won't start:**
- Verify API server is running on the configured endpoint
- Check API key is set correctly
- Ensure port 8081 is available

**Can't create resources:**
- Verify API key has proper permissions
- Check browser console for errors
- Validate JSON syntax in spec fields

**Blank page:**
- Check browser console for JavaScript errors
- Verify static files are embedded (rebuild if necessary)
- Try a different browser
