# kubelink-usb (K8s-USBIP)

> **Early-Stage Software (v0.1-alpha)** — This project is in an early development phase and not yet production-ready. APIs, CRDs, and behavior may change without notice. This codebase has been generated almost entirely (~100%) by AI (GitHub Copilot / Claude) and should be reviewed carefully before any production use.

A Kubernetes operator and node agent for sharing USB devices across cluster nodes via USB/IP with a security-first approval flow, backup/restore capabilities, and automated health monitoring.

## Architecture

```text
┌─────────────────────────────────────────────────────────────────────┐
│ CONTROLLER (cmd/controller)                                         │
├─────────────────────────────────────────────────────────────────────┤
│ ✅ USBDeviceReconciler — watches USBDevice, initializes status      │
│ ✅ ApprovalReconciler — processes approvals, propagates to device   │
│ ✅ USBConnectionReconciler — tunnel lifecycle orchestration         │
│ ✅ BackupReconciler — collects + snapshots config to storage        │
│ ✅ RestoreReconciler — validates + applies + revalidates            │
│ ✅ HealthMonitor — checks consistency, triggers auto-restore        │
└─────────────────────────────────────────────────────────────────────┘
                                 ↓ kubectl apply
                        ┌────────────────────────┐
                        │ Kubernetes API Server  │
                        │                        │
                        │ CRDs:                  │
                        │  USBDevice             │
                        │  USBDeviceApproval     │
                        │  USBDevicePolicy       │
                        │  USBConnection         │
                        │  USBDeviceWhitelist    │
                        │  USBBackupConfig       │
                        │  USBBackup             │
                        │  USBRestore            │
                        └────────────────────────┘
                                 ↓ watches
┌─────────────────────────────────────────────────────────────────────┐
│ AGENT (cmd/agent)                                                   │
├─────────────────────────────────────────────────────────────────────┤
│ ✅ Discovery — fsnotify on /dev, logs normalized events             │
│ ✅ Server.Export() — calls usbipd bind via CommandRunner            │
│ ✅ Client.Attach() — calls usbip attach via CommandRunner           │
│ ✅ USB/IP protocol — DevList + Import frames, TCP server/client     │
│ ✅ Device Fingerprinting — DNS-safe, deterministic CR names         │
└─────────────────────────────────────────────────────────────────────┘
```

## Quick Start

1. Build binaries:
   ```bash
   make build          # outputs bin/controller + bin/agent + bin/kubectl-usb
   ```
2. Install CRDs:
   ```bash
   make install        # kubectl apply -f config/crd/bases/
   ```
3. Run controller locally (needs kubeconfig):
   ```bash
   make run-controller
   ```
4. Run agent locally:
   ```bash
   make run-agent
   ```
5. Build container images:
   ```bash
   make docker-build   # builds both controller + agent images (version from project.env)
   ```
6. Deploy via Helm:
   ```bash
   helm install kubelink-usb charts/kubelink-usb
   ```

## CRD Resources (8 types)

| Resource | Scope | Short | Purpose |
|----------|-------|-------|---------|
| `USBDevice` | Cluster | `usbdev` | Discovered device representation |
| `USBDeviceApproval` | Cluster | `usbappr` | Manual/auto approval requests |
| `USBDevicePolicy` | Cluster | `usbpol` | Security rules (selectors, restrictions) |
| `USBConnection` | Namespaced | `usbconn` | Tunnel lifecycle (export↔attach) |
| `USBDeviceWhitelist` | Cluster | `usbwl` | Known-safe device registry |
| `USBBackupConfig` | Cluster | `usbbc` | Backup storage destination (PVC/ConfigMap/S3) |
| `USBBackup` | Cluster | `usbbk` | Backup request + result + checksum |
| `USBRestore` | Cluster | `usbrs` | Restore request with dry-run + health validation |

## CI Automation

GitHub Actions workflow `.github/workflows/unit-tests.yml` validates:
- strict linting (`gofmt` check + `go vet` + `golangci-lint` with `revive`)
- exported symbol documentation enforcement (`revive/exported` + `revive/package-comments`)
- unit tests (`make test`) — 80+ test functions across 15+ files
- coverage gates (`make coverage-check`, overall minimum 80%, currently **80.0%**)
- binary builds (`make build`) with uploaded artifacts
- container image builds for controller and agent (multi-arch: amd64 + arm64)
- Helm chart validation (`helm lint` + `helm template`)
- generated documentation consistency (`make docs` + committed output)
- automatic image publishing to GHCR on `push` to `main`

## Branch Protection

Configure repository **Branch protection / Rulesets** in GitHub settings:
- require pull requests before merging
- require at least one approval
- require status checks to pass (at minimum the `CI / lint` check)

## License

Repository license: **Apache-2.0** (`LICENSE`).

## Code Documentation

- Unified commenting structure defined in `docs/CODE_REFERENCE.md`
- Detailed architecture and workflow diagrams in `docs/architecture.md`
- Auto-generated Mermaid diagrams from `@component`/`@relates`/`@state`/`@flow` annotations in `docs/DIAGRAMS.md`
- API documentation generated from Go doc comments via [gomarkdoc](https://github.com/princjef/gomarkdoc)
- Comment documentation enforced by `golangci-lint` + `revive` (`exported`, `package-comments` rules)
- Dependency graph generation via [goda](https://github.com/loov/goda) + Graphviz
- Auto-generated reference:
  ```bash
  make docs          # CODE_REFERENCE.md + per-package DOC.md + DIAGRAMS.md
  make docs-deps     # dependency-graph.svg (requires goda + graphviz)
  make helm-lint     # validate Helm chart
  ```

## Security Highlights

- Device approval workflow (`PendingApproval` default phase)
- Policy-driven restrictions (namespaces/nodes/concurrency)
- Optional encryption requirement in policy model (TLS 1.3 baseline)
- Finalizer hook for export cleanup
- Backup integrity validation (SHA-256 checksums)
- Auto-restore with health monitoring (10min cooldown, 3 max retries)
