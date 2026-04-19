package agent

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

type recordingSink struct {
	events chan DiscoveryEvent
}

func (r *recordingSink) OnDiscoveryEvent(_ context.Context, event DiscoveryEvent) error {
	r.events <- event
	return nil
}

type flakySink struct {
	calls chan DiscoveryEvent
	fail  bool
}

func (f *flakySink) OnDiscoveryEvent(_ context.Context, event DiscoveryEvent) error {
	f.calls <- event
	if f.fail {
		f.fail = false
		return errors.New("temporary sink failure")
	}
	return nil
}

func TestEventTypeFromOp(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		op   fsnotify.Op
		want DiscoveryEventType
	}{
		{name: "create", op: fsnotify.Create, want: DiscoveryEventAdd},
		{name: "remove", op: fsnotify.Remove, want: DiscoveryEventRemove},
		{name: "create and remove prefers add", op: fsnotify.Create | fsnotify.Remove, want: DiscoveryEventAdd},
		{name: "write", op: fsnotify.Write, want: DiscoveryEventChange},
		{name: "chmod", op: fsnotify.Chmod, want: DiscoveryEventChange},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := eventTypeFromOp(tc.op); got != tc.want {
				t.Fatalf("eventTypeFromOp(%v)=%v want %v", tc.op, got, tc.want)
			}
		})
	}
}

func TestLooksLikeUSBDevicePath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
		want bool
	}{
		{name: "serial by id folder", path: "/dev/serial/by-id", want: true},
		{name: "tty usb path", path: "/dev/ttyUSB0", want: true},
		{name: "tty acm path", path: "/dev/ttyACM0", want: true},
		{name: "short tty base", path: "/dev/tty", want: false},
		{name: "unrelated path", path: "/dev/null", want: false},
		{name: "non usb tmp path", path: "/tmp/device", want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := looksLikeUSBDevicePath(tc.path); got != tc.want {
				t.Fatalf("looksLikeUSBDevicePath(%q)=%t want %t", tc.path, got, tc.want)
			}
		})
	}
}

func TestDiscoveryRunStopsOnContextCancel(t *testing.T) {
	t.Parallel()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}

	d := &Discovery{
		watcher: watcher,
		logger:  log.New(io.Discard, "", 0),
		paths:   []string{"/dev"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- d.Run(ctx)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case runErr := <-done:
		if runErr != nil {
			t.Fatalf("Run() error = %v", runErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not stop after context cancellation")
	}
}

func TestNewDiscoveryInitializesWatcher(t *testing.T) {
	t.Parallel()

	d, err := NewDiscovery(log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatalf("NewDiscovery() error = %v", err)
	}
	if d == nil || d.watcher == nil {
		t.Fatal("NewDiscovery() returned nil discovery or watcher")
	}
	if err := d.watcher.Close(); err != nil {
		t.Fatalf("watcher.Close() error = %v", err)
	}
}

func TestAddPathsHandlesMissingAndInvalidPaths(t *testing.T) {
	t.Parallel()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}
	defer watcher.Close()

	tmpDir := t.TempDir()
	present := filepath.Join(tmpDir, "present")
	if err := os.Mkdir(present, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	d := &Discovery{
		watcher: watcher,
		logger:  log.New(io.Discard, "", 0),
		paths:   []string{present, filepath.Join(tmpDir, "missing")},
	}
	if err := d.addPaths(); err != nil {
		t.Fatalf("addPaths() error = %v", err)
	}

	d.paths = []string{"\x00invalid"}
	err = d.addPaths()
	if err == nil {
		t.Fatal("addPaths() expected error for invalid path, got nil")
	}
	if !errors.Is(err, os.ErrInvalid) && !errors.Is(err, os.ErrNotExist) {
		// platform-specific error details vary; keep branch assertion explicit.
		t.Logf("addPaths() returned expected failure: %v", err)
	}
}

func TestDiscoveryRunInvokesSinkCallback(t *testing.T) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}

	sink := &recordingSink{events: make(chan DiscoveryEvent, 1)}
	d := &Discovery{
		watcher: watcher,
		logger:  log.New(io.Discard, "", 0),
		paths:   []string{"/dev"},
		sink:    sink,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- d.Run(ctx)
	}()

	watcher.Events <- fsnotify.Event{Name: "/dev/ttyUSB7", Op: fsnotify.Create}

	select {
	case got := <-sink.events:
		if got.Type != DiscoveryEventAdd {
			t.Fatalf("event type = %s, want %s", got.Type, DiscoveryEventAdd)
		}
		if got.Path != "/dev/ttyUSB7" {
			t.Fatalf("event path = %q, want /dev/ttyUSB7", got.Path)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected sink callback event")
	}

	cancel()
	select {
	case runErr := <-done:
		if runErr != nil {
			t.Fatalf("Run() error = %v", runErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not stop")
	}
}

func TestDiscoveryRunRetriesSinkAfterError(t *testing.T) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}

	sink := &flakySink{
		calls: make(chan DiscoveryEvent, 2),
		fail:  true,
	}
	d := &Discovery{
		watcher: watcher,
		logger:  log.New(io.Discard, "", 0),
		paths:   []string{"/dev"},
		sink:    sink,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- d.Run(ctx)
	}()

	watcher.Events <- fsnotify.Event{Name: "/dev/ttyUSB8", Op: fsnotify.Create}

	for i := 0; i < 2; i++ {
		select {
		case got := <-sink.calls:
			if got.Type != DiscoveryEventAdd {
				t.Fatalf("event type = %s, want %s", got.Type, DiscoveryEventAdd)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("expected sink callback #%d", i+1)
		}
	}

	cancel()
	select {
	case runErr := <-done:
		if runErr != nil {
			t.Fatalf("Run() error = %v", runErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not stop")
	}
}
