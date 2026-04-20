package metrics

import (
	"testing"
	"time"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
)

func TestObserveDiscoveryEventIncrementsCounter(t *testing.T) {
	before := testutil.ToFloat64(discoveryEventsTotal.WithLabelValues("add"))
	ObserveDiscoveryEvent("add")
	after := testutil.ToFloat64(discoveryEventsTotal.WithLabelValues("add"))
	if after-before != 1 {
		t.Fatalf("counter delta = %v, want 1", after-before)
	}
}

func TestUpdateDevicePhaseAdjustsGauge(t *testing.T) {
	metricsMu.Lock()
	devicePhaseCounts = map[string]int{}
	usbDevicesGauge.WithLabelValues("PendingApproval").Set(0)
	usbDevicesGauge.WithLabelValues("Disconnected").Set(0)
	metricsMu.Unlock()

	UpdateDevicePhase("", "PendingApproval")
	UpdateDevicePhase("PendingApproval", "Disconnected")

	if got := testutil.ToFloat64(usbDevicesGauge.WithLabelValues("PendingApproval")); got != 0 {
		t.Fatalf("pending gauge = %v, want 0", got)
	}
	if got := testutil.ToFloat64(usbDevicesGauge.WithLabelValues("Disconnected")); got != 1 {
		t.Fatalf("disconnected gauge = %v, want 1", got)
	}
}

func TestObserveApprovalDurationIncrementsHistogram(t *testing.T) {
	before := &dto.Metric{}
	if err := approvalLatencySeconds.Write(before); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	ObserveApprovalDuration(2 * time.Second)

	after := &dto.Metric{}
	if err := approvalLatencySeconds.Write(after); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if after.GetHistogram().GetSampleCount() <= before.GetHistogram().GetSampleCount() {
		t.Fatalf("histogram sample count did not increase")
	}
}

func TestRecordPhaseTransitionEvent(t *testing.T) {
	recorder := record.NewFakeRecorder(2)
	obj := &usbv1alpha1.USBDevice{ObjectMeta: metav1.ObjectMeta{Name: "dev-a"}}
	RecordPhaseTransitionEvent(recorder, obj, "device", "Approved", "Disconnected")

	select {
	case event := <-recorder.Events:
		if event == "" {
			t.Fatal("expected event string")
		}
	case <-time.After(time.Second):
		t.Fatal("expected phase transition event")
	}
}
