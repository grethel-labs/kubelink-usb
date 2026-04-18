# Architecture

## Component Overview

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

## End-to-End Workflow

```mermaid
sequenceDiagram
    participant User as Physical User
    participant Node as Source Node (Agent)
    participant K8s as Kubernetes API
    participant Ctrl as Controller
    participant Policy as Policy Engine
    participant UI as UI / kubectl
    participant Client as Client Node (Agent)
    participant Pod as Pod / VM

    User->>Node: Plug in USB device
    Node->>Node: Discovery (fsnotify on /dev)
    Node->>K8s: Create USBDevice CR (Phase: PendingApproval)
    Ctrl->>K8s: Add finalizer, initialize status
    Ctrl->>Policy: Policy check (whitelist/blacklist/auto-approve)
    alt On whitelist or auto-approve
        Ctrl->>K8s: Phase → Approved
    else On blacklist
        Ctrl->>K8s: Phase → Denied
    else Unknown device
        UI->>K8s: Create USBDeviceApproval CR (manual)
        Ctrl->>K8s: Phase → Approved / Denied
    end
    Note over K8s: Only proceed if Phase=Approved
    K8s->>Node: Agent exports device via USB/IP (usbipd bind)
    Pod->>K8s: Create USBConnection CR (namespace-scoped)
    Ctrl->>Client: Agent on client node imports (usbip attach)
    Client->>Pod: /dev/ttyUSB* available in Pod
```

## CRD Relationships

```mermaid
erDiagram
    USBDevice ||--o{ USBDeviceApproval : "referenced by"
    USBDevice ||--o{ USBConnection : "referenced by"
    USBDevicePolicy ||--o{ USBDeviceApproval : "governs"
    USBConnection }o--|| Pod : "serves device to"

    USBDevice {
        string busID
        string nodeName
        string vendorID
        string productID
        string phase "PendingApproval|Approved|Denied|Disconnected"
    }
    USBDeviceApproval {
        string deviceRef
        string requester
        string phase "Pending|Approved|Denied"
        string approvedBy
    }
    USBDevicePolicy {
        string vendorID "selector"
        string productID "selector"
        string mode "manual|auto"
        bool autoApproveKnownDevices
    }
    USBConnection {
        string deviceRef
        string clientNode
        string phase "Pending|Connecting|Connected|Failed"
        string clientDevicePath
    }
```

## Phase Transitions

### USBDevice Lifecycle

```mermaid
stateDiagram-v2
    [*] --> PendingApproval: Agent creates CR
    PendingApproval --> Approved: Approval granted (manual or auto)
    PendingApproval --> Denied: Policy denies or manual reject
    Approved --> Exported: Agent exports via USB/IP
    Exported --> Disconnected: Device unplugged
    Disconnected --> PendingApproval: Device reconnected (if not on whitelist)
    Disconnected --> Approved: Device reconnected (on whitelist)
    Denied --> [*]: CR cleanup
```

### USBConnection Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Pending: Connection requested
    Pending --> Connecting: Device approved + export ready
    Connecting --> Connected: Attach successful
    Connected --> Disconnected: Network or device failure
    Disconnected --> Connecting: Reconnect attempt
    Disconnected --> Failed: Max retries exhausted
    Connected --> [*]: Connection deleted (finalizer cleanup)
    Failed --> [*]: Connection deleted
```

## Security Model

- Manual approval by default (`PendingApproval` → `Approved`)
- Policy whitelist/blacklist controls via `USBDevicePolicy` selector
- Auto-approve known devices via fingerprint whitelist
- Optional mTLS encryption for USB/IP tunnels (`requireEncryption` flag)
- Network isolation via automatic `NetworkPolicy` generation
- HID device class blocking (`denyHumanInterfaceDevices`)
- Namespace-scoped connections with allowed-namespace restrictions
- Finalizer-based cleanup for exported devices and tunnel teardown
- Max concurrent connections limit per device
