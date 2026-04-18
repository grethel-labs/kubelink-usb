---
applyTo: ".github/workflows/**/*.yml"
---

- Use explicit, least-privilege `permissions`.
- Keep CI deterministic: pin setup actions to major versions and use Make targets where possible.
- For build workflows, validate binaries/images on pull requests; publish only from trusted branches/tags.
