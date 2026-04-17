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
		{name: "write", op: fsnotify.Write, want: DiscoveryEventChange},
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
