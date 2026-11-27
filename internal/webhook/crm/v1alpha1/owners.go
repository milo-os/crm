// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"
	"fmt"
	"net/http"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

// CRM's own records form one ownership tree, so deleting a Company deletes
// everything about it through Milo's garbage collector:
//
//	Company
//	├── Person
//	├── Opportunity
//	└── OrganizationOnboarding
//	    └── OnboardingInvitation
//
// Each record has exactly one owner: the garbage collector only deletes a
// record once all its owners are gone. Records other Milo services own, such
// as the Organization and BillingAccount an onboarding creates, are never
// owned by CRM records.
//
// TODO(#13): an invitation is owned by its onboarding, not its Person, so
// deleting a Person doesn't delete their invitations. CRM's controller
// deletes them, with a finalizer on the Person.

// +kubebuilder:webhook:path=/mutate-crm-miloapis-com-v1alpha1-person,mutating=true,failurePolicy=fail,sideEffects=None,groups=crm.miloapis.com,resources=people,verbs=create,versions=v1alpha1,name=mperson.crm.miloapis.com,admissionReviewVersions=v1,serviceName=crm-webhook
// +kubebuilder:webhook:path=/mutate-crm-miloapis-com-v1alpha1-opportunity,mutating=true,failurePolicy=fail,sideEffects=None,groups=crm.miloapis.com,resources=opportunities,verbs=create,versions=v1alpha1,name=mopportunity.crm.miloapis.com,admissionReviewVersions=v1,serviceName=crm-webhook
// +kubebuilder:webhook:path=/mutate-crm-miloapis-com-v1alpha1-organizationonboarding,mutating=true,failurePolicy=fail,sideEffects=None,groups=crm.miloapis.com,resources=organizationonboardings,verbs=create,versions=v1alpha1,name=morganizationonboarding.crm.miloapis.com,admissionReviewVersions=v1,serviceName=crm-webhook
// +kubebuilder:webhook:path=/mutate-crm-miloapis-com-v1alpha1-onboardinginvitation,mutating=true,failurePolicy=fail,sideEffects=None,groups=crm.miloapis.com,resources=onboardinginvitations,verbs=create,versions=v1alpha1,name=monboardinginvitation.crm.miloapis.com,admissionReviewVersions=v1,serviceName=crm-webhook

// ownerDefaulter makes a new record owned by the record it belongs to. It
// only runs on create: the reference it follows can't change.
//
// It edits the request's own JSON instead of a typed object, so its patch
// only adds the owner reference. Marshalling a typed object would also add
// every field the client left out as an empty value, and the schema would
// then report those fields as invalid instead of required.
type ownerDefaulter struct {
	reader client.Reader
	scheme *runtime.Scheme
	// newOwner returns an empty object of the owner's kind.
	newOwner func() client.Object
	// ownerName is the path in the record to its owner's name.
	ownerName []string
}

var _ admission.Handler = &ownerDefaulter{}

// registerOwnerDefaulters registers the defaulting webhook of every CRM kind
// that has an owner.
func registerOwnerDefaulters(server webhook.Server, reader client.Reader, scheme *runtime.Scheme) {
	company := func() client.Object { return &crmv1alpha1.Company{} }
	onboarding := func() client.Object { return &crmv1alpha1.OrganizationOnboarding{} }
	for path, d := range map[string]*ownerDefaulter{
		"/mutate-crm-miloapis-com-v1alpha1-person":                 {newOwner: company, ownerName: []string{"spec", "companyRef", "name"}},
		"/mutate-crm-miloapis-com-v1alpha1-opportunity":            {newOwner: company, ownerName: []string{"spec", "companyRef", "name"}},
		"/mutate-crm-miloapis-com-v1alpha1-organizationonboarding": {newOwner: company, ownerName: []string{"spec", "companyRef", "name"}},
		"/mutate-crm-miloapis-com-v1alpha1-onboardinginvitation":   {newOwner: onboarding, ownerName: []string{"spec", "onboardingRef", "name"}},
	} {
		d.reader, d.scheme = reader, scheme
		server.Register(path, &webhook.Admission{Handler: d})
	}
}

// Handle adds an owner reference to the record's owner. A missing owner name
// is left to the schema, and a missing owner to the validating webhook,
// which reject them with clearer errors; any other read failure refuses the
// request.
func (d *ownerDefaulter) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &unstructured.Unstructured{}
	if err := obj.UnmarshalJSON(req.Object.Raw); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	name, _, _ := unstructured.NestedString(obj.Object, d.ownerName...)
	if name == "" {
		return admission.Allowed("")
	}

	owner := d.newOwner()
	if err := d.reader.Get(ctx, client.ObjectKey{Name: name}, owner); err != nil {
		if apierrors.IsNotFound(err) {
			return admission.Allowed("")
		}
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("reading owner %q: %w", name, err))
	}
	if err := controllerutil.SetOwnerReference(owner, obj, d.scheme); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	defaulted, err := obj.MarshalJSON()
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, defaulted)
}
