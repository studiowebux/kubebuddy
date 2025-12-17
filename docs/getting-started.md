---
title: Getting Started
description: Quick start guide for KubeBuddy
tags: [tutorial, quickstart, getting-started]
---

# Getting Started

## Start Server

Start server with admin API key:

```bash
export KUBEBUDDY_ADMIN_API_KEY=your-secret-key
kubebuddy server --db ~/kubebuddy.db --create-admin-key
```

Server starts on `http://localhost:8080` (use `--port` to change).

Environment variables:

```bash
export KUBEBUDDY_PORT=3000
export KUBEBUDDY_DB=~/kubebuddy.db
export KUBEBUDDY_CREATE_ADMIN_KEY=true
export KUBEBUDDY_ADMIN_API_KEY=your-secret-key
kubebuddy server
```

## Configure Client

Set client environment:

```bash
export KUBEBUDDY_API_KEY=your-secret-key
export KUBEBUDDY_ENDPOINT=http://localhost:8080
```

## Create API Keys

Create additional API keys:

```bash
kubebuddy apikey create --name dev-key --scope readwrite
```

Scopes:
- `admin`: Can manage API keys
- `readwrite`: Can read and modify resources
- `readonly`: Can only read resources

## First Steps

List computes:

```bash
kubebuddy compute list
```

Create a compute:

```bash
kubebuddy compute create \
  --name "server-01" \
  --type baremetal \
  --provider ovh \
  --region us-east \
  --monthly-cost 199.99
```

List services:

```bash
kubebuddy service list
```

Create a service:

```bash
kubebuddy service create \
  --name "web-server" \
  --min-spec '{"cores":1,"memory":2048}' \
  --max-spec '{"cores":2,"memory":4096}'
```

Find suitable computes:

```bash
kubebuddy plan web-server
```
