// SPDX-License-Identifier: AGPL-3.0-only

// Package crm holds admission validation for the crm.miloapis.com API group:
// the checks the CRD schemas can't express, because they read other records
// or depend on who is making the request. The webhooks call these functions
// and turn the returned field errors into API errors.
//
// References are checked when they are set or changed, not on every update,
// so deleting a record later never blocks unrelated edits to the records
// that name it. Nothing here deletes or changes the records references point
// to. Deleting is left to Milo's garbage collector: CRM's own records are
// owned by the record they belong to (internal/webhook/crm/v1alpha1/owners.go),
// and never own records other services keep, such as organizations.
package crm

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	billingv1alpha1 "go.miloapis.com/billing/api/v1alpha1"
	iamv1alpha1 "go.miloapis.com/milo/pkg/apis/iam/v1alpha1"
	resourcemanagerv1alpha1 "go.miloapis.com/milo/pkg/apis/resourcemanager/v1alpha1"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

// getCompany reads the Company a reference names. It returns a field error
// when the Company doesn't exist or can't be read; a read failure is an
// internal error, so the request is refused rather than let through.
func getCompany(ctx context.Context, c client.Reader, path *field.Path, ref crmv1alpha1.CompanyReference) (*crmv1alpha1.Company, field.ErrorList) {
	company := &crmv1alpha1.Company{}
	return company, get(ctx, c, path.Child("name"), "Company", ref.Name, company)
}

// getPerson reads the Person a reference names.
func getPerson(ctx context.Context, c client.Reader, path *field.Path, ref crmv1alpha1.PersonReference) (*crmv1alpha1.Person, field.ErrorList) {
	person := &crmv1alpha1.Person{}
	return person, get(ctx, c, path.Child("name"), "Person", ref.Name, person)
}

// getOpportunity reads the Opportunity a reference names.
func getOpportunity(ctx context.Context, c client.Reader, path *field.Path, ref crmv1alpha1.OpportunityReference) (*crmv1alpha1.Opportunity, field.ErrorList) {
	opportunity := &crmv1alpha1.Opportunity{}
	return opportunity, get(ctx, c, path.Child("name"), "Opportunity", ref.Name, opportunity)
}

// getOnboarding reads the OrganizationOnboarding a reference names.
func getOnboarding(ctx context.Context, c client.Reader, path *field.Path, ref crmv1alpha1.OrganizationOnboardingReference) (*crmv1alpha1.OrganizationOnboarding, field.ErrorList) {
	onboarding := &crmv1alpha1.OrganizationOnboarding{}
	return onboarding, get(ctx, c, path.Child("name"), "OrganizationOnboarding", ref.Name, onboarding)
}

// get reads a cluster-scoped record into obj.
func get(ctx context.Context, c client.Reader, path *field.Path, kind, name string, obj client.Object) field.ErrorList {
	return getIn(ctx, c, path, kind, "", name, obj)
}

// getIn reads a record into obj. A missing record is an invalid reference;
// any other read failure is an internal error, so the request is refused
// rather than let through.
func getIn(ctx context.Context, c client.Reader, path *field.Path, kind, namespace, name string, obj client.Object) field.ErrorList {
	err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, obj)
	switch {
	case err == nil:
		return nil
	case apierrors.IsNotFound(err):
		message := kind + " not found"
		if namespace != "" {
			message += " in namespace " + namespace
		}
		return field.ErrorList{field.Invalid(path, name, message)}
	default:
		return field.ErrorList{field.InternalError(path, fmt.Errorf("reading %s %q: %w", kind, name, err))}
	}
}

// organizationNamespace is the namespace Milo creates for an organization's
// own resources, such as its billing accounts.
func organizationNamespace(organization string) string {
	return "organization-" + organization
}

// getOrganization reads the Milo Organization a reference names.
func getOrganization(ctx context.Context, c client.Reader, path *field.Path, ref crmv1alpha1.OrganizationReference) field.ErrorList {
	return get(ctx, c, path.Child("name"), "Organization", ref.Name, &resourcemanagerv1alpha1.Organization{})
}

// getUser reads the Milo User a reference names.
func getUser(ctx context.Context, c client.Reader, path *field.Path, ref crmv1alpha1.UserReference) (*iamv1alpha1.User, field.ErrorList) {
	user := &iamv1alpha1.User{}
	return user, get(ctx, c, path.Child("name"), "User", ref.Name, user)
}

// getBillingAccount reads a BillingAccount in an organization's namespace.
func getBillingAccount(ctx context.Context, c client.Reader, path *field.Path, organization string, ref crmv1alpha1.BillingAccountReference) field.ErrorList {
	return getIn(ctx, c, path.Child("name"), "BillingAccount", organizationNamespace(organization), ref.Name, &billingv1alpha1.BillingAccount{})
}

// personOfCompany reports a Person who doesn't work for the given Company.
func personOfCompany(path *field.Path, person *crmv1alpha1.Person, company, owner string) field.ErrorList {
	if person.Spec.CompanyRef.Name == company {
		return nil
	}
	return field.ErrorList{field.Invalid(path, person.Name, fmt.Sprintf(
		"Person works for Company %q, not the %s's Company %q", person.Spec.CompanyRef.Name, owner, company))}
}
