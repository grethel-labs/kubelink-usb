---
applyTo: "internal/{security,usbip,utils,agent}/**/*.go"
---

- Treat these packages as unit-test-first areas.
- Add table-driven tests for deterministic logic and malformed input handling.
- Keep protocol/security behavior explicit and avoid hidden defaults.
- Ensure API/encoding behavior is stable and backwards-compatible unless intentionally changed.
- Every exported type MUST have a descriptive doc comment explaining purpose and architecture context.
- Use `@component` annotations on types that represent system building blocks (Discovery, Server, Client, Engine).
- Use `@flow` annotations on types that implement multi-step processing logic.
- Before creating new `@component` node IDs, search existing annotations (`grep -rn '@component' --include='*.go'`) and reuse matching IDs (e.g. `AgentServer`, `AgentClient`, `PolicyEngine`, `K8sAPI`). See the node ID table in `copilot-instructions.md`.
- Follow the doc comment convention defined in `copilot-instructions.md`.
