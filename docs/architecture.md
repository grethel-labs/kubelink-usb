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

    subgraph Backup System
        I[USBBackupConfig CR] --> J[BackupReconciler]
        J --> K[USBBackup CR]
        K --> L[BackupStorage]
        L --> M[ConfigMap / PVC / S3]
        N[RestoreReconciler] --> O[USBRestore CR]
        P[HealthMonitor] --> N
    end
```

## Implementation Status

```mermaid
pie title Component Completion (v1.0)
    "Complete" : 45
    "Stub/Placeholder" : 30
    "Not Started" : 25
```

| Component | Status | Notes |
|-----------|--------|-------|
| CRD API Types (8 resources) | ✅ Complete | USBDevice, Approval, Policy, Connection, Whitelist, BackupConfig, Backup, Restore |
| USBDevice Reconciler | ✅ Complete | Finalizer + status init (PendingApproval, LastSeen, Healthy) |
| Backup Controller | ✅ Complete | Snapshot collection, storage write, retention enforcement |
| Restore Controller | ✅ Complete | 5-phase lifecycle, dry-run, connection revalidation |
| Health Monitor | ✅ Complete | Consistency checks, auto-restore (10min cooldown, 3 retries) |
| Discovery Watcher | ✅ Complete | fsnotify on /dev, event normalization, USB path filtering |
| TLS Baseline | ✅ Complete | TLS 1.3 minimum config |
| Whitelist (in-memory) | ✅ Complete | Thread-safe string set |
| USB/IP BasicHeader | ✅ Complete | 6-byte encode/decode |
| ConfigMap Backup Storage | ✅ Complete | Thread-safe in-memory map |
| Policy Engine | ⚠️ Stub | `Allows()` → always true |
| Approval Controller | ⚠️ Stub | Returns nil (no-op) |
| Connection Controller | ⚠️ Stub | Returns nil (no-op) |
| Agent Client (Attach/Detach) | ⚠️ Stub | Returns nil |
| Agent Server (Export/Unexport) | ⚠️ Stub | Returns nil |
| USB/IP Client (Connect) | ⚠️ Stub | Returns nil |
| USB/IP Server (Serve) | ⚠️ Stub | Returns nil |
| PVC Backup Storage | ⚠️ Stub | Interface only |
| S3 Backup Storage | ⚠️ Stub | Interface only |
| Device Fingerprinting | ❌ Missing | Needed for deterministic CR names |
| Discovery→CR Bridge | ❌ Missing | Discovery logs but doesn't create CRs |
| Full USB/IP Protocol | ❌ Missing | Only BasicHeader, no device list/import frames |

## End-to-End Workflow (Target State)

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
    USBDeviceWhitelist ||--o{ USBDevice : "auto-approves"
    USBBackupConfig ||--o{ USBBackup : "configures"
    USBBackup ||--o{ USBRestore : "source for"

    USBDevice {
        string busID
        string nodeName
        string vendorID
        string productID
        string serialNumber
        string phase "PendingApproval|Approved|Denied|Disconnected"
    }
    USBDeviceApproval {
        string deviceRef
        string requester
        string phase "Pending|Approved|Denied"
        string approvedBy
        time expiresAt
    }
    USBDevicePolicy {
        string vendorID "selector"
        string productID "selector"
        string mode "manual|auto"
        bool autoApproveKnownDevices
        bool denyHumanInterfaceDevices
    }
    USBConnection {
        string deviceRef
        string clientNode
        string phase "Pending|Connecting|Connected|Failed"
        string clientDevicePath
    }
    USBDeviceWhitelist {
        string entries "fingerprint list"
        int32 entryCount
    }
    USBBackupConfig {
        string schedule
        int32 retentionCount
        string destinationType "pvc|configmap|s3"
        bool autoRestoreEnabled
    }
    USBBackup {
        string triggerType "manual|scheduled"
        string phase "InProgress|Completed|Failed"
        string checksum "sha256"
        string size
    }
    USBRestore {
        string backupRef
        string triggerType "manual|automatic"
        string phase "Validating|Restoring|RevalidatingConnections|Completed|Failed"
        bool dryRun
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

### USBBackup Lifecycle

```mermaid
stateDiagram-v2
    [*] --> InProgress: Backup created
    InProgress --> Completed: Snapshot written + checksum computed
    InProgress --> Failed: Storage error or resource collection failure
    Completed --> [*]: Retained or deleted by retention policy
```

### USBRestore Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Validating: Restore created
    Validating --> Restoring: Backup exists + checksum valid
    Validating --> Completed: DryRun mode (validation only)
    Validating --> Failed: Backup not found or checksum mismatch
    Restoring --> RevalidatingConnections: Resources applied
    RevalidatingConnections --> Completed: All connections validated
    RevalidatingConnections --> Completed: Invalid connections terminated
    Restoring --> Failed: Apply error
```

## Backup/Restore Architecture

```mermaid
flowchart TD
    subgraph Backup Flow
        A[BackupReconciler] -->|List| B[Whitelists + Policies + Approvals]
        B -->|CreateSnapshot| C[Snapshot with SHA-256 checksum]
        C -->|Write| D[BackupStorage Interface]
        D --> E[ConfigMap Storage ✅]
        D --> F[PVC Storage ⚠️ stub]
        D --> G[S3 Storage ⚠️ stub]
    end

    subgraph Restore Flow
        H[RestoreReconciler] -->|Phase 1| I[Validate backup + checksum]
        I -->|Phase 2| J[Delete + recreate CRs from snapshot]
        J -->|Phase 3| K[Revalidate all USBConnections]
        K -->|Phase 4| L[Mark Completed]
    end

    subgraph Health Monitor
        M[HealthMonitor.Check] -->|Unhealthy?| N[MaybeTriggerAutoRestore]
        N -->|Cooldown OK?| O[Create USBRestore CR]
        N -->|Max retries?| P[Give up]
    end
```

## Security Model

- Manual approval by default (`PendingApproval` → `Approved`)
- Policy whitelist/blacklist controls via `USBDevicePolicy` selector
- Auto-approve known devices via fingerprint whitelist
- Optional mTLS encryption for USB/IP tunnels (`requireEncryption` flag)
- Network isolation via automatic `NetworkPolicy` generation (planned)
- HID device class blocking (`denyHumanInterfaceDevices`)
- Namespace-scoped connections with allowed-namespace restrictions
- Finalizer-based cleanup for exported devices and tunnel teardown
- Max concurrent connections limit per device
- Backup integrity via SHA-256 checksums
- Auto-restore with cooldown (10min) and retry limits (max 3)
