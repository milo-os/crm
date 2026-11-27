// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"
	"errors"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

func createRequest(raw string) admission.Request {
	return admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
		Operation: admissionv1.Create,
		Object:    runtime.RawExtension{Raw: []byte(raw)},
	}}
}

func companyOwned(reader client.Reader, s *runtime.Scheme) *ownerDefaulter {
	return &ownerDefaulter{reader: reader, scheme: s,
		newOwner:  func() client.Object { return &crmv1alpha1.Company{} },
		ownerName: []string{"spec", "companyRef", "name"}}
}

func TestOwnerDefaulter(t *testing.T) {
	s := runtime.NewScheme()
	if err := crmv1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	acme := &crmv1alpha1.Company{ObjectMeta: metav1.ObjectMeta{Name: "acme", UID: types.UID("acme-uid")}}
	onboarding := &crmv1alpha1.OrganizationOnboarding{ObjectMeta: metav1.ObjectMeta{Name: "acme-onboarding", UID: types.UID("onboarding-uid")}}
	reader := fake.NewClientBuilder().WithScheme(s).WithObjects(acme, onboarding).Build()
	ctx := context.Background()

	t.Run("adds only the owner reference to the Company", func(t *testing.T) {
		// givenName is missing on purpose: the patch must not add it, so
		// the schema still reports it as required.
		resp := companyOwned(reader, s).Handle(ctx, createRequest(
			`{"apiVersion":"crm.miloapis.com/v1alpha1","kind":"Person","metadata":{"name":"jane"},"spec":{"email":"jane@example.com","companyRef":{"name":"acme"}}}`))
		if !resp.Allowed {
			t.Fatalf("expected the request to be allowed, got %v", resp.Result)
		}
		if len(resp.Patches) != 1 || resp.Patches[0].Path != "/metadata/ownerReferences" {
			t.Fatalf("expected one patch to /metadata/ownerReferences, got %+v", resp.Patches)
		}
		refs := resp.Patches[0].Value.([]any)
		ref := refs[0].(map[string]any)
		want := map[string]any{"apiVersion": "crm.miloapis.com/v1alpha1", "kind": "Company", "name": "acme", "uid": "acme-uid"}
		for k, v := range want {
			if ref[k] != v {
				t.Errorf("owner reference %s = %v, want %v", k, ref[k], v)
			}
		}
		if _, ok := ref["controller"]; ok {
			t.Error("owner reference should not be a controller reference")
		}
		if _, ok := ref["blockOwnerDeletion"]; ok {
			t.Error("owner reference should not block the owner's deletion")
		}
	})

	t.Run("an invitation is owned by its onboarding", func(t *testing.T) {
		d := &ownerDefaulter{reader: reader, scheme: s,
			newOwner:  func() client.Object { return &crmv1alpha1.OrganizationOnboarding{} },
			ownerName: []string{"spec", "onboardingRef", "name"}}
		resp := d.Handle(ctx, createRequest(
			`{"apiVersion":"crm.miloapis.com/v1alpha1","kind":"OnboardingInvitation","metadata":{"name":"i"},"spec":{"onboardingRef":{"name":"acme-onboarding"},"personRef":{"name":"jane"}}}`))
		if len(resp.Patches) != 1 {
			t.Fatalf("expected one patch, got %+v", resp.Patches)
		}
		ref := resp.Patches[0].Value.([]any)[0].(map[string]any)
		if ref["kind"] != "OrganizationOnboarding" || ref["name"] != "acme-onboarding" {
			t.Errorf("owner reference is %v, want the onboarding", ref)
		}
	})

	t.Run("no owner name is left to the schema", func(t *testing.T) {
		resp := companyOwned(reader, s).Handle(ctx, createRequest(
			`{"apiVersion":"crm.miloapis.com/v1alpha1","kind":"Person","metadata":{"name":"jane"},"spec":{}}`))
		if !resp.Allowed || len(resp.Patches) != 0 {
			t.Errorf("expected allowed with no patch, got allowed=%v patches=%+v", resp.Allowed, resp.Patches)
		}
	})

	t.Run("a missing owner is left to validation", func(t *testing.T) {
		resp := companyOwned(reader, s).Handle(ctx, createRequest(
			`{"apiVersion":"crm.miloapis.com/v1alpha1","kind":"Person","metadata":{"name":"jane"},"spec":{"companyRef":{"name":"initech"}}}`))
		if !resp.Allowed || len(resp.Patches) != 0 {
			t.Errorf("expected allowed with no patch, got allowed=%v patches=%+v", resp.Allowed, resp.Patches)
		}
	})

	t.Run("a read failure refuses the request", func(t *testing.T) {
		failing := fake.NewClientBuilder().WithScheme(s).WithInterceptorFuncs(interceptor.Funcs{
			Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
				return errors.New("connection refused")
			},
		}).Build()
		resp := companyOwned(failing, s).Handle(ctx, createRequest(
			`{"apiVersion":"crm.miloapis.com/v1alpha1","kind":"Person","metadata":{"name":"jane"},"spec":{"companyRef":{"name":"acme"}}}`))
		if resp.Allowed {
			t.Error("expected the request to be refused")
		}
	})
}
