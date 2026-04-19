# KubeLink-USB — Issue Tracker

> Priorisierte Issues für die weitere Entwicklung. Stand: April 2026.

## Kritischer Pfad (v1.0 MVP)

- **Issue #1: Device Fingerprinting** ⬆️ HÖCHSTE PRIORITÄT
  Labels: `enhancement`, `phase-1`, `blocking`
  Status: ❌ Nicht begonnen
  Beschreibung: `DeviceFingerprint()` Funktion in `internal/utils/` für deterministische, DNS-konforme CR-Namen.
  Akzeptanz: Gleiche Geräte erzeugen immer denselben Namen, BusID-Fallback für Geräte ohne Serial.
  Blockiert: Issue #2 (Discovery→CR Bridge)

- **Issue #2: Discovery-zu-CR-Bridge** ⬆️ HÖCHSTE PRIORITÄT
  Labels: `enhancement`, `phase-1`, `blocking`
  Status: ❌ Nicht begonnen
  Beschreibung: Agent erstellt automatisch USBDevice-CRs bei Discovery-Events.
  Akzeptanz: `add`→CR erstellen, `remove`→Phase=Disconnected, Reconnect-Erkennung via Serial.
  Blockiert: Alle weiteren Phasen
  Abhängig von: Issue #1

- **Issue #3: Policy-Engine implementieren**
  Labels: `enhancement`, `security`, `phase-2`
  Status: ⚠️ Stub (Allows() → true)
  Beschreibung: VendorID/ProductID/Node-Selector-Matching, Restriction-Auswertung, HID-Blocking.
  Akzeptanz: Policy-Selektoren matchen korrekt, denyHumanInterfaceDevices blockiert HID.

- **Issue #4: Approval Workflow implementieren**
  Labels: `enhancement`, `phase-2`
  Status: ⚠️ Stub (No-Op Controller)
  Beschreibung: ApprovalReconciler verarbeitet USBDeviceApproval-CRs und setzt Device-Phase.
  Akzeptanz: Approve→Approved, Deny→Denied, Ablaufzeit-Prüfung, Auto-Approve via Whitelist.

- **Issue #5: USB Connection Controller implementieren**
  Labels: `enhancement`, `phase-3`
  Status: ⚠️ Stub (No-Op Controller)
  Beschreibung: Tunnel-Lifecycle orchestrieren: Export→Attach→Status-Update mit Finalizer-Cleanup.
  Akzeptanz: Phase-Transitions Pending→Connecting→Connected→Failed, TunnelInfo befüllt.

- **Issue #6: Server-seitiger Export (usbipd bind)**
  Labels: `enhancement`, `phase-3`
  Status: ⚠️ Stub (nil return)
  Beschreibung: Export/Unexport via os/exec-basierte usbipd-Aufrufe auf Source-Node.
  Akzeptanz: Export führt `usbipd bind` aus, Fehler korrekt propagiert.

- **Issue #7: Client-seitiger Import (usbip attach)**
  Labels: `enhancement`, `phase-3`
  Status: ⚠️ Stub (nil return)
  Beschreibung: Attach/Detach via os/exec auf Client-Node.
  Akzeptanz: Attach gibt Device-Path zurück, Detach entfernt VHCI-Port.

- **Issue #8: Vollständiges USB/IP-Protokoll**
  Labels: `enhancement`, `phase-3`
  Status: 🔶 Partial (nur BasicHeader)
  Beschreibung: DevList/Import Request/Response Frames, Transfer Submissions, Server+Client.
  Akzeptanz: Vollständiger In-Process Device-List-Exchange.

## Verbesserungen (v1.1+)

- **Issue #9: Network Isolation & Encryption**
  Labels: `enhancement`, `security`, `phase-5`
  Status: ❌ Nicht begonnen
  Beschreibung: Optionales mTLS für USB/IP-Tunnel, NetworkPolicy automatische Erstellung.
  Akzeptanz: `requireEncryption: true` → mTLS, `networkIsolation: true` → NetworkPolicy.

- **Issue #10: Resilience - Reconnect Logic**
  Labels: `enhancement`, `phase-4`
  Status: ❌ Nicht begonnen
  Beschreibung: Retry mit Backoff, Disconnect-Timeout, Hotplug-Handling.
  Akzeptanz: Netzwerkausfall → automatische Retries, permanenter Verlust → Failed + Event.

- **Issue #11: Device Hotplug Handling**
  Labels: `enhancement`, `phase-4`
  Status: ❌ Nicht begonnen
  Beschreibung: Gerät rausziehen → Disconnected, reinstecken (via Serial) → Reconnect.
  Akzeptanz: Automatischer Reconnect via SerialNumber-Match.

- **Issue #12: Webhook Implementation**
  Labels: `enhancement`, `phase-7`
  Status: ❌ Nicht begonnen
  Beschreibung: Validation Webhook für Policies (VID/PID-Format), Mutation Webhook für Defaults.
  Akzeptanz: Ungültige VendorID → Reject, fehlende Felder → Defaults.

- **Issue #13: CLI Tool (kubectl-usb)**
  Labels: `enhancement`, `phase-6`
  Status: ❌ Nicht begonnen
  Beschreibung: kubectl-Plugin: `list`, `approve`, `deny`, `connect`, `disconnect`.
  Akzeptanz: Jeder Befehl mit Standard-Kubeconfig, Tabellenausgabe.

- **Issue #14: Metrics & Observability**
  Labels: `enhancement`, `phase-8`
  Status: ❌ Nicht begonnen
  Beschreibung: Prometheus Metrics + Kubernetes Events für Statusübergänge.
  Akzeptanz: Gauges für Devices/Connections, Counter für Discovery, Histogram für Approval.

- **Issue #15: Documentation & Examples**
  Labels: `enhancement`, `documentation`
  Status: 🔶 Partial
  Beschreibung: Setup-Guide für k0s/MicroK8s, Beispiel: Zigbee2MQTT über 2 Nodes.
  Akzeptanz: Komplettes Tutorial mit funktionierendem Beispiel.

- **Issue #16: Multi-Architecture Support**
  Labels: `enhancement`, `phase-9`
  Status: ❌ Nicht begonnen
  Beschreibung: ARM64 (Raspberry Pi) + amd64 Container Images.
  Akzeptanz: Buildx Multi-Platform, GHCR-Publishing für beide Architekturen.

- **Issue #17: PVC Backup Storage**
  Labels: `enhancement`
  Status: ⚠️ Stub (Interface only)
  Beschreibung: `PVCStorage` Write/Read/List/Delete implementieren.
  Akzeptanz: Backups auf PVC-Mount persistiert, Roundtrip-Test.

- **Issue #18: S3 Backup Storage**
  Labels: `enhancement`
  Status: ⚠️ Stub (Interface only)
  Beschreibung: `S3Storage` Write/Read/List/Delete implementieren.
  Akzeptanz: Backups in S3-Bucket persistiert, Roundtrip-Test.

- **Issue #19: Helm Chart**
  Labels: `enhancement`, `phase-9`
  Status: ❌ Nicht begonnen
  Beschreibung: Helm Chart für einfaches Cluster-Deployment.
  Akzeptanz: `helm install kubelink-usb` deployed Controller + Agent DaemonSet + CRDs.

## Bereits erledigt ✅

- ~~Issue: CRD API Types~~ → 8 Ressourcen mit DeepCopy ✅
- ~~Issue: USBDevice Reconciler~~ → Finalizer + Status-Init ✅
- ~~Issue: Discovery Watcher~~ → fsnotify + Event-Normalisierung ✅
- ~~Issue: Backup/Restore System~~ → Snapshot, Storage, Controller, HealthMonitor ✅
- ~~Issue: TLS Baseline~~ → TLS 1.3 Config ✅
- ~~Issue: Whitelist~~ → In-Memory Set ✅
- ~~Issue: USB/IP BasicHeader~~ → Encode/Decode ✅
- ~~Issue: CI Pipeline~~ → Lint, Test, Coverage, Build, Images, Docs, Publish ✅
