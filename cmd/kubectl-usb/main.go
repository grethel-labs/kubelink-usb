// Package main implements the kubectl-usb plugin for listing USB devices.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	ctx := context.Background()
	kubeClient, err := newKubeClient()
	if err != nil {
		fatalf("failed to create kubernetes client: %v", err)
	}

	switch os.Args[1] {
	case "list":
		if err := runList(ctx, kubeClient, os.Args[2:]); err != nil {
			fatalf("list failed: %v", err)
		}
	case "approve":
		if err := runApprovalAction(ctx, kubeClient, os.Args[2:], "Approved"); err != nil {
			fatalf("approve failed: %v", err)
		}
	case "deny":
		if err := runApprovalAction(ctx, kubeClient, os.Args[2:], "Denied"); err != nil {
			fatalf("deny failed: %v", err)
		}
	case "connect":
		if err := runConnectionAction(ctx, kubeClient, os.Args[2:], "Pending"); err != nil {
			fatalf("connect failed: %v", err)
		}
	case "disconnect":
		if err := runConnectionAction(ctx, kubeClient, os.Args[2:], "Disconnected"); err != nil {
			fatalf("disconnect failed: %v", err)
		}
	default:
		usage()
		os.Exit(1)
	}
}

func newKubeClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(usbv1alpha1.AddToScheme(scheme))
	return client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme})
}

func runList(ctx context.Context, kubeClient client.Client, args []string) error {
	cmd := flag.NewFlagSet("list", flag.ContinueOnError)
	_ = cmd.Parse(args)

	var devices usbv1alpha1.USBDeviceList
	if err := kubeClient.List(ctx, &devices); err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tNODE\tVENDOR\tPRODUCT\tPHASE")
	for _, d := range devices.Items {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", d.Name, d.Spec.NodeName, d.Spec.VendorID, d.Spec.ProductID, d.Status.Phase)
	}
	return w.Flush()
}

func runApprovalAction(ctx context.Context, kubeClient client.Client, args []string, phase string) error {
	if len(args) < 1 {
		return fmt.Errorf("approval name is required")
	}
	name := args[0]

	var approval usbv1alpha1.USBDeviceApproval
	if err := kubeClient.Get(ctx, types.NamespacedName{Name: name}, &approval); err != nil {
		return err
	}

	approval.Spec.Phase = phase
	if user := os.Getenv("USER"); user != "" {
		approval.Spec.ApprovedBy = user
	}
	now := metav1.NewTime(time.Now())
	approval.Spec.ApprovedAt = &now

	if err := kubeClient.Update(ctx, &approval); err != nil {
		return err
	}
	fmt.Printf("%s %s\n", strings.ToLower(phase), name)
	return nil
}

func runConnectionAction(ctx context.Context, kubeClient client.Client, args []string, phase string) error {
	if len(args) < 1 {
		return fmt.Errorf("connection must be in namespace/name format")
	}
	nn, err := parseNamespacedName(args[0])
	if err != nil {
		return err
	}

	var conn usbv1alpha1.USBConnection
	if err := kubeClient.Get(ctx, nn, &conn); err != nil {
		return err
	}
	conn.Status.Phase = phase
	if err := kubeClient.Status().Update(ctx, &conn); err != nil {
		return err
	}
	fmt.Printf("%s %s/%s\n", strings.ToLower(phase), nn.Namespace, nn.Name)
	return nil
}

func parseNamespacedName(value string) (types.NamespacedName, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return types.NamespacedName{}, fmt.Errorf("invalid value %q: expected namespace/name", value)
	}
	return types.NamespacedName{Namespace: parts[0], Name: parts[1]}, nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: kubectl-usb <list|approve|deny|connect|disconnect> [args]")
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
