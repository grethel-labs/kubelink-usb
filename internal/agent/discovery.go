package agent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// Discovery watches local device paths and emits add/remove/change events.
type Discovery struct {
	watcher *fsnotify.Watcher
	logger  *log.Logger
	paths   []string
}

// DiscoveryEventType is the normalized event category.
type DiscoveryEventType string

const (
	DiscoveryEventAdd    DiscoveryEventType = "add"
	DiscoveryEventRemove DiscoveryEventType = "remove"
	DiscoveryEventChange DiscoveryEventType = "change"
)

// NewDiscovery creates a watcher for device paths commonly used for USB serial devices.
func NewDiscovery(logger *log.Logger) (*Discovery, error) {
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
			d.logger.Printf("udev event=%s path=%s raw=%s", eventTypeFromOp(event.Op), event.Name, event.Op.String())
		}
	}
}

func looksLikeUSBDevicePath(path string) bool {
	base := filepath.Base(path)
	return base == "by-id" || strings.HasPrefix(base, "ttyUSB") || strings.HasPrefix(base, "ttyACM")
}
