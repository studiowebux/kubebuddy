# Resource Reference

## Resource Keys

Services use these keys in `min_spec` and `max_spec`:

| Key      | Unit  | Description      | Component Type |
|----------|-------|------------------|----------------|
| `cores`  | count | CPU threads      | cpu            |
| `memory` | MB    | System RAM       | ram, memory    |
| `vram`   | MB    | GPU video memory | gpu            |
| `nvme`   | GB    | Storage capacity | storage, nvme, ssd, hdd |

## Component Specs

### CPU

**Type:** `cpu`

**Spec fields** (any):
- `threads`, `thread_count` - Thread count
- `cores`, `core_count` - Core count

Maps to `cores` resource.

```bash
kubebuddy component create \
  --type cpu \
  --manufacturer Intel \
  --model "Xeon Gold 6258R" \
  --specs '{"threads":56}'
```

Total: `threads * quantity`

### RAM

**Type:** `ram` or `memory`

**Spec fields:**
- Fields ending in `_gb`: `capacity_gb`, `size_gb`, `memory_gb` (in GB, converted to MB)
- Other fields: `memory`, `size` (in MB)

Maps to `memory` resource (stored in MB).

```bash
# Using GB field
kubebuddy component create \
  --type ram \
  --manufacturer Samsung \
  --model "32GB DDR4" \
  --specs '{"capacity_gb":32}'

# Using MB field
kubebuddy component create \
  --type ram \
  --manufacturer Samsung \
  --model "32GB DDR4" \
  --specs '{"memory":32768}'
```

Total: `capacity_gb * 1024 * quantity` or `memory * quantity`

### GPU

**Type:** `gpu`

**Spec fields for VRAM:**
- Fields ending in `_gb`: `vram_gb`, `memory_gb`, `video_memory_gb` (in GB, converted to MB)
- Other fields: `vram`, `memory` (in MB)

Maps to `vram` resource (stored in MB).

```bash
# Using GB field
kubebuddy component create \
  --type gpu \
  --manufacturer NVIDIA \
  --model "RTX 4090" \
  --specs '{"vram_gb":24}'

# Using MB field
kubebuddy component create \
  --type gpu \
  --manufacturer NVIDIA \
  --model "RTX 4090" \
  --specs '{"vram":24576}'
```

Total: `vram_gb * 1024 * quantity` or `vram * quantity`

### Storage

**Type:** `storage`, `nvme`, `ssd`, `hdd`

**Spec fields** (any):
- `size`, `capacity_gb`, `storage_gb`, `capacity`

Maps to `nvme` resource (GB).

```bash
kubebuddy component create \
  --type nvme \
  --manufacturer Samsung \
  --model "980 Pro" \
  --specs '{"capacity_gb":1000}'
```

Total: `capacity_gb * quantity` (accounts for RAID)

## Units

**Memory and VRAM:**
- Storage: MB
- Display: GB
- Service specs: MB

**Storage:**
- Storage: GB
- Display: GB
- Service specs: GB

**CPU:**
- Storage: count
- Display: count
- Service specs: count

## Service Examples

Basic service:

```bash
kubebuddy service create \
  --name "web-server" \
  --min-spec '{"cores":2,"memory":4096,"nvme":100}' \
  --max-spec '{"cores":4,"memory":8192,"nvme":200}'
```

High memory database:

```bash
kubebuddy service create \
  --name "postgres" \
  --min-spec '{"cores":8,"memory":32768,"nvme":500}' \
  --max-spec '{"cores":16,"memory":65536,"nvme":2000}'
```

GPU workload:

```bash
kubebuddy service create \
  --name "ml-inference" \
  --min-spec '{"cores":4,"memory":16384,"vram":12288,"nvme":200}' \
  --max-spec '{"cores":8,"memory":32768,"vram":24576,"nvme":500}'
```

## Common Mistakes

**Wrong: Using GB values in service specs**
```bash
--min-spec '{"memory":20}'  # 20 MB, not 20 GB
```

**Correct: Service specs always use MB for memory/vram**
```bash
--min-spec '{"memory":20480}'  # 20 GB = 20480 MB
```

**Component specs: Use `_gb` suffix for GB values**
```bash
# Correct
--specs '{"capacity_gb":32}'  # 32 GB
--specs '{"vram_gb":24}'      # 24 GB

# Also correct (using MB directly)
--specs '{"memory":32768}'    # 32 GB = 32768 MB
--specs '{"vram":24576}'      # 24 GB = 24576 MB
```

**Wrong: Storage in MB**
```bash
--min-spec '{"nvme":307200}'  # Storage should be in GB
```

**Correct: Storage in GB**
```bash
--min-spec '{"nvme":300}'  # 300 GB
```

**Wrong: Invalid keys**
```bash
--min-spec '{"cpu":4,"ram":8192}'  # Use 'cores' and 'memory'
```

**Correct: Standard keys**
```bash
--min-spec '{"cores":4,"memory":8192}'
```

## Quick Reference

**Component Specs:**
- CPU: `threads` → `cores` (count)
- RAM: `capacity_gb` (GB) or `memory` (MB) → `memory` (stored as MB)
- GPU: `vram_gb` (GB) or `vram` (MB) → `vram` (stored as MB)
- Storage: `capacity_gb` → `nvme` (GB)

**Service Specs:**
- `cores`: CPU count
- `memory`: RAM in MB
- `vram`: GPU memory in MB
- `nvme`: Storage in GB

**Rule:** Component fields ending in `_gb` are in GB (converted to MB). Service specs always use MB for memory/vram.
