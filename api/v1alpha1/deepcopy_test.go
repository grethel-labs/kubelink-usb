package v1alpha1

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestDeepCopyObjectAndNestedFields(t *testing.T) {
	t.Parallel()

	now := metav1.NewTime(time.Unix(1710000000, 0))
	policy := &USBDevicePolicy{
		Spec: USBDevicePolicySpec{
			Selector: USBDevicePolicySelector{
				VendorID:  "1d6b",
				ProductID: "0002",
				NodeNames: []string{"node-a"},
			},
			Restrictions: USBDeviceRestrictions{
				AllowedNodes:         []string{"node-a"},
				AllowedNamespaces:    []string{"default"},
				AllowedDeviceClasses: []string{"serial"},
			},
		},
	}
	connection := &USBConnection{
		Spec: USBConnectionSpec{
			DeviceRef: USBConnectionDeviceRef{Name: "dev-1"},
			PodSelector: &USBConnectionPodSelector{
				MatchLabels: map[string]string{"app": "consumer"},
			},
		},
		Status: USBConnectionStatus{
			TunnelInfo: &USBConnectionTunnelInfo{ServerHost: "10.0.0.2", ServerPort: 3240},
			Conditions: []metav1.Condition{{Type: "Ready", Status: metav1.ConditionTrue}},
		},
	}
	approval := &USBDeviceApproval{
		Spec: USBDeviceApprovalSpec{
			DeviceRef:  USBDeviceApprovalDeviceRef{Name: "dev-1"},
			PolicyRef:  &USBDeviceApprovalPolicyRef{Name: "default"},
			ApprovedAt: &now,
		},
		Status: USBDeviceApprovalStatus{
			ApprovedAt: &now,
		},
	}
	device := &USBDevice{
		Spec: USBDeviceSpec{
			BusID:      "1-1",
			DevicePath: "/dev/ttyUSB0",
		},
		Status: USBDeviceStatus{
			ConnectionInfo: &USBDeviceConnectionInfo{Host: "10.0.0.3", Port: 3240},
			LastSeen:       &now,
		},
	}

	policyCopy := policy.DeepCopy()
	connectionCopy := connection.DeepCopy()
	approvalCopy := approval.DeepCopy()
	deviceCopy := device.DeepCopy()

	policyCopy.Spec.Selector.NodeNames[0] = "node-b"
	policyCopy.Spec.Restrictions.AllowedNamespaces[0] = "kube-system"
	connectionCopy.Spec.PodSelector.MatchLabels["app"] = "other"
	connectionCopy.Status.Conditions[0].Type = "Changed"
	approvalCopy.Spec.PolicyRef.Name = "updated"
	approvalCopy.Status.ApprovedBy = "admin"
	deviceCopy.Status.ConnectionInfo.Host = "127.0.0.1"

	if got := policy.Spec.Selector.NodeNames[0]; got != "node-a" {
		t.Fatalf("policy selector mutated: got %q", got)
	}
	if got := policy.Spec.Restrictions.AllowedNamespaces[0]; got != "default" {
		t.Fatalf("policy restrictions mutated: got %q", got)
	}
	if got := connection.Spec.PodSelector.MatchLabels["app"]; got != "consumer" {
		t.Fatalf("connection selector mutated: got %q", got)
	}
	if got := connection.Status.Conditions[0].Type; got != "Ready" {
		t.Fatalf("connection condition mutated: got %q", got)
	}
	if got := approval.Spec.PolicyRef.Name; got != "default" {
		t.Fatalf("approval policyRef mutated: got %q", got)
	}
	if got := device.Status.ConnectionInfo.Host; got != "10.0.0.3" {
		t.Fatalf("device connectionInfo mutated: got %q", got)
	}

	var object runtime.Object = device.DeepCopyObject()
	if _, ok := object.(*USBDevice); !ok {
		t.Fatalf("DeepCopyObject returned %T, expected *USBDevice", object)
	}
	if _, ok := connection.DeepCopyObject().(*USBConnection); !ok {
		t.Fatalf("USBConnection.DeepCopyObject returned unexpected type")
	}
	if _, ok := approval.DeepCopyObject().(*USBDeviceApproval); !ok {
		t.Fatalf("USBDeviceApproval.DeepCopyObject returned unexpected type")
	}
	if _, ok := policy.DeepCopyObject().(*USBDevicePolicy); !ok {
		t.Fatalf("USBDevicePolicy.DeepCopyObject returned unexpected type")
	}
}

func TestDeepCopyNilReceivers(t *testing.T) {
	t.Parallel()

	var (
		connection            *USBConnection
		connectionList        *USBConnectionList
		connectionSpec        *USBConnectionSpec
		connectionStatus      *USBConnectionStatus
		policy                *USBDevicePolicy
		policyList            *USBDevicePolicyList
		policySelector        *USBDevicePolicySelector
		policySpec            *USBDevicePolicySpec
		policyStatus          *USBDevicePolicyStatus
		restrictions          *USBDeviceRestrictions
		lifecycle             *USBDeviceLifecycle
		approval              *USBDeviceApproval
		approvalList          *USBDeviceApprovalList
		approvalSpec          *USBDeviceApprovalSpec
		approvalStatus        *USBDeviceApprovalStatus
		approvalConfig        *USBDeviceApprovalConfig
		approvalDeviceRef     *USBDeviceApprovalDeviceRef
		approvalPolicyRef     *USBDeviceApprovalPolicyRef
		device                *USBDevice
		deviceList            *USBDeviceList
		deviceSpec            *USBDeviceSpec
		deviceStatus          *USBDeviceStatus
		deviceConnectionInfo  *USBDeviceConnectionInfo
		connectionPodSelector *USBConnectionPodSelector
		connectionDeviceRef   *USBConnectionDeviceRef
		connectionTunnelInfo  *USBConnectionTunnelInfo

		// New backup/restore/whitelist types
		whitelist          *USBDeviceWhitelist
		whitelistList      *USBDeviceWhitelistList
		whitelistSpec      *USBDeviceWhitelistSpec
		whitelistStatus    *USBDeviceWhitelistStatus
		whitelistEntry     *WhitelistEntry
		backupConfig       *USBBackupConfig
		backupConfigList   *USBBackupConfigList
		backupConfigSpec   *USBBackupConfigSpec
		backupConfigStatus *USBBackupConfigStatus
		backupDest         *BackupDestination
		backupDestPVC      *BackupDestinationPVC
		backupDestS3       *BackupDestinationS3
		backupDestSecret   *BackupDestinationSecret
		backupDestCM       *BackupDestinationConfigMap
		autoRestore        *AutoRestoreConfig
		usbBackup          *USBBackup
		backupList         *USBBackupList
		backupSpec         *USBBackupSpec
		backupStatus       *USBBackupStatus
		backupItemCounts   *BackupItemCounts
		restore            *USBRestore
		restoreList        *USBRestoreList
		restoreSpec        *USBRestoreSpec
		restoreStatus      *USBRestoreStatus
		restoreBackupRef   *RestoreBackupRef
		preRestoreHealth   *PreRestoreHealthCheck
		connRevalidation   *ConnectionRevalidation
		restoreItemCounts  *RestoreItemCounts
	)

	if connection.DeepCopy() != nil ||
		connectionList.DeepCopy() != nil ||
		connectionSpec.DeepCopy() != nil ||
		connectionStatus.DeepCopy() != nil ||
		policy.DeepCopy() != nil ||
		policyList.DeepCopy() != nil ||
		policySelector.DeepCopy() != nil ||
		policySpec.DeepCopy() != nil ||
		policyStatus.DeepCopy() != nil ||
		restrictions.DeepCopy() != nil ||
		lifecycle.DeepCopy() != nil ||
		approval.DeepCopy() != nil ||
		approvalList.DeepCopy() != nil ||
		approvalSpec.DeepCopy() != nil ||
		approvalStatus.DeepCopy() != nil ||
		approvalConfig.DeepCopy() != nil ||
		approvalDeviceRef.DeepCopy() != nil ||
		approvalPolicyRef.DeepCopy() != nil ||
		device.DeepCopy() != nil ||
		deviceList.DeepCopy() != nil ||
		deviceSpec.DeepCopy() != nil ||
		deviceStatus.DeepCopy() != nil ||
		deviceConnectionInfo.DeepCopy() != nil ||
		connectionPodSelector.DeepCopy() != nil ||
		connectionDeviceRef.DeepCopy() != nil ||
		connectionTunnelInfo.DeepCopy() != nil ||
		whitelist.DeepCopy() != nil ||
		whitelistList.DeepCopy() != nil ||
		whitelistSpec.DeepCopy() != nil ||
		whitelistStatus.DeepCopy() != nil ||
		whitelistEntry.DeepCopy() != nil ||
		backupConfig.DeepCopy() != nil ||
		backupConfigList.DeepCopy() != nil ||
		backupConfigSpec.DeepCopy() != nil ||
		backupConfigStatus.DeepCopy() != nil ||
		backupDest.DeepCopy() != nil ||
		backupDestPVC.DeepCopy() != nil ||
		backupDestS3.DeepCopy() != nil ||
		backupDestSecret.DeepCopy() != nil ||
		backupDestCM.DeepCopy() != nil ||
		autoRestore.DeepCopy() != nil ||
		usbBackup.DeepCopy() != nil ||
		backupList.DeepCopy() != nil ||
		backupSpec.DeepCopy() != nil ||
		backupStatus.DeepCopy() != nil ||
		backupItemCounts.DeepCopy() != nil ||
		restore.DeepCopy() != nil ||
		restoreList.DeepCopy() != nil ||
		restoreSpec.DeepCopy() != nil ||
		restoreStatus.DeepCopy() != nil ||
		restoreBackupRef.DeepCopy() != nil ||
		preRestoreHealth.DeepCopy() != nil ||
		connRevalidation.DeepCopy() != nil ||
		restoreItemCounts.DeepCopy() != nil {
		t.Fatal("nil DeepCopy receiver should return nil")
	}
}

func TestAddToSchemeRegistersKinds(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme() error = %v", err)
	}

	for _, obj := range []runtime.Object{
		&USBDevice{},
		&USBConnection{},
		&USBDevicePolicy{},
		&USBDeviceApproval{},
		&USBDeviceWhitelist{},
		&USBBackupConfig{},
		&USBBackup{},
		&USBRestore{},
	} {
		kinds, _, err := scheme.ObjectKinds(obj)
		if err != nil {
			t.Fatalf("ObjectKinds(%T) error = %v", obj, err)
		}
		if len(kinds) == 0 {
			t.Fatalf("ObjectKinds(%T) returned no kinds", obj)
		}
	}
}

func TestDeepCopyGeneratedCoverageForReferenceAndListTypes(t *testing.T) {
	t.Parallel()

	now := metav1.NewTime(time.Unix(1710000300, 0))

	connectionList := &USBConnectionList{
		Items: []USBConnection{
			{
				Spec: USBConnectionSpec{
					DeviceRef: USBConnectionDeviceRef{Name: "dev-a"},
					PodSelector: &USBConnectionPodSelector{
						MatchLabels: map[string]string{"app": "x"},
					},
				},
			},
		},
	}
	approvalList := &USBDeviceApprovalList{
		Items: []USBDeviceApproval{
			{
				Spec: USBDeviceApprovalSpec{
					DeviceRef: USBDeviceApprovalDeviceRef{Name: "dev-a"},
					PolicyRef: &USBDeviceApprovalPolicyRef{Name: "policy-a"},
				},
			},
		},
	}
	deviceList := &USBDeviceList{
		Items: []USBDevice{
			{
				Spec: USBDeviceSpec{
					BusID:    "1-1",
					NodeName: "node-a",
				},
				Status: USBDeviceStatus{
					ConnectionInfo: &USBDeviceConnectionInfo{Host: "10.0.0.1"},
					LastSeen:       &now,
				},
			},
		},
	}
	policyList := &USBDevicePolicyList{
		Items: []USBDevicePolicy{
			{
				Spec: USBDevicePolicySpec{
					Selector: USBDevicePolicySelector{NodeNames: []string{"node-a"}},
					Restrictions: USBDeviceRestrictions{
						AllowedNodes: []string{"node-a"},
					},
					Lifecycle: USBDeviceLifecycle{ReconnectAttempts: 3},
				},
				Status: USBDevicePolicyStatus{ObservedGeneration: 1},
			},
		},
	}

	if connectionList.DeepCopyObject() == nil ||
		approvalList.DeepCopyObject() == nil ||
		deviceList.DeepCopyObject() == nil ||
		policyList.DeepCopyObject() == nil {
		t.Fatal("DeepCopyObject() expected non-nil for list objects")
	}

	connectionListCopy := connectionList.DeepCopy()
	connectionListCopy.Items[0].Spec.PodSelector.MatchLabels["app"] = "changed"
	if got := connectionList.Items[0].Spec.PodSelector.MatchLabels["app"]; got != "x" {
		t.Fatalf("connection list original map mutated: got %q", got)
	}

	approvalListCopy := approvalList.DeepCopy()
	approvalListCopy.Items[0].Spec.PolicyRef.Name = "changed"
	if got := approvalList.Items[0].Spec.PolicyRef.Name; got != "policy-a" {
		t.Fatalf("approval list original policyRef mutated: got %q", got)
	}

	deviceListCopy := deviceList.DeepCopy()
	deviceListCopy.Items[0].Status.ConnectionInfo.Host = "127.0.0.1"
	if got := deviceList.Items[0].Status.ConnectionInfo.Host; got != "10.0.0.1" {
		t.Fatalf("device list original connectionInfo mutated: got %q", got)
	}

	policyListCopy := policyList.DeepCopy()
	policyListCopy.Items[0].Spec.Selector.NodeNames[0] = "node-b"
	if got := policyList.Items[0].Spec.Selector.NodeNames[0]; got != "node-a" {
		t.Fatalf("policy list original selector mutated: got %q", got)
	}

	var (
		connRefIn         = USBConnectionDeviceRef{Name: "dev-a", Namespace: "ns-a"}
		connRefOut        USBConnectionDeviceRef
		tunnelIn          = USBConnectionTunnelInfo{ServerHost: "h", ServerPort: 1, Protocol: "tcp"}
		tunnelOut         USBConnectionTunnelInfo
		approvalRefIn     = USBDeviceApprovalDeviceRef{Name: "dev-a"}
		approvalRefOut    USBDeviceApprovalDeviceRef
		policyRefIn       = USBDeviceApprovalPolicyRef{Name: "policy-a", Namespace: "ns-a"}
		policyRefOut      USBDeviceApprovalPolicyRef
		connectionInfoIn  = USBDeviceConnectionInfo{Host: "h", Port: 1, ExportedBusID: "1-1"}
		connectionInfoOut USBDeviceConnectionInfo
		lifecycleIn       = USBDeviceLifecycle{ReconnectAttempts: 2}
		lifecycleOut      USBDeviceLifecycle
		specIn            = USBDeviceSpec{BusID: "1-1", DevicePath: "/dev/ttyUSB0"}
		specOut           USBDeviceSpec
		statusIn          = USBDevicePolicyStatus{ObservedGeneration: 7}
		statusOut         USBDevicePolicyStatus
	)

	connRefIn.DeepCopyInto(&connRefOut)
	tunnelIn.DeepCopyInto(&tunnelOut)
	approvalRefIn.DeepCopyInto(&approvalRefOut)
	policyRefIn.DeepCopyInto(&policyRefOut)
	connectionInfoIn.DeepCopyInto(&connectionInfoOut)
	lifecycleIn.DeepCopyInto(&lifecycleOut)
	specIn.DeepCopyInto(&specOut)
	statusIn.DeepCopyInto(&statusOut)

	if connRefOut.Name != connRefIn.Name ||
		tunnelOut.ServerHost != tunnelIn.ServerHost ||
		approvalRefOut.Name != approvalRefIn.Name ||
		policyRefOut.Name != policyRefIn.Name ||
		connectionInfoOut.Host != connectionInfoIn.Host ||
		lifecycleOut.ReconnectAttempts != lifecycleIn.ReconnectAttempts ||
		specOut.DevicePath != specIn.DevicePath ||
		statusOut.ObservedGeneration != statusIn.ObservedGeneration {
		t.Fatal("DeepCopyInto() did not copy expected values")
	}
}

func TestDeepCopyNonNilReceiversForAllHelperTypes(t *testing.T) {
	t.Parallel()

	now := metav1.NewTime(time.Unix(1710000600, 0))

	connRef := (&USBConnectionDeviceRef{Name: "dev-a"}).DeepCopy()
	if connRef == nil || connRef.Name != "dev-a" {
		t.Fatal("USBConnectionDeviceRef.DeepCopy() returned unexpected value")
	}

	podSelector := (&USBConnectionPodSelector{MatchLabels: map[string]string{"app": "demo"}}).DeepCopy()
	if podSelector == nil || podSelector.MatchLabels["app"] != "demo" {
		t.Fatal("USBConnectionPodSelector.DeepCopy() returned unexpected value")
	}

	connSpec := (&USBConnectionSpec{
		DeviceRef:   USBConnectionDeviceRef{Name: "dev-a"},
		PodSelector: &USBConnectionPodSelector{MatchLabels: map[string]string{"app": "demo"}},
	}).DeepCopy()
	if connSpec == nil || connSpec.PodSelector == nil {
		t.Fatal("USBConnectionSpec.DeepCopy() returned unexpected value")
	}

	connStatus := (&USBConnectionStatus{
		LastRetryTime: &metav1.Time{Time: time.Now()},
		TunnelInfo:    &USBConnectionTunnelInfo{ServerHost: "10.0.0.1"},
		Conditions:    []metav1.Condition{{Type: "Ready", Status: metav1.ConditionTrue}},
	}).DeepCopy()
	if connStatus == nil || connStatus.TunnelInfo == nil || connStatus.LastRetryTime == nil {
		t.Fatal("USBConnectionStatus.DeepCopy() returned unexpected value")
	}

	tunnel := (&USBConnectionTunnelInfo{ServerHost: "10.0.0.1"}).DeepCopy()
	if tunnel == nil || tunnel.ServerHost != "10.0.0.1" {
		t.Fatal("USBConnectionTunnelInfo.DeepCopy() returned unexpected value")
	}

	approvalConfig := (&USBDeviceApprovalConfig{Mode: "manual"}).DeepCopy()
	if approvalConfig == nil || approvalConfig.Mode != "manual" {
		t.Fatal("USBDeviceApprovalConfig.DeepCopy() returned unexpected value")
	}

	approvalDeviceRef := (&USBDeviceApprovalDeviceRef{Name: "dev-a"}).DeepCopy()
	if approvalDeviceRef == nil || approvalDeviceRef.Name != "dev-a" {
		t.Fatal("USBDeviceApprovalDeviceRef.DeepCopy() returned unexpected value")
	}

	approvalPolicyRef := (&USBDeviceApprovalPolicyRef{Name: "policy-a"}).DeepCopy()
	if approvalPolicyRef == nil || approvalPolicyRef.Name != "policy-a" {
		t.Fatal("USBDeviceApprovalPolicyRef.DeepCopy() returned unexpected value")
	}

	approvalSpec := (&USBDeviceApprovalSpec{
		DeviceRef:  USBDeviceApprovalDeviceRef{Name: "dev-a"},
		PolicyRef:  &USBDeviceApprovalPolicyRef{Name: "policy-a"},
		ApprovedAt: &now,
		ExpiresAt:  &now,
	}).DeepCopy()
	if approvalSpec == nil || approvalSpec.PolicyRef == nil || approvalSpec.ApprovedAt == nil || approvalSpec.ExpiresAt == nil {
		t.Fatal("USBDeviceApprovalSpec.DeepCopy() returned unexpected value")
	}

	approvalSpecWithoutOptional := (&USBDeviceApprovalSpec{
		DeviceRef: USBDeviceApprovalDeviceRef{Name: "dev-a"},
	}).DeepCopy()
	if approvalSpecWithoutOptional == nil {
		t.Fatal("USBDeviceApprovalSpec.DeepCopy() expected non-nil")
	}

	approvalStatus := (&USBDeviceApprovalStatus{
		ApprovedAt: &now,
		ExpiresAt:  &now,
	}).DeepCopy()
	if approvalStatus == nil || approvalStatus.ApprovedAt == nil || approvalStatus.ExpiresAt == nil {
		t.Fatal("USBDeviceApprovalStatus.DeepCopy() returned unexpected value")
	}

	approvalStatusWithoutOptional := (&USBDeviceApprovalStatus{}).DeepCopy()
	if approvalStatusWithoutOptional == nil {
		t.Fatal("USBDeviceApprovalStatus.DeepCopy() expected non-nil")
	}

	connectionInfo := (&USBDeviceConnectionInfo{Host: "10.0.0.1"}).DeepCopy()
	if connectionInfo == nil || connectionInfo.Host != "10.0.0.1" {
		t.Fatal("USBDeviceConnectionInfo.DeepCopy() returned unexpected value")
	}

	lifecycle := (&USBDeviceLifecycle{ReconnectAttempts: 3}).DeepCopy()
	if lifecycle == nil || lifecycle.ReconnectAttempts != 3 {
		t.Fatal("USBDeviceLifecycle.DeepCopy() returned unexpected value")
	}

	policySelector := (&USBDevicePolicySelector{NodeNames: []string{"node-a"}}).DeepCopy()
	if policySelector == nil || len(policySelector.NodeNames) != 1 {
		t.Fatal("USBDevicePolicySelector.DeepCopy() returned unexpected value")
	}

	policySpec := (&USBDevicePolicySpec{
		Selector: USBDevicePolicySelector{NodeNames: []string{"node-a"}},
		Approval: USBDeviceApprovalConfig{Mode: "manual"},
		Restrictions: USBDeviceRestrictions{
			AllowedNodes: []string{"node-a"},
		},
	}).DeepCopy()
	if policySpec == nil || len(policySpec.Selector.NodeNames) == 0 {
		t.Fatal("USBDevicePolicySpec.DeepCopy() returned unexpected value")
	}

	policyStatus := (&USBDevicePolicyStatus{ObservedGeneration: 5}).DeepCopy()
	if policyStatus == nil || policyStatus.ObservedGeneration != 5 {
		t.Fatal("USBDevicePolicyStatus.DeepCopy() returned unexpected value")
	}

	restrictions := (&USBDeviceRestrictions{
		AllowedNodes:         []string{"node-a"},
		AllowedNamespaces:    []string{"default"},
		AllowedDeviceClasses: []string{"serial"},
	}).DeepCopy()
	if restrictions == nil || len(restrictions.AllowedNodes) != 1 {
		t.Fatal("USBDeviceRestrictions.DeepCopy() returned unexpected value")
	}

	spec := (&USBDeviceSpec{BusID: "1-1", DevicePath: "/dev/ttyUSB0"}).DeepCopy()
	if spec == nil || spec.DevicePath != "/dev/ttyUSB0" {
		t.Fatal("USBDeviceSpec.DeepCopy() returned unexpected value")
	}

	deviceStatus := (&USBDeviceStatus{
		ConnectionInfo: &USBDeviceConnectionInfo{Host: "10.0.0.1"},
		LastSeen:       &now,
	}).DeepCopy()
	if deviceStatus == nil || deviceStatus.ConnectionInfo == nil || deviceStatus.LastSeen == nil {
		t.Fatal("USBDeviceStatus.DeepCopy() returned unexpected value")
	}
}

func TestDeepCopyBackupRestoreWhitelistTypes(t *testing.T) {
	t.Parallel()

	now := metav1.NewTime(time.Unix(1710001000, 0))

	// --- USBDeviceWhitelist ---
	wl := &USBDeviceWhitelist{
		ObjectMeta: metav1.ObjectMeta{Name: "wl-1"},
		Spec: USBDeviceWhitelistSpec{
			Entries: []WhitelistEntry{
				{Fingerprint: "fp-1", AddedAt: &now, AddedBy: "system", PolicyRef: "default/policy-1"},
			},
		},
		Status: USBDeviceWhitelistStatus{EntryCount: 1, LastUpdated: &now, ObservedGeneration: 2},
	}
	wlCopy := wl.DeepCopy()
	wlCopy.Spec.Entries[0].Fingerprint = "changed"
	if wl.Spec.Entries[0].Fingerprint != "fp-1" {
		t.Fatal("whitelist entry mutated")
	}
	if wl.DeepCopyObject() == nil {
		t.Fatal("DeepCopyObject() expected non-nil")
	}

	wlList := &USBDeviceWhitelistList{Items: []USBDeviceWhitelist{*wl}}
	wlListCopy := wlList.DeepCopy()
	wlListCopy.Items[0].Spec.Entries[0].Fingerprint = "changed"
	if wlList.Items[0].Spec.Entries[0].Fingerprint != "fp-1" {
		t.Fatal("whitelist list entry mutated")
	}
	if wlList.DeepCopyObject() == nil {
		t.Fatal("DeepCopyObject() expected non-nil")
	}

	// --- WhitelistEntry ---
	entry := (&WhitelistEntry{Fingerprint: "fp-a", AddedAt: &now}).DeepCopy()
	if entry == nil || entry.Fingerprint != "fp-a" || entry.AddedAt == nil {
		t.Fatal("WhitelistEntry.DeepCopy() returned unexpected value")
	}

	// --- USBBackupConfig ---
	bc := &USBBackupConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "config"},
		Spec: USBBackupConfigSpec{
			Schedule:       "0 */6 * * *",
			RetentionCount: 10,
			Destination: BackupDestination{
				Type:      "s3",
				PVC:       &BackupDestinationPVC{ClaimName: "test", SubPath: "backups/"},
				S3:        &BackupDestinationS3{Bucket: "b", Endpoint: "e", Region: "r", SecretRef: &BackupDestinationSecret{Name: "s"}},
				ConfigMap: &BackupDestinationConfigMap{Name: "cm"},
			},
			AutoRestore: AutoRestoreConfig{Enabled: true},
		},
		Status: USBBackupConfigStatus{LastBackupTime: &now, LastBackupName: "bk-1", BackupCount: 5, HealthStatus: "Healthy"},
	}
	bcCopy := bc.DeepCopy()
	bcCopy.Spec.Destination.S3.Bucket = "changed"
	if bc.Spec.Destination.S3.Bucket != "b" {
		t.Fatal("backup config S3 bucket mutated")
	}
	bcCopy.Spec.Destination.S3.SecretRef.Name = "changed"
	if bc.Spec.Destination.S3.SecretRef.Name != "s" {
		t.Fatal("backup config S3 secret mutated")
	}
	bcCopy.Spec.Destination.PVC.ClaimName = "changed"
	if bc.Spec.Destination.PVC.ClaimName != "test" {
		t.Fatal("backup config PVC mutated")
	}
	bcCopy.Spec.Destination.ConfigMap.Name = "changed"
	if bc.Spec.Destination.ConfigMap.Name != "cm" {
		t.Fatal("backup config ConfigMap mutated")
	}
	if bc.DeepCopyObject() == nil {
		t.Fatal("DeepCopyObject() expected non-nil")
	}

	bcList := &USBBackupConfigList{Items: []USBBackupConfig{*bc}}
	if bcList.DeepCopy() == nil || bcList.DeepCopyObject() == nil {
		t.Fatal("USBBackupConfigList DeepCopy expected non-nil")
	}

	// --- BackupDestination helper types ---
	destPVC := (&BackupDestinationPVC{ClaimName: "c"}).DeepCopy()
	if destPVC == nil || destPVC.ClaimName != "c" {
		t.Fatal("BackupDestinationPVC.DeepCopy() unexpected value")
	}
	destS3 := (&BackupDestinationS3{Bucket: "b", SecretRef: &BackupDestinationSecret{Name: "s"}}).DeepCopy()
	if destS3 == nil || destS3.SecretRef.Name != "s" {
		t.Fatal("BackupDestinationS3.DeepCopy() unexpected value")
	}
	destSecret := (&BackupDestinationSecret{Name: "s"}).DeepCopy()
	if destSecret == nil || destSecret.Name != "s" {
		t.Fatal("BackupDestinationSecret.DeepCopy() unexpected value")
	}
	destCM := (&BackupDestinationConfigMap{Name: "cm"}).DeepCopy()
	if destCM == nil || destCM.Name != "cm" {
		t.Fatal("BackupDestinationConfigMap.DeepCopy() unexpected value")
	}
	dest := (&BackupDestination{Type: "pvc", PVC: &BackupDestinationPVC{ClaimName: "c"}}).DeepCopy()
	if dest == nil || dest.PVC.ClaimName != "c" {
		t.Fatal("BackupDestination.DeepCopy() unexpected value")
	}
	autoRestoreConfig := (&AutoRestoreConfig{Enabled: true}).DeepCopy()
	if autoRestoreConfig == nil || !autoRestoreConfig.Enabled {
		t.Fatal("AutoRestoreConfig.DeepCopy() unexpected value")
	}
	bcSpec := (&USBBackupConfigSpec{Schedule: "* * * * *", RetentionCount: 5, Destination: BackupDestination{Type: "configmap"}}).DeepCopy()
	if bcSpec == nil || bcSpec.Schedule != "* * * * *" {
		t.Fatal("USBBackupConfigSpec.DeepCopy() unexpected value")
	}
	bcStatus := (&USBBackupConfigStatus{LastBackupTime: &now, BackupCount: 3}).DeepCopy()
	if bcStatus == nil || bcStatus.BackupCount != 3 || bcStatus.LastBackupTime == nil {
		t.Fatal("USBBackupConfigStatus.DeepCopy() unexpected value")
	}

	// --- USBBackup ---
	bk := &USBBackup{
		ObjectMeta: metav1.ObjectMeta{Name: "bk-1"},
		Spec:       USBBackupSpec{TriggerType: "manual", TriggeredBy: "admin"},
		Status:     USBBackupStatus{Phase: "Completed", CompletedAt: &now, Checksum: "sha256:abc", ItemCounts: &BackupItemCounts{WhitelistEntries: 5, Policies: 2, Approvals: 3}},
	}
	bkCopy := bk.DeepCopy()
	bkCopy.Status.ItemCounts.WhitelistEntries = 99
	if bk.Status.ItemCounts.WhitelistEntries != 5 {
		t.Fatal("backup item counts mutated")
	}
	if bk.DeepCopyObject() == nil {
		t.Fatal("DeepCopyObject() expected non-nil")
	}

	bkList := &USBBackupList{Items: []USBBackup{*bk}}
	if bkList.DeepCopy() == nil || bkList.DeepCopyObject() == nil {
		t.Fatal("USBBackupList DeepCopy expected non-nil")
	}

	bkSpec := (&USBBackupSpec{TriggerType: "scheduled"}).DeepCopy()
	if bkSpec == nil || bkSpec.TriggerType != "scheduled" {
		t.Fatal("USBBackupSpec.DeepCopy() unexpected value")
	}
	bkStatus := (&USBBackupStatus{Phase: "InProgress", CompletedAt: &now, ItemCounts: &BackupItemCounts{Policies: 1}}).DeepCopy()
	if bkStatus == nil || bkStatus.ItemCounts.Policies != 1 || bkStatus.CompletedAt == nil {
		t.Fatal("USBBackupStatus.DeepCopy() unexpected value")
	}
	itemCounts := (&BackupItemCounts{WhitelistEntries: 10}).DeepCopy()
	if itemCounts == nil || itemCounts.WhitelistEntries != 10 {
		t.Fatal("BackupItemCounts.DeepCopy() unexpected value")
	}

	// --- USBRestore ---
	rs := &USBRestore{
		ObjectMeta: metav1.ObjectMeta{Name: "restore-1"},
		Spec: USBRestoreSpec{
			BackupRef:   RestoreBackupRef{Name: "bk-1"},
			TriggerType: "automatic",
			TriggeredBy: "health-monitor",
			DryRun:      true,
		},
		Status: USBRestoreStatus{
			Phase:                  "Completed",
			PreRestoreHealthCheck:  &PreRestoreHealthCheck{Status: "Healthy", Reason: "ok", CheckedAt: &now},
			RestoredItems:          &RestoreItemCounts{WhitelistEntries: 5, Policies: 2, Approvals: 3},
			ConnectionRevalidation: &ConnectionRevalidation{Total: 10, Valid: 8, Terminated: 2, TerminatedConnections: []string{"conn-a", "conn-b"}},
			CompletedAt:            &now,
		},
	}
	rsCopy := rs.DeepCopy()
	rsCopy.Status.ConnectionRevalidation.TerminatedConnections[0] = "changed"
	if rs.Status.ConnectionRevalidation.TerminatedConnections[0] != "conn-a" {
		t.Fatal("restore terminated connections mutated")
	}
	rsCopy.Status.PreRestoreHealthCheck.Status = "changed"
	if rs.Status.PreRestoreHealthCheck.Status != "Healthy" {
		t.Fatal("restore pre-health-check mutated")
	}
	rsCopy.Status.RestoredItems.Policies = 99
	if rs.Status.RestoredItems.Policies != 2 {
		t.Fatal("restore item counts mutated")
	}
	if rs.DeepCopyObject() == nil {
		t.Fatal("DeepCopyObject() expected non-nil")
	}

	rsList := &USBRestoreList{Items: []USBRestore{*rs}}
	if rsList.DeepCopy() == nil || rsList.DeepCopyObject() == nil {
		t.Fatal("USBRestoreList DeepCopy expected non-nil")
	}

	rsSpec := (&USBRestoreSpec{BackupRef: RestoreBackupRef{Name: "bk-1"}, TriggerType: "manual"}).DeepCopy()
	if rsSpec == nil || rsSpec.BackupRef.Name != "bk-1" {
		t.Fatal("USBRestoreSpec.DeepCopy() unexpected value")
	}
	rsStatus := (&USBRestoreStatus{
		Phase:                  "Restoring",
		PreRestoreHealthCheck:  &PreRestoreHealthCheck{Status: "ok", CheckedAt: &now},
		RestoredItems:          &RestoreItemCounts{Approvals: 5},
		ConnectionRevalidation: &ConnectionRevalidation{Total: 3, TerminatedConnections: []string{"c1"}},
		CompletedAt:            &now,
	}).DeepCopy()
	if rsStatus == nil || rsStatus.PreRestoreHealthCheck == nil || rsStatus.RestoredItems == nil || rsStatus.ConnectionRevalidation == nil || rsStatus.CompletedAt == nil {
		t.Fatal("USBRestoreStatus.DeepCopy() unexpected value")
	}
	backupRef := (&RestoreBackupRef{Name: "bk-a"}).DeepCopy()
	if backupRef == nil || backupRef.Name != "bk-a" {
		t.Fatal("RestoreBackupRef.DeepCopy() unexpected value")
	}
	preHealth := (&PreRestoreHealthCheck{Status: "ok", CheckedAt: &now}).DeepCopy()
	if preHealth == nil || preHealth.CheckedAt == nil {
		t.Fatal("PreRestoreHealthCheck.DeepCopy() unexpected value")
	}
	connReval := (&ConnectionRevalidation{Total: 5, TerminatedConnections: []string{"c1", "c2"}}).DeepCopy()
	if connReval == nil || len(connReval.TerminatedConnections) != 2 {
		t.Fatal("ConnectionRevalidation.DeepCopy() unexpected value")
	}
	rsItemCounts := (&RestoreItemCounts{WhitelistEntries: 3}).DeepCopy()
	if rsItemCounts == nil || rsItemCounts.WhitelistEntries != 3 {
		t.Fatal("RestoreItemCounts.DeepCopy() unexpected value")
	}
}
