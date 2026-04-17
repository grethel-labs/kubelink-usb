# Kubernetes USB-Fabric Operator (K8s-USBIP)

This repository contains the initial scaffold for a Kubernetes operator and node agent that expose USB devices across nodes via USB/IP with a security-first approval flow.

## Architecture

```text
+----------------------+       +-------------------------+
| Node A               |       | Kubernetes Controller   |
| USB Device + Agent   +------>+ USBDevice Reconciler    |
| (Discovery/export)   |       | + Approval integration  |
+----------+-----------+       +------------+------------+
           |                                 |
           | USB/IP tunnel                   | CRDs
           v                                 v
+----------+-----------+       +-------------------------+
| Node B               |       | API Resources           |
| Agent (attach) + Pod |<------+ USBDevice              |
| /dev/ttyUSB*         |       | USBConnection          |
+----------------------+       | USBDevicePolicy        |
                               | USBDeviceApproval      |
                               +-------------------------+
```

## Quick start

1. Build binaries:
   ```bash
   make build
   ```
2. Install CRDs:
   ```bash
   make install
   ```
3. Run controller locally (needs kubeconfig):
   ```bash
   make run
   ```
4. Build container images:
   ```bash
   make docker-build
   ```

## Security highlights

- Device approval workflow (`PendingApproval` default phase)
- Policy-driven restrictions (namespaces/nodes/concurrency)
- Optional encryption requirement in policy model
- Finalizer hook for export cleanup
