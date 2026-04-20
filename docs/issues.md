# KubeLink-USB — Issue Tracker

> Priorisierte Issues für die weitere Entwicklung. Stand: April 2026.

## Kritischer Pfad (v1.0 MVP)

- **Issue #2: Discovery-zu-CR-Bridge** ⬆️ HÖCHSTE PRIORITÄT
  Labels: `enhancement`, `phase-1`, `blocking`
  Status: 🔶 Teilweise (Discovery loggt Events, aber erstellt noch keine K8s-CRs)
  Beschreibung: Agent erstellt automatisch USBDevice-CRs bei Discovery-Events.
  Akzeptanz: `add`→CR erstellen, `remove`→Phase=Disconnected, Reconnect-Erkennung via Serial.
  Verbleibend: K8s-Client-Initialisierung im Agent, Event-Callback mit CR-Erstellung.

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

- **Issue #18: S3 Backup Storage (Real SDK)**
  Labels: `enhancement`
  Status: ⚠️ Mock (In-Memory, kein echtes S3 SDK)
  Beschreibung: `S3Storage` Write/Read/List/Delete mit echtem AWS SDK implementieren.
  Akzeptanz: Backups in S3-Bucket persistiert, Roundtrip-Test.

- **Issue #19: Helm Chart**
  Labels: `enhancement`, `phase-9`
  Status: ❌ Nicht begonnen
  Beschreibung: Helm Chart für einfaches Cluster-Deployment.
  Akzeptanz: `helm install kubelink-usb` deployed Controller + Agent DaemonSet + CRDs.

## Bereits erledigt ✅

- ~~Issue #1: Device Fingerprinting~~ → DNS-label-safe, deterministic CR names ✅
- ~~Issue #3: Policy-Engine implementieren~~ → Vendor/Product/Node/HID/Class Matching ✅
- ~~Issue #4: Approval Workflow~~ → Approve/Deny/Expire mit Device-Phase-Propagation ✅
- ~~Issue #5: USB Connection Controller~~ → Tunnel-Lifecycle mit Finalizer ✅
- ~~Issue #6: Server-seitiger Export~~ → CommandRunner + usbipd bind/unbind ✅
- ~~Issue #7: Client-seitiger Import~~ → CommandRunner + usbip attach/detach ✅
- ~~Issue #8: Vollständiges USB/IP-Protokoll~~ → DevList/Import Frames + Server/Client ✅
- ~~Issue #17: PVC Backup Storage~~ → File-basiert mit 0o600 Permissions ✅
- ~~Issue: CRD API Types~~ → 8 Ressourcen mit DeepCopy ✅
- ~~Issue: USBDevice Reconciler~~ → Finalizer + Status-Init ✅
- ~~Issue: Discovery Watcher~~ → fsnotify + Event-Normalisierung ✅
- ~~Issue: Backup/Restore System~~ → Snapshot, Storage, Controller, HealthMonitor ✅
- ~~Issue: TLS Baseline~~ → TLS 1.3 Config ✅
- ~~Issue: Whitelist~~ → In-Memory Set ✅
- ~~Issue: CI Pipeline~~ → Lint, Test, Coverage, Build, Images, Docs, Publish ✅
