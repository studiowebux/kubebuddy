# RAID Configuration

## Supported RAID Levels

### RAID0 (Striping)

Capacity: Sum of all disks

```bash
kubebuddy component assign \
  --compute server-01 \
  --component nvme-1tb \
  --quantity 2 \
  --raid raid0 \
  --raid-group stripe-1

# Effective capacity: 2 * 1000 = 2000 GB
```

Use case: Maximum performance, no redundancy

### RAID1 (Mirroring)

Capacity: Size of smallest disk

```bash
kubebuddy component assign \
  --compute server-01 \
  --component nvme-1tb \
  --quantity 2 \
  --raid raid1 \
  --raid-group mirror-1

# Effective capacity: 1000 GB (smallest disk)
```

Use case: Maximum redundancy, survives 1 disk failure

### RAID5 (Striping with Parity)

Capacity: (n-1) * smallest disk (minimum 3 disks)

```bash
kubebuddy component assign \
  --compute server-01 \
  --component ssd-2tb \
  --quantity 4 \
  --raid raid5 \
  --raid-group parity-1

# Effective capacity: (4-1) * 2000 = 6000 GB
```

Use case: Good balance, survives 1 disk failure

### RAID6 (Double Parity)

Capacity: (n-2) * smallest disk (minimum 4 disks)

```bash
kubebuddy component assign \
  --compute server-01 \
  --component ssd-2tb \
  --quantity 6 \
  --raid raid6 \
  --raid-group double-parity-1

# Effective capacity: (6-2) * 2000 = 8000 GB
```

Use case: High redundancy, survives 2 disk failures

### RAID10 (Mirrored Stripes)

Capacity: Sum / 2 (minimum 4 disks, even count)

```bash
kubebuddy component assign \
  --compute server-01 \
  --component nvme-1tb \
  --quantity 4 \
  --raid raid10 \
  --raid-group mirror-stripe-1

# Effective capacity: (4 * 1000) / 2 = 2000 GB
```

Use case: High performance + redundancy, survives multiple failures

## RAID Groups

Components with the same `--raid-group` form a single RAID array. Effective capacity is calculated automatically.

### Single Array

```bash
# All drives in one RAID5 array
kubebuddy component assign \
  --compute server-01 \
  --component ssd-4tb \
  --quantity 5 \
  --raid raid5 \
  --raid-group storage-1
```

### Multiple Arrays

```bash
# First RAID1 array
kubebuddy component assign \
  --compute server-01 \
  --component nvme-500gb \
  --quantity 2 \
  --raid raid1 \
  --raid-group os-mirror

# Second RAID5 array
kubebuddy component assign \
  --compute server-01 \
  --component ssd-4tb \
  --quantity 4 \
  --raid raid5 \
  --raid-group data-storage
```

## Mixed Disk Sizes

RAID calculations use the smallest disk in the array.

```bash
# RAID5 with mixed sizes
kubebuddy component assign \
  --compute server-01 \
  --component ssd-2tb \
  --quantity 2 \
  --raid raid5 \
  --raid-group mixed-1

kubebuddy component assign \
  --compute server-01 \
  --component ssd-4tb \
  --quantity 2 \
  --raid raid5 \
  --raid-group mixed-1

# Effective capacity: (4-1) * 2000 = 6000 GB (smallest disk = 2TB)
```

## Non-RAID Storage

Omit `--raid` and `--raid-group` for standalone disks.

```bash
kubebuddy component assign \
  --compute server-01 \
  --component nvme-1tb \
  --quantity 1

# Effective capacity: 1000 GB
```

## Capacity Calculation Examples

### RAID0 Example

4 x 1TB drives = 4TB total

### RAID1 Example

2 x 1TB drives = 1TB total
4 x 2TB drives = 2TB total (smallest)

### RAID5 Example

3 x 2TB drives = 4TB total (3-1 = 2 drives)
5 x 4TB drives = 16TB total (5-1 = 4 drives)

### RAID6 Example

4 x 2TB drives = 4TB total (4-2 = 2 drives)
6 x 4TB drives = 16TB total (6-2 = 4 drives)

### RAID10 Example

4 x 1TB drives = 2TB total (4 / 2)
8 x 2TB drives = 8TB total (16 / 2)
