# KubeLink-USB — Implementierungsfortschritt & Roadmap

> Stand: April 2026 — Basierend auf tatsächlichem Code-Review aller Dateien.

---

## Gesamtfortschritt für v1.0

```
Gesamt: ██████████████████░░ ~90%

CRD-API-Typen:        ████████████████████ 100%  (8 Ressourcen, DeepCopy, Scheme-Registration)
USBDevice Controller:  ████████████████████ 100%  (Finalizer, Status-Init, Deletion-Handling)
Backup/Restore:        ████████████████████ 100%  (Snapshot, Storage, Controller, HealthMonitor)
Discovery Watcher:     ████████████████████ 100%  (fsnotify, Event-Normalisierung, Pfad-Filter)
TLS + Whitelist:       ████████████████████ 100%  (TLS 1.3 Config, In-Memory-Set)
Device Fingerprinting: ████████████████████ 100%  (DNS-safe, deterministic CR names)
Policy Engine:         ████████████████████ 100%  (Vendor/Product/Node/HID/Class matching)
Approval Controller:   ████████████████████ 100%  (Approve/Deny/Expire, device phase propagation)
Connection Controller: ████████████████████ 100%  (Tunnel lifecycle: Pending→Connecting→Connected→Failed)
Agent Export/Import:   ████████████████████ 100%  (CommandRunner interface, usbipd/usbip exec)
USB/IP Protocol:       ████████████████████ 100%  (DevList + Import frames, TCP server/client)
Discovery→CR Bridge:   ████████░░░░░░░░░░░░  40%  (Discovery logs only, K8s-Client-Integration in Agent pending)
```

## Aktuelle Coverage-Zahlen

| Package | Coverage | CI-Minimum | Ziel |
|---------|----------|------------|------|
| **Gesamt** | **81.0%** | 80% | 85% |
| `api/v1alpha1` | 98.9% | 80% | 80% |
| `internal/security` | 94.3% | 80% | 90% |
| `internal/usbip` | 57.2% | 50% | 75% |
| `internal/utils` | 100.0% | 80% | 90% |
| `internal/backup` | 91.2% | — | 85% |
| `internal/controller` | ~70% | — | 85% |
| `internal/agent` | ~69% | — | 80% |
| `cmd/*` | 0.0% | — | — |

**Tests:** 81+ Testfunktionen in 15+ Dateien

---

## Was ist fertig ✅

### CRD-API-Typen (8 Ressourcen)
- [x] `USBDevice` — Geräte-Discovery-CR (Phase: PendingApproval→Approved→Connected→Disconnected)
- [x] `USBDeviceApproval` — Manuelle/Auto-Genehmigungen mit Ablaufzeit
- [x] `USBDevicePolicy` — Policy-Regeln (Vendor/Product/Node-Selektoren, Restrictions)
- [x] `USBConnection` — Tunnel-Lifecycle-CR (Phase: Pending→Connecting→Connected→Failed)
- [x] `USBDeviceWhitelist` — Bekannte sichere Geräte-Registry (Fingerprint-basiert)
- [x] `USBBackupConfig` — Backup-Speicherziel-Konfiguration (PVC/ConfigMap/S3)
- [x] `USBBackup` — Backup-Anfrage + Ergebnis (Phase: InProgress→Completed/Failed)
- [x] `USBRestore` — Restore-Anfrage mit DryRun + Health-Validierung
- [x] DeepCopy für alle Typen generiert und getestet (6 Tests)
- [x] Scheme-Registration in `groupversion_info.go`

### USBDevice Reconciler
- [x] Finalizer `kubelink-usb.io/cleanup-export` automatisch setzen
- [x] Status initialisieren: Phase=PendingApproval, LastSeen=now(), Health=Healthy
- [x] Deletion-Handling: Finalizer entfernen bei DeletionTimestamp
- [x] Tests: 3 Testfunktionen (fake-client-basiert)

### Device Fingerprinting
- [x] `DeviceFingerprint(nodeName, vendorID, productID, serialNumber, busID)` — DNS-label-safe Namen
- [x] `sanitizeDNSLabel()` — Lowercase, unsafe-char-Ersetzung, Dash-Collapse, 63-Char-Limit
- [x] BusID-basierter Fallback für Geräte ohne Seriennummer
- [x] Tests: 8 Table-Driven (normale Eingaben, Sonderzeichen, leere Felder, Truncation)

### Policy Engine
- [x] `Engine.Allows()` — Vendor/Product/Node-Selector-Matching
- [x] `allowedNodes`-Check
- [x] `allowedDeviceClasses`-Check
- [x] `denyHumanInterfaceDevices`-Check (HID, 03, 0x03)
- [x] `MatchesSelector()` — Prüft ob Policy auf Device zutrifft
- [x] Case-insensitive Matching für alle Felder
- [x] Tests: 15 Table-Driven (Match/Mismatch/HID/Node-Deny/Class-Filter)

### Approval Controller
- [x] `USBDeviceApproval` verarbeiten
- [x] Device-Phase von PendingApproval → Approved/Denied
- [x] Ablaufzeit-Prüfung (expiresAt)
- [x] Fehlende Device-Erkennung → Denied
- [x] Already-processed Skip (Idempotenz)
- [x] Tests: 6 Fake-Client (Approve, Deny, Expired, Missing Device, NotFound, AlreadyProcessed)

### USB Connection Controller
- [x] Finalizer `kubelink-usb.io/cleanup-tunnel`
- [x] Phase-Transitions: Pending → Connecting → Connected → Failed
- [x] Device-Approval-Check vor Verbindung
- [x] TunnelInfo aus Device.ConnectionInfo befüllen
- [x] Deletion-Handling: Finalizer entfernen bei DeletionTimestamp
- [x] Tests: 5 Fake-Client (Init, Unapproved, Connected, NotFound, Deletion)

### Agent Server (Export/Unexport)
- [x] `Export(ctx, busID)` — führt `usbipd bind --busid` aus
- [x] `Unexport(ctx, busID)` — führt `usbipd unbind --busid` aus
- [x] `CommandRunner` Interface für Testbarkeit (Mock-basiert)
- [x] Input-Validierung (leere BusID)
- [x] Tests: 5 (Success, EmptyBusID, CommandFailure für Export+Unexport)

### Agent Client (Attach/Detach)
- [x] `Attach(ctx, remote, busID)` — führt `usbip attach` aus
- [x] `Detach(ctx, port)` — führt `usbip detach` aus
- [x] `parseDevicePath()` — extrahiert /dev/ttyUSB* oder /dev/ttyACM* aus Output
- [x] Input-Validierung (leere Remote/BusID/Port)
- [x] Tests: 8 (Success, EmptyRemote, EmptyBusID, CommandFailure, NoDevicePath, Detach)

### USB/IP Protokoll
- [x] `BasicHeader` Struct (Version + Code + Status, Big-Endian)
- [x] `DevListRequest/Response` — Encode/Decode mit Geräteliste
- [x] `ImportRequest/Response` — Encode/Decode mit BusID + Device-Info
- [x] Operation-Codes: OPReqDevList, OPRepDevList, OPReqImport, OPRepImport
- [x] Max-Device-Limit (256) gegen Overflow
- [x] Tests: 7 (DevList Roundtrip, Import Roundtrip, BasicHeader, Truncated, BadOpcode)

### USB/IP Server
- [x] TCP-Listener mit Context-Cancellation
- [x] `DeviceProvider` Interface für Device-Listen
- [x] DevList-Request-Handler + Import-Request-Handler
- [x] Graceful Shutdown

### USB/IP Client
- [x] `Connect()` — Verbindung + DevList-Abfrage
- [x] `ListRemoteDevices()` — DevList Request/Response
- [x] `ImportDevice()` — Import Request/Response mit Status-Check
- [x] Integration-Test: Server↔Client DevList-Roundtrip

### Backup-System
- [x] `BackupStorage`-Interface (Write/Read/List/Delete)
- [x] `ConfigMapStorage` — Thread-safe In-Memory-Speicher
- [x] `PVCStorage` — File-basiert mit 0o600 Permissions
- [x] `S3Storage` — In-Memory-Mock (Interface vorhanden)
- [x] Snapshot-Envelope: JSON mit Version, CreatedAt, SHA-256 Checksum
- [x] Tests: 13 Testfunktionen

### Backup/Restore Controller + Health Monitor
- [x] Backup: Sammelt + Snapshot + Retention
- [x] Restore: Multi-Phase + DryRun + Revalidierung
- [x] Health Monitor: Consistency-Check + Auto-Restore
- [x] Tests: 19 Testfunktionen

### Discovery Watcher
- [x] fsnotify auf /dev, /dev/serial, /dev/serial/by-id
- [x] Event-Normalisierung + USB-Pfad-Filter + Graceful Shutdown
- [x] Tests: 5 Testfunktionen

### Security Baseline
- [x] TLS 1.3+ Config + In-Memory Whitelist
- [x] Tests: 3+ Testfunktionen

### CI/CD
- [x] GitHub Actions: lint → test → coverage → build → images → docs → publish
- [x] Coverage-Gate: 80% (aktuell 81.0%)

---

## Was fehlt für v1.0 ❌

### Discovery→CR Bridge (verbleibend ~1-2 Tage)
- [ ] Event-Callback mit K8s-Client in Discovery
- [ ] `add`-Event → `USBDevice`-CR erstellen (nutzt DeviceFingerprint)
- [ ] `remove`-Event → `USBDevice.Status.Phase = Disconnected`
- [ ] Reconnect-Erkennung via SerialNumber
- [ ] `cmd/agent/main.go` — K8s-Client-Initialisierung (in-cluster Config)
- [ ] Tests: Fake-Client-basierte CR-Erstellung

---

## Optionale Verbesserungen (v1.1+)

### Resilience & Lifecycle
- [ ] Reconnect-Logik (Retry mit konfiguriertem Backoff)
- [ ] Disconnect-Timeout
- [ ] Device-Hotplug-Handling

### Security & Encryption
- [ ] mTLS für USB/IP-Tunnel
- [ ] cert-manager-Integration
- [ ] Network Isolation (automatische NetworkPolicy)

### CLI & UI
- [ ] kubectl-usb Plugin

### Webhooks
- [ ] Validating/Mutating Webhooks

### Observability
- [ ] Prometheus Metrics
- [ ] Kubernetes Events

### Distribution
- [ ] Multi-Architecture Images (ARM64 + amd64)
- [ ] Helm Chart
- [ ] Real S3 Backup Storage

---

## Kritischer Pfad (verbleibend)

```
Verbleibendes für v1.0-MVP:     ~1-2 Tage Arbeit
└── Discovery→CR Bridge
    ├── K8s-Client in Agent initialisieren
    ├── fsnotify-Events → USBDevice-CRs erstellen/updaten
    └── Tests (Fake-Client)
```

Alle Kernfunktionalitäten (Phases 1-3) sind implementiert.
Die Discovery→CR Bridge ist das letzte fehlende Stück für den
vollständigen End-to-End-Flow vom USB-Einstecken bis zur Tunnel-Verbindung.

---

## Teststrategie

### Testarten pro Komponente

| Komponente | Art | Tests | Status |
|------------|-----|-------|--------|
| CRD DeepCopy | Unit | 6 | ✅ |
| Discovery | Unit | 5 | ✅ |
| Device Fingerprinting | Unit (Table) | 8 | ✅ |
| Agent Client | Unit (Mock) | 8 | ✅ |
| Agent Server | Unit (Mock) | 5 | ✅ |
| USB/IP Protocol | Unit | 7 | ✅ |
| USB/IP Client/Server | Integration | 1 | ✅ |
| Backup Snapshot | Unit | 8 | ✅ |
| Backup Storage | Unit | 5 | ✅ |
| Security (Policy Engine) | Unit (Table) | 15 | ✅ |
| Security (Whitelist+TLS) | Unit | 3 | ✅ |
| USBDevice Controller | Fake-Client | 3 | ✅ |
| Approval Controller | Fake-Client | 6 | ✅ |
| Connection Controller | Fake-Client | 5 | ✅ |
| Backup Controller | Fake-Client | 5 | ✅ |
| Restore Controller | Fake-Client | 6 | ✅ |
| Health Monitor | Fake-Client | 8 | ✅ |
| Utils (Net+Udev) | Unit | 2 | ✅ |

### CI-Gates

| Gate | Aktuell | Status |
|------|---------|--------|
| `make lint` | gofmt + go vet | ✅ Besteht |
| `make test` | 81+ Tests, alle grün | ✅ Besteht |
| `make coverage-check` | 81.0% ≥ 80% | ✅ Besteht |
| `make build` | bin/controller + bin/agent | ✅ Besteht |
| `make docs` + git diff | CODE_REFERENCE.md aktuell | ✅ Besteht |
