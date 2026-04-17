package agent

import (
	"testing"

	"github.com/fsnotify/fsnotify"
)

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
		{name: "short tty base", path: "/dev/tty", want: true},
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
