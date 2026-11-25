// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	billingv1alpha1 "go.miloapis.com/billing/api/v1alpha1"
	iamv1alpha1 "go.miloapis.com/milo/pkg/apis/iam/v1alpha1"
	resourcemanagerv1alpha1 "go.miloapis.com/milo/pkg/apis/resourcemanager/v1alpha1"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

// The webhook tests only check the wiring: each webhook calls its rules for
// the right operation and turns the result into the right API error. The
// rules themselves are tested in internal/validation/crm.

// emptyReader finds nothing, so every reference is missing.
func emptyReader(t *testing.T) client.Reader {
	t.Helper()
	s := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{
		crmv1alpha1.AddToScheme, resourcemanagerv1alpha1.AddToScheme, iamv1alpha1.AddToScheme, billingv1alpha1.AddToScheme,
	} {
		if err := add(s); err != nil {
			t.Fatal(err)
		}
	}
	return fake.NewClientBuilder().WithScheme(s).Build()
}

// assertInvalid checks that err is an Invalid API error for the given kind
// and name.
func assertInvalid(t *testing.T, err error, kind, name string) {
	t.Helper()
	if !apierrors.IsInvalid(err) {
		t.Fatalf("expected an Invalid API error, got %v", err)
	}
	details := err.(*apierrors.StatusError).ErrStatus.Details
	if details.Kind != kind || details.Name != name {
		t.Errorf("error is for %s %q, want %s %q", details.Kind, details.Name, kind, name)
	}
}

func TestInvalid(t *testing.T) {
	if err := invalid("Person", "jane", nil); err != nil {
		t.Errorf("no field errors should give no error, got %v", err)
	}
	err := invalid("Person", "jane", field.ErrorList{field.Invalid(field.NewPath("spec"), "x", "bad")})
	assertInvalid(t, err, "Person", "jane")
}
