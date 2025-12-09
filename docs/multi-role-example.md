---
title: Multi-Role Baremetal Host Example
description: Example of a single baremetal host running multiple services using role-based placement
tags: [examples, placement, tags, multi-role]
---

# Multi-Role Baremetal Host Example

This example demonstrates how to configure a single baremetal host to run multiple services (database, web server, monitoring, and HAProxy) using role-based tagging and placement rules.

## Create Baremetal Host with Multiple Roles

```bash
# Create baremetal host with multiple role tags
kubebuddy compute create \
  --name "prod-server-01" \
  --type baremetal \
  --provider ovh \
  --region us-east \
  --tags "env=prod,role-database=true,role-web=true,role-monitoring=true,role-haproxy=true" \
  --monthly-cost 299.99
```

## Define Services with Role-Based Placement

### PostgreSQL Database Service

```bash
kubebuddy service create \
  --name "postgres-db" \
  --min-spec '{"cores":2,"memory":4096,"nvme":50}' \
  --max-spec '{"cores":4,"memory":8192,"nvme":100}' \
  --placement '{"affinity":[{"matchExpressions":[{"key":"role-database","operator":"Exists"}]}]}'
```

### Nginx Web Server Service

```bash
kubebuddy service create \
  --name "nginx-web" \
  --min-spec '{"cores":1,"memory":2048,"nvme":20}' \
  --max-spec '{"cores":2,"memory":4096,"nvme":50}' \
  --placement '{"affinity":[{"matchExpressions":[{"key":"role-web","operator":"Exists"}]}]}'
```

### Prometheus Monitoring Service

```bash
kubebuddy service create \
  --name "prometheus-monitoring" \
  --min-spec '{"cores":1,"memory":2048,"nvme":30}' \
  --max-spec '{"cores":2,"memory":4096,"nvme":50}' \
  --placement '{"affinity":[{"matchExpressions":[{"key":"role-monitoring","operator":"Exists"}]}]}'
```

### HAProxy Load Balancer Service

```bash
kubebuddy service create \
  --name "haproxy-lb" \
  --min-spec '{"cores":1,"memory":1024,"nvme":10}' \
  --max-spec '{"cores":2,"memory":2048,"nvme":20}' \
  --placement '{"affinity":[{"matchExpressions":[{"key":"role-haproxy","operator":"Exists"}]}]}'
```

## Plan and Assign Services

```bash
# Plan and assign all services to the baremetal host
kubebuddy plan postgres-db --assign
kubebuddy plan nginx-web --assign
kubebuddy plan prometheus-monitoring --assign
kubebuddy plan haproxy-lb --assign
```

## Verify Assignments

```bash
# List all assignments on the host
kubebuddy assignment list --compute prod-server-01

# Check total resource allocation
kubebuddy compute get prod-server-01
```

## Key Points

- **Single Host, Multiple Roles**: One baremetal server hosts all services by having multiple role tags
- **Tag Pattern**: Use `role-<service-type>=true` for each capability the host provides
- **Placement Matching**: Services use `MatchExpressions` with `Exists` operator to find hosts with specific roles
- **Resource Isolation**: Each service has defined min/max specs for capacity planning
- **Flexibility**: Easy to add/remove roles by updating tags without changing service definitions
