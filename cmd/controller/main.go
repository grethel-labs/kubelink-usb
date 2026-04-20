// Package main is the entrypoint for the kubelink-usb controller manager.
package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
	"github.com/grethel-labs/kubelink-usb/internal/backup"
	"github.com/grethel-labs/kubelink-usb/internal/controller"
	kmetrics "github.com/grethel-labs/kubelink-usb/internal/metrics"
	usbwebhook "github.com/grethel-labs/kubelink-usb/internal/webhook"
)

var (
	scheme  = runtime.NewScheme()
	version = "dev"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(usbv1alpha1.AddToScheme(scheme))
}

func main() {
	var metricsAddr, probeAddr string
	var enableLeaderElection bool

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")
	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	setupLog := ctrl.Log.WithName("setup")
	setupLog.Info("starting controller", "version", version)
	kmetrics.Register()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "kubelink-usb.kubelink-usb.io",
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
	})
	if err != nil {
		setupLog.Error(err, "failed to create manager")
		os.Exit(1)
	}

	if err := (&controller.USBDeviceReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("usbdevice-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "failed to create USBDevice controller")
		os.Exit(1)
	}

	if err := (&controller.ApprovalReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("approval-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "failed to create Approval controller")
		os.Exit(1)
	}

	if err := (&controller.USBConnectionReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("usbconnection-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "failed to create USBConnection controller")
		os.Exit(1)
	}

	if err := ctrl.NewWebhookManagedBy(mgr).
		For(&usbv1alpha1.USBDevicePolicy{}).
		WithValidator(usbwebhook.NewPolicyValidator()).
		Complete(); err != nil {
		setupLog.Error(err, "failed to register USBDevicePolicy webhook")
		os.Exit(1)
	}
	if err := ctrl.NewWebhookManagedBy(mgr).
		For(&usbv1alpha1.USBDevice{}).
		WithDefaulter(usbwebhook.NewDeviceDefaulter()).
		Complete(); err != nil {
		setupLog.Error(err, "failed to register USBDevice webhook")
		os.Exit(1)
	}

	// Initialize backup storage from default configmap-based destination.
	// In production, this would be loaded from a USBBackupConfig CR.
	backupStorage := backup.NewConfigMapStorage("kubelink-backup-store")

	if err := (&controller.BackupReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Storage: backupStorage}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "failed to create Backup controller")
		os.Exit(1)
	}
	if err := (&controller.RestoreReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Storage: backupStorage}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "failed to create Restore controller")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "failed to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "failed to set up ready check")
		os.Exit(1)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "manager exited non-zero")
		os.Exit(1)
	}
}
