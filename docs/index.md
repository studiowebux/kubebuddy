---
title: Overview
description: KubeBuddy architecture and global configuration
tags: [architecture, configuration, getting-started]
---

# Overview

Capacity planning tool for managing compute resources and services. Server/client architecture with API key authentication.

## Architecture

- Server: SQLite database, REST API
- Client: CLI commands communicate via HTTP
- Authentication: API key with scopes (admin, readwrite, readonly)

## Shell Completion

Enable shell completion for faster command usage:

**Bash:**
```bash
kubebuddy completion bash | sudo tee /etc/bash_completion.d/kubebuddy
```

**Zsh:**
```bash
kubebuddy completion zsh > ~/.zsh/completions/_kubebuddy
```

Completion includes commands, flags, compute IDs, service IDs, component IDs, and more.

## Global Flags

All commands support:

- `--endpoint`: API endpoint (default: http://localhost:8080 or KUBEBUDDY_ENDPOINT env)
- `--api-key`: API key (default: KUBEBUDDY_API_KEY env)
- `--version`, `-v`: Show version
