# lkm - Local Kubernetes Manager

TUI app for managing k3d/kind/minikube clusters.

## Features

- Cluster lifecycle management (list, create, delete, start, stop)
- Support for k3d, kind, and minikube
- Real-time status monitoring
- Simple TUI interface

## Installation

```bash
mise install
mise run build
./bin/lkm
```

Or install directly:

```bash
mise run install
lkm
```

## Requirements

At least one of:
- [k3d](https://k3d.io/)
- [kind](https://kind.sigs.k8s.io/)
- [minikube](https://minikube.sigs.k8s.io/)

## Usage

```bash
lkm
```

### Keybindings

**List View:**
- `j/k` or `↑/↓` - Navigate
- `c` - Create cluster
- `d` - Delete selected
- `s` - Start selected
- `x` - Stop selected
- `Enter` - Detail view
- `r` - Refresh
- `?` - Help
- `q` - Quit

**Detail View:**
- `Esc` - Back to list
- `d` - Delete cluster
- `s` - Start cluster
- `x` - Stop cluster

**Create View:**
- `Tab` - Next field
- `Enter` - Submit
- `Esc` - Cancel

## Development

```bash
# Run tests
mise run test

# Lint
mise run lint

# Build
mise run build

# Run
mise run run
```

## Logging

Logs are written to `~/.local/share/lkm/lkm.log`.

## License

MIT
