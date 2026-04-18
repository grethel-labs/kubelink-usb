- **Issue #1: Implement USBConnection Reconciler (Tunnel Management)**  
  Labels: `enhancement`, `help wanted`  
  Acceptance: Connection CRD wird erstellt, Agent auf Server-Node führt usbip bind aus, Agent auf Client-Node führt usbip attach aus.

- **Issue #2: Security Layer - Approval Workflow**  
  Labels: `enhancement`, `help wanted`  
  Acceptance: USBDeviceApproval CRD wird bei neuem Gerät automatisch erstellt, Controller wartet auf Approval bevor Connection erlaubt wird.

- **Issue #3: Network Isolation & Encryption**  
  Labels: `enhancement`, `help wanted`  
  Acceptance: Optionales mTLS für USB/IP Tunnel, NetworkPolicy automatische Erstellung.

- **Issue #4: Resilience - Reconnect Logic**  
  Labels: `enhancement`, `bug`  
  Acceptance: Bei Netzwerk-Ausfall: Retry mit Backoff, bei permanentem Verlust: Status auf Failed, Event an User.

- **Issue #5: Webhook Implementation**  
  Labels: `enhancement`, `help wanted`  
  Acceptance: Validation Webhook für Policies (regex für VID/PID), Mutation Webhook für Default-Werte.

- **Issue #6: Device Hotplug Handling**  
  Labels: `enhancement`, `bug`  
  Acceptance: Gerät rausziehen → Status Disconnected, gleiches Gerät reinstecken (erkennen via Serial) → Reconnect.

- **Issue #7: CLI Tool (kubectl-usb)**  
  Labels: `enhancement`, `help wanted`  
  Acceptance: `kubectl usb approve <device>`, `kubectl usb list`, `kubectl usb connect <device> <pod>`.

- **Issue #8: Metrics & Observability**  
  Labels: `enhancement`, `help wanted`  
  Acceptance: Prometheus Metrics für aktive Tunnel, Fehlerraten, Device-Discovery-Rate.

- **Issue #9: Documentation & Examples**  
  Labels: `enhancement`, `help wanted`  
  Acceptance: Komplettes Setup-Guide für k0s/MicroK8s, Beispiel: Zigbee2MQTT über 2 Nodes.

- **Issue #10: Multi-Architecture Support**  
  Labels: `enhancement`, `help wanted`  
  Acceptance: ARM64 Support (Raspberry Pi), amd64, Container Images für beide.
