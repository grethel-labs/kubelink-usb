package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	registerOnce sync.Once
	metricsMu    sync.Mutex

	discoveryEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "kubelink_usb_discovery_events_total", Help: "Total discovery watcher events by type."},
		[]string{"event_type"},
	)
	phaseTransitionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "kubelink_usb_phase_transitions_total", Help: "Total phase transitions by component."},
		[]string{"component", "from", "to"},
	)
	usbDevicesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "kubelink_usb_devices", Help: "Observed USB devices by phase."},
		[]string{"phase"},
	)
	usbConnectionsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "kubelink_usb_connections", Help: "Observed USB connections by phase."},
		[]string{"phase"},
	)
	approvalLatencySeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "kubelink_usb_approval_latency_seconds",
			Help:    "Latency between approval request creation and reconciliation completion.",
			Buckets: prometheus.DefBuckets,
		},
	)

	devicePhaseCounts     = map[string]int{}
	connectionPhaseCounts = map[string]int{}
)

// Register ensures custom metrics are registered exactly once.
func Register() {
	registerOnce.Do(func() {
		registerMetric(discoveryEventsTotal)
		registerMetric(phaseTransitionsTotal)
		registerMetric(usbDevicesGauge)
		registerMetric(usbConnectionsGauge)
		registerMetric(approvalLatencySeconds)
	})
}

func registerMetric(collector prometheus.Collector) {
	if err := ctrlmetrics.Registry.Register(collector); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			panic(err)
		}
	}
}

// ObserveDiscoveryEvent increments discovery event counters.
func ObserveDiscoveryEvent(eventType string) {
	Register()
	discoveryEventsTotal.WithLabelValues(eventType).Inc()
}

// ObserveApprovalDuration records approval processing latency.
func ObserveApprovalDuration(duration time.Duration) {
	Register()
	approvalLatencySeconds.Observe(duration.Seconds())
}

// ObservePhaseTransition records transition counts.
func ObservePhaseTransition(component, from, to string) {
	Register()
	if from == to {
		return
	}
	phaseTransitionsTotal.WithLabelValues(component, emptyToUnknown(from), emptyToUnknown(to)).Inc()
}

// UpdateDevicePhase updates device phase gauges and transition counters.
func UpdateDevicePhase(from, to string) {
	Register()
	updatePhaseGauge(devicePhaseCounts, usbDevicesGauge, "device", from, to)
}

// UpdateConnectionPhase updates connection phase gauges and transition counters.
func UpdateConnectionPhase(from, to string) {
	Register()
	updatePhaseGauge(connectionPhaseCounts, usbConnectionsGauge, "connection", from, to)
}

func updatePhaseGauge(phaseCounts map[string]int, gauge *prometheus.GaugeVec, component, from, to string) {
	if from == to {
		return
	}

	metricsMu.Lock()
	defer metricsMu.Unlock()

	if from != "" {
		if phaseCounts[from] > 0 {
			phaseCounts[from]--
		}
		gauge.WithLabelValues(from).Set(float64(phaseCounts[from]))
	}
	if to != "" {
		phaseCounts[to]++
		gauge.WithLabelValues(to).Set(float64(phaseCounts[to]))
	}

	ObservePhaseTransition(component, from, to)
}

// RecordPhaseTransitionEvent emits Kubernetes events for phase transitions.
func RecordPhaseTransitionEvent(recorder record.EventRecorder, obj runtime.Object, component, from, to string) {
	if recorder == nil || obj == nil || from == to {
		return
	}
	recorder.Eventf(obj, corev1.EventTypeNormal, "PhaseTransition", "%s phase transitioned from %s to %s", component, emptyToUnknown(from), emptyToUnknown(to))
}

func emptyToUnknown(v string) string {
	if v == "" {
		return "unknown"
	}
	return v
}
