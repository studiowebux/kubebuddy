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
