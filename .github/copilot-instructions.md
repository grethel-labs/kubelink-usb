# Repository Agent Instructions

## Mission
- Keep changes minimal, testable, and aligned with Kubernetes operator conventions used in this repo.
- Prefer focused updates in existing files over large refactors.

## Folder Responsibilities
- `api/v1alpha1/`: CRD API types only. Keep schema and status/spec definitions here.
- `cmd/controller/`, `cmd/agent/`: process entrypoints and wiring only.
- `internal/controller/`: reconcile and Kubernetes object lifecycle logic.
- `internal/agent/`: node-local discovery/event logic.
- `internal/security/`: policy and crypto/tls guardrails.
- `internal/usbip/`: USB/IP protocol/data handling.
- `internal/utils/`: pure helper functions with no side effects.
- `config/`: Kubernetes manifests and deployment config.
- `.github/workflows/`: CI checks (tests, coverage, build validations).

## Go Doc Comment Convention

Every exported symbol (type, function, method, constant, variable) MUST have a Go doc comment.
Comments are enforced by `golangci-lint` with the `revive/exported` rule.

### Structure for types

```go
// TypeName is/does <one-line purpose>.
// <2-3 sentence description: what it represents, why it exists, where it fits
// in the architecture.>
//
// @component NodeID["Label"] --> TargetID["Label"]
// @relates TypeA ||--o{ TypeB : "relationship label"
// @state StateA --> StateB : trigger description
// @flow StepA["Label"] --> StepB["Label"]
type TypeName struct {
```

### Structure for functions/methods

```go
// FuncName <one-line what it does in active voice>.
// <Optional 1-2 sentence context about when/why this is called.>
//
// Intent: <why this exists>
// Inputs: <important parameters>
// Outputs: <return values meaning>
// Errors: <error contract>
func FuncName(...) ... {
```

### Diagram Annotation Tags

These tags are extracted by `hack/generate-diagrams.sh` to auto-generate Mermaid diagrams:

| Tag | Diagram Type | Syntax | Use for |
|-----|-------------|--------|---------|
| `@component` | Flowchart | `NodeID["Label"] --> TargetID["Label"]` | Architecture overview, component wiring |
| `@relates` | ER Diagram | `TypeA \|\|--o{ TypeB : "label"` | CRD type relationships |
| `@state` | State Diagram | `StateA --> StateB : trigger` | Phase/status lifecycle transitions |
| `@flow` | Flowchart | `StepA["Label"] --> StepB["Label"]` | Reconcile/processing logic flows |

Rules:
- Place annotations on the **type** that owns the behavior (reconciler, CRD status type).
- Use Mermaid-valid syntax after the tag — the generator copies them verbatim.
- `@state` annotations go on Status types or Reconciler types that drive phase transitions.
- `@flow` annotations go on Reconciler types to document reconcile decision logic.
- `@component` annotations go on top-level CRD types or reconciler types to show wiring.
- `@relates` annotations go on CRD types to document ER relationships.
- Node IDs must be unique across the codebase (prefix with context if needed).
- Every new reconciler MUST have at least `@component` and `@flow` annotations.
- Every CRD type with a Phase field MUST have `@state` annotations on its Status type.

### Reusing existing Node IDs

Before adding a new `@component` or `@relates` annotation, **search `docs/DIAGRAMS.md` and existing `@component` annotations** for node IDs that already represent the target concept. The diagram generator groups edges into clusters by shared node IDs — using a new ID for an existing concept creates orphaned mini-diagrams instead of a connected architecture view.

Checklist:
1. Run `grep -rn '@component' --include='*.go' | grep -v _test.go` to see all existing node IDs.
2. If the target you're connecting to already has an ID (e.g. `Storage["Backup Storage"]`, `PolicyEngine["Policy Engine"]`, `USBDevice["USBDevice CR"]`), reuse that exact ID.
3. Only create a new node ID when no existing ID matches the concept.
4. Keep labels consistent — the same node ID must always carry the same `["Label"]`.

Common established node IDs:

| Node ID | Label | Represents |
|---------|-------|------------|
| `Storage` | `"Backup Storage"` | BackupStorage interface / storage layer |
| `PolicyEngine` | `"Policy Engine"` | security.Engine / policy evaluation |
| `USBDevice` | `"USBDevice CR"` | USBDevice custom resource |
| `PolicyCR` | `"USBDevicePolicy CR"` | USBDevicePolicy custom resource |
| `AgentServer` | `"Agent Export"` | agent.Server / usbipd export |
| `AgentClient` | `"Agent Attach"` | agent.Client / usbip attach |
| `K8sAPI` | `"Kubernetes API"` | Kubernetes API server |
| `CMStorage` | `"ConfigMapStorage"` | ConfigMap backup backend |
| `PVCBackupStorage` | `"PVCStorage"` | PVC backup backend |
| `S3BackupStorage` | `"S3Storage"` | S3 backup backend |

### Examples

```go
// USBDeviceStatus defines the observed state of USBDevice.
// Phase tracks the device lifecycle from discovery through approval.
//
// @state [*] --> PendingApproval : Agent discovers device
// @state PendingApproval --> Approved : Approval granted
// @state Approved --> Exported : Agent exports via USB/IP
type USBDeviceStatus struct {

// BackupReconciler reconciles USBBackup objects.
// It collects security-relevant CRs, creates a checksummed snapshot,
// and persists it via the configured storage backend.
//
// @component BackupReconciler["Backup Reconciler"] --> Storage["Backup Storage"]
// @flow CollectCRs["List CRs"] --> CreateSnapshot["Create snapshot"]
// @flow CreateSnapshot --> WriteStorage["Write to storage"]
type BackupReconciler struct {
```

## Testing Expectations
- Add/adjust unit tests in the same PR for every new rule or bugfix.
- Prefer table-driven tests for deterministic helpers/protocol logic.
- Keep controller tests fake-client based unless envtest is explicitly required.

## CI Expectations
- Changes in `internal/*` should pass `go test ./...` and `make build`.
- Workflow changes must follow least-privilege permissions.
- `golangci-lint` with `revive` enforces doc comments — fix lint errors before pushing.
