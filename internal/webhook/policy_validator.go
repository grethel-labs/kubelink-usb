package webhook

import (
	"context"
	"fmt"
	"regexp"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

var hexIDRegex = regexp.MustCompile(`^[0-9a-fA-F]{4}$`)

// PolicyValidator validates USBDevicePolicy selector IDs.
type PolicyValidator struct{}

// NewPolicyValidator creates a policy validator.
func NewPolicyValidator() admission.CustomValidator {
	return &PolicyValidator{}
}

func (v *PolicyValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, validatePolicyObject(obj)
}

func (v *PolicyValidator) ValidateUpdate(_ context.Context, _, newObj runtime.Object) (admission.Warnings, error) {
	return nil, validatePolicyObject(newObj)
}

func (v *PolicyValidator) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validatePolicyObject(obj runtime.Object) error {
	policy, ok := obj.(*usbv1alpha1.USBDevicePolicy)
	if !ok {
		return fmt.Errorf("expected USBDevicePolicy, got %T", obj)
	}

	var allErrs field.ErrorList
	if policy.Spec.Selector.VendorID != "" && !hexIDRegex.MatchString(policy.Spec.Selector.VendorID) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "selector", "vendorID"), policy.Spec.Selector.VendorID, "must be 4-digit hex value"))
	}
	if policy.Spec.Selector.ProductID != "" && !hexIDRegex.MatchString(policy.Spec.Selector.ProductID) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "selector", "productID"), policy.Spec.Selector.ProductID, "must be 4-digit hex value"))
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: usbv1alpha1.GroupVersion.Group, Kind: "USBDevicePolicy"},
		policy.Name,
		allErrs,
	)
}
