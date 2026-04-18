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

## Testing Expectations
- Add/adjust unit tests in the same PR for every new rule or bugfix.
- Prefer table-driven tests for deterministic helpers/protocol logic.
- Keep controller tests fake-client based unless envtest is explicitly required.

## CI Expectations
- Changes in `internal/*` should pass `go test ./...` and `make build`.
- Workflow changes must follow least-privilege permissions.
