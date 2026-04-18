# Architecture

```mermaid
flowchart LR
    A[USB Device on Node A] --> B[Agent DaemonSet]
    B --> C[USBDevice CR]
    C --> D[Controller/Reconciler]
    D --> E[USBDeviceApproval CR]
    D --> F[USBConnection CR]
    F --> G[Agent on Client Node]
    G --> H[Pod via /dev/ttyUSB*]
```

## Security Model

- Manual approval by default (`PendingApproval` -> `Approved`)
- Policy whitelist/blacklist controls
- Optional encryption requirement through policy flag
- Finalizer-based cleanup for exported devices
