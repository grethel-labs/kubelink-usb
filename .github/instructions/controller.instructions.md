---
applyTo: "internal/controller/**/*.go"
---

- Restrict changes to reconcile behavior, status/finalizer transitions, and controller wiring.
- Do not add direct network or filesystem dependencies in controller unit tests.
- Prefer fake controller-runtime client tests for decision-path validation.
- When adding new reconciliation branches, add explicit tests for success and error paths.
- Every reconciler type MUST have `@component` and `@flow` annotations documenting the reconcile decision tree.
- When adding phase transitions, add matching `@state` annotations on the relevant Status type in `api/v1alpha1/`.
- Before creating new `@component` node IDs, search existing annotations (`grep -rn '@component' --include='*.go'`) and reuse matching IDs (e.g. `Storage`, `PolicyEngine`, `USBDevice`). See the node ID table in `copilot-instructions.md`.
- Follow the doc comment convention defined in `copilot-instructions.md` — descriptive comments, not just names.
