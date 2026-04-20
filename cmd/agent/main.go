// Package main is the entrypoint for the kubelink-usb node agent.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
	"github.com/grethel-labs/kubelink-usb/internal/agent"
	"github.com/grethel-labs/kubelink-usb/internal/metrics"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	metrics.Register()
	logger := log.Default()
	logger.Printf("starting agent version=%s", version)

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(usbv1alpha1.AddToScheme(scheme))

	cfg := ctrl.GetConfigOrDie()
	kubeClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		log.Fatalf("failed to initialize kubernetes client: %v", err)
	}

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName, err = os.Hostname()
		if err != nil {
			log.Fatalf("failed to determine node name: %v", err)
		}
	}

	bridge := agent.NewUSBDeviceBridge(kubeClient, nodeName, logger)
	discovery, err := agent.NewDiscoveryWithSink(logger, bridge)
	if err != nil {
		log.Fatalf("failed to initialize discovery watcher: %v", err)
	}

	if err := discovery.Run(ctx); err != nil {
		log.Fatalf("discovery watcher failed: %v", err)
	}
}
