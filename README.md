# kubelink-usb (K8s-USBIP)

A Kubernetes operator and node agent for sharing USB devices across cluster nodes via USB/IP with a security-first approval flow, backup/restore capabilities, and automated health monitoring.

## Project Status

| Area | Status | Coverage |
|------|--------|----------|
| **CRD API Types** (8 resources) | ✅ Complete | 98.9% |
| **USBDevice Reconciler** (finalizer + status) | ✅ Complete | 84.0% |
| **Backup/Restore System** (snapshot, storage, controllers) | ✅ Complete | 91.2% |
| **Health Monitor** (consistency checks, auto-restore) | ✅ Complete | — |
| **Discovery Watcher** (fsnotify, event normalization) | ✅ Complete | 69.0% |
| **TLS Baseline** + Whitelist | ✅ Complete | 100.0% |
| **USB/IP Protocol** (BasicHeader only) | 🔶 Partial | 100.0% |
| **Policy Engine** (always returns true) | ⚠️ Stub | 100.0% |
| **Approval Controller** | ⚠️ Stub | — |
| **Connection Controller** | ⚠️ Stub | — |
| **Agent Client/Server** (Attach/Export) | ⚠️ Stub | — |
| **Overall** | ~45% of v1.0 | **85.3%** |

## Architecture

```text
┌─────────────────────────────────────────────────────────────────────┐
│ CONTROLLER (cmd/controller)                                         │
├─────────────────────────────────────────────────────────────────────┤
│ ✅ USBDeviceReconciler — watches USBDevice, initializes status      │
│ ⚠️  ApprovalReconciler — placeholder (Phase 2)                      │
│ ⚠️  USBConnectionReconciler — placeholder (Phase 3)                 │
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
│ ⚠️  Server.Export() — placeholder (should call usbipd bind)         │
│ ⚠️  Client.Attach() — placeholder (should call usbip attach)        │
│ ⚠️  USB/IP protocol — BasicHeader only, import/export stubs         │
└─────────────────────────────────────────────────────────────────────┘
```

## Quick Start

1. Build binaries:
   ```bash
   make build          # outputs bin/controller + bin/agent
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
   make docker-build   # builds both controller + agent images
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
- strict linting (`gofmt` check + `go vet`)
- unit tests (`make test`) — 53 test functions across 15 files
- coverage gates (`make coverage-check`, overall minimum 80%, currently **85.3%**)
- binary builds (`make build`) with uploaded artifacts
- container image builds for controller and agent
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
- Implementation roadmap with progress tracking in `docs/TODO.md`
- Auto-generated reference:
  ```bash
  make docs
  ```

## Security Highlights

- Device approval workflow (`PendingApproval` default phase)
- Policy-driven restrictions (namespaces/nodes/concurrency)
- Optional encryption requirement in policy model (TLS 1.3 baseline)
- Finalizer hook for export cleanup
- Backup integrity validation (SHA-256 checksums)
- Auto-restore with health monitoring (10min cooldown, 3 max retries)
