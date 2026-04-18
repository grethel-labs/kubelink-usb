---
applyTo: "internal/{security,usbip,utils,agent}/**/*.go"
---

- Treat these packages as unit-test-first areas.
- Add table-driven tests for deterministic logic and malformed input handling.
- Keep protocol/security behavior explicit and avoid hidden defaults.
- Ensure API/encoding behavior is stable and backwards-compatible unless intentionally changed.
