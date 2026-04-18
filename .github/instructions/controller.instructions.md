---
applyTo: "internal/controller/**/*.go"
---

- Restrict changes to reconcile behavior, status/finalizer transitions, and controller wiring.
- Do not add direct network or filesystem dependencies in controller unit tests.
- Prefer fake controller-runtime client tests for decision-path validation.
- When adding new reconciliation branches, add explicit tests for success and error paths.
