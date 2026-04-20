package agent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/grethel-labs/kubelink-usb/internal/metrics"
)

// Discovery watches local device paths and emits add/remove/change events.
// It monitors /dev, /dev/serial, and /dev/serial/by-id using fsnotify for
// real-time USB device hotplug detection. Events are normalized into a
// simple add/remove/change taxonomy before being dispatched to the sink.
//
// @component Discovery["fsnotify Watcher"] --> Bridge["Discovery→CR Bridge"]
// @flow WatchPaths["Watch /dev paths"] --> NormalizeEvent["Normalize event type"]
// @flow NormalizeEvent --> FilterUSB{"USB path?"}
// @flow FilterUSB -->|yes| DispatchSink["Dispatch to sink"]
// @flow FilterUSB -->|no| Ignore["Ignore"]
type Discovery struct {
	watcher *fsnotify.Watcher
	logger  *log.Logger
	paths   []string
	sink    DiscoveryEventSink
}

// DiscoveryEventType is the normalized event category.
type DiscoveryEventType string

// DiscoveryEventType constants define the normalized event categories.
const (
	DiscoveryEventAdd    DiscoveryEventType = "add"
	DiscoveryEventRemove DiscoveryEventType = "remove"
	DiscoveryEventChange DiscoveryEventType = "change"
)

// DiscoveryEvent is a normalized watcher event.
type DiscoveryEvent struct {
	Type DiscoveryEventType
	Path string
}

// DiscoveryEventSink receives normalized discovery events.
type DiscoveryEventSink interface {
	OnDiscoveryEvent(context.Context, DiscoveryEvent) error
}

// NewDiscovery creates a watcher for device paths commonly used for USB serial devices.
func NewDiscovery(logger *log.Logger) (*Discovery, error) {
	return NewDiscoveryWithSink(logger, nil)
}

// NewDiscoveryWithSink creates a watcher with an optional event sink callback.
func NewDiscoveryWithSink(logger *log.Logger, sink DiscoveryEventSink) (*Discovery, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create fsnotify watcher: %w", err)
	}
	if logger == nil {
		logger = log.Default()
	}

	d := &Discovery{
		watcher: watcher,
		logger:  logger,
		sink:    sink,
		paths: []string{
			"/dev",
			"/dev/serial",
			"/dev/serial/by-id",
		},
	}

	if err := d.addPaths(); err != nil {
		watcher.Close()
		return nil, err
	}
	return d, nil
}

func (d *Discovery) addPaths() error {
	for _, p := range d.paths {
		if _, err := os.Stat(p); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("stat %s: %w", p, err)
		}
		if err := d.watcher.Add(p); err != nil {
			return fmt.Errorf("watch %s: %w", p, err)
		}
	}
	return nil
}

func eventTypeFromOp(op fsnotify.Op) DiscoveryEventType {
	switch {
	case op&fsnotify.Create == fsnotify.Create:
		return DiscoveryEventAdd
	case op&fsnotify.Remove == fsnotify.Remove:
		return DiscoveryEventRemove
	default:
		return DiscoveryEventChange
	}
}

// Run blocks until context cancellation and logs normalized USB-related events.
func (d *Discovery) Run(ctx context.Context) error {
	defer d.watcher.Close()
	d.logger.Printf("usb discovery watcher started on %v", d.paths)

	for {
		select {
		case <-ctx.Done():
			d.logger.Printf("usb discovery watcher stopping: %v", ctx.Err())
			return nil
		case err, ok := <-d.watcher.Errors:
			if !ok {
				return nil
			}
			d.logger.Printf("udev watch error: %v", err)
		case event, ok := <-d.watcher.Events:
			if !ok {
				return nil
			}
			if !looksLikeUSBDevicePath(event.Name) {
				continue
			}
			discoveryEvent := DiscoveryEvent{
				Type: eventTypeFromOp(event.Op),
				Path: event.Name,
			}
			metrics.ObserveDiscoveryEvent(string(discoveryEvent.Type))
			d.logger.Printf("udev event=%s path=%s raw=%s", discoveryEvent.Type, event.Name, event.Op.String())
			if d.sink != nil {
				if err := d.sink.OnDiscoveryEvent(ctx, discoveryEvent); err != nil {
					d.logger.Printf("discovery sink error for path=%s: %v", event.Name, err)
					timer := time.NewTimer(time.Second)
					select {
					case <-ctx.Done():
						timer.Stop()
					case <-timer.C:
						if retryErr := d.sink.OnDiscoveryEvent(ctx, discoveryEvent); retryErr != nil {
							d.logger.Printf("discovery sink retry failed for path=%s: %v", event.Name, retryErr)
						}
					}
				}
			}
		}
	}
}

func looksLikeUSBDevicePath(path string) bool {
	base := filepath.Base(path)
	return base == "by-id" || strings.HasPrefix(base, "ttyUSB") || strings.HasPrefix(base, "ttyACM")
}
