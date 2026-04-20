package agent

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
	"github.com/grethel-labs/kubelink-usb/internal/metrics"
	"github.com/grethel-labs/kubelink-usb/internal/utils"
)

// USBDeviceBridge syncs discovery events into USBDevice CRs.
type USBDeviceBridge struct {
	client   client.Client
	nodeName string
	logger   *log.Logger
}

// NewUSBDeviceBridge returns a discovery sink that bridges events to USBDevice CRs.
func NewUSBDeviceBridge(kubeClient client.Client, nodeName string, logger *log.Logger) *USBDeviceBridge {
	if logger == nil {
		logger = log.Default()
	}
	return &USBDeviceBridge{client: kubeClient, nodeName: nodeName, logger: logger}
}

// OnDiscoveryEvent handles add/remove discovery events.
func (b *USBDeviceBridge) OnDiscoveryEvent(ctx context.Context, event DiscoveryEvent) error {
	if b.client == nil {
		return fmt.Errorf("kubernetes client is not configured")
	}

	switch event.Type {
	case DiscoveryEventAdd:
		return b.handleAdd(ctx, event.Path)
	case DiscoveryEventRemove:
		return b.handleRemove(ctx, event.Path)
	default:
		return nil
	}
}

func (b *USBDeviceBridge) handleAdd(ctx context.Context, path string) error {
	serial := serialFromPath(path)
	busID := filepath.Base(path)
	name := utils.DeviceFingerprint(b.nodeName, "0000", "0000", serial, busID)

	var existing usbv1alpha1.USBDevice
	err := b.client.Get(ctx, types.NamespacedName{Name: name}, &existing)
	if apierrors.IsNotFound(err) {
		device := &usbv1alpha1.USBDevice{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: usbv1alpha1.USBDeviceSpec{
				BusID:        busID,
				DevicePath:   path,
				NodeName:     b.nodeName,
				VendorID:     "0000",
				ProductID:    "0000",
				SerialNumber: serial,
			},
		}
		if createErr := b.client.Create(ctx, device); createErr != nil {
			return fmt.Errorf("create usbdevice %s: %w", name, createErr)
		}
		b.logger.Printf("created usbdevice %s for path=%s", name, path)
		return nil
	}
	if err != nil {
		return err
	}

	existing.Spec.BusID = busID
	existing.Spec.DevicePath = path
	existing.Spec.NodeName = b.nodeName
	existing.Spec.SerialNumber = serial
	if updateErr := b.client.Update(ctx, &existing); updateErr != nil {
		return fmt.Errorf("update usbdevice %s: %w", name, updateErr)
	}

	if existing.Status.Phase == "Disconnected" {
		oldPhase := existing.Status.Phase
		now := metav1.Now()
		existing.Status.Phase = "PendingApproval"
		existing.Status.LastSeen = &now
		if statusErr := b.client.Status().Update(ctx, &existing); statusErr != nil {
			return fmt.Errorf("update usbdevice status %s: %w", name, statusErr)
		}
		metrics.UpdateDevicePhase(oldPhase, existing.Status.Phase)
	}

	b.logger.Printf("updated usbdevice %s for path=%s", name, path)
	return nil
}

func (b *USBDeviceBridge) handleRemove(ctx context.Context, path string) error {
	var deviceList usbv1alpha1.USBDeviceList
	if err := b.client.List(ctx, &deviceList); err != nil {
		return fmt.Errorf("list usbdevices: %w", err)
	}

	removedSerial := serialFromPath(path)
	for i := range deviceList.Items {
		device := &deviceList.Items[i]
		if !matchesRemovedPath(device, path, removedSerial) {
			continue
		}
		oldPhase := device.Status.Phase
		now := metav1.Now()
		device.Status.Phase = "Disconnected"
		device.Status.LastSeen = &now
		if err := b.client.Status().Update(ctx, device); err != nil {
			return fmt.Errorf("set disconnected for %s: %w", device.Name, err)
		}
		metrics.UpdateDevicePhase(oldPhase, device.Status.Phase)
		b.logger.Printf("marked usbdevice %s disconnected for path=%s", device.Name, path)
	}
	return nil
}

func serialFromPath(path string) string {
	if strings.Contains(path, "/dev/serial/by-id/") {
		return filepath.Base(path)
	}
	return ""
}

func matchesRemovedPath(device *usbv1alpha1.USBDevice, path, serial string) bool {
	if device.Spec.DevicePath == path {
		return true
	}
	if serial != "" && device.Spec.SerialNumber == serial {
		return true
	}
	return filepath.Base(device.Spec.DevicePath) == filepath.Base(path)
}
