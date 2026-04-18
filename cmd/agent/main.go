package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/grethel-labs/kubelink-usb/internal/agent"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	discovery, err := agent.NewDiscovery(log.Default())
	if err != nil {
		log.Fatalf("failed to initialize discovery watcher: %v", err)
	}

	if err := discovery.Run(ctx); err != nil {
		log.Fatalf("discovery watcher failed: %v", err)
	}
}
