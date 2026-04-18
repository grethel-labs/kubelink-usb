# kubelink-usb (K8s-USBIP)

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

## Agent instructions for LLM workflows

- Global agent guidance lives in `.github/copilot-instructions.md`.
- Folder-specific instructions live in `.github/instructions/*.instructions.md`.
- Add or adjust these instruction files when introducing new architectural areas.

## CI automation

GitHub Actions workflow `.github/workflows/unit-tests.yml` now validates:
- strict linting on every commit (`gofmt` check + `go vet`)
- unit tests (`make test`)
- coverage gates (`make coverage-check`, overall minimum 80%)
- binary builds (`make build`) with uploaded artifacts (`bin/controller`, `bin/agent`)
- container image builds for controller and agent (`Dockerfile`, `Dockerfile.agent`)
- generated documentation consistency (`make docs` + committed output)
- automatic image publishing to GHCR on `push` to `main` (`latest` + commit SHA tags)

## Branch protection and approval policy

To enforce "no direct push to `main` except approved flow", configure repository **Branch protection / Rulesets** in GitHub settings:
- require pull requests before merging
- require at least one approval
- require status checks to pass (at minimum the `CI / lint` check)
- optionally restrict who can push directly to `main`

Without these repository settings, workflows can fail a commit but cannot block a direct push by themselves.

## License and legal notes for publishing artifacts

Repository license: **Apache-2.0** (`LICENSE`).

Before publishing artifacts, verify:
- dependency licenses are compatible with your intended distribution model
- required notices/attribution files are included for redistributed components
- no proprietary material, secrets, or personal data is packaged
- trademark and export-control obligations are assessed for your target regions

This is engineering guidance and not legal advice.

## Code documentation system

- Unified commenting structure is defined in `docs/CODE_REFERENCE.md`.
- Architecture/reference markdown is generated from code with:
  ```bash
  make docs
  ```
- Generated output (including Mermaid diagrams) is committed in:
  - `docs/CODE_REFERENCE.md`

## Security highlights

- Device approval workflow (`PendingApproval` default phase)
- Policy-driven restrictions (namespaces/nodes/concurrency)
- Optional encryption requirement in policy model
- Finalizer hook for export cleanup
