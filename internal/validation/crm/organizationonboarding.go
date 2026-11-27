// SPDX-License-Identifier: AGPL-3.0-only

package crm

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

// ValidateOrganizationOnboardingCreate rejects an onboarding whose Company
// or Opportunity doesn't exist, whose Opportunity is with another Company,
// or that is owned by anything but its Company. Both references can't be changed afterwards, so they are only
// checked here.
func ValidateOrganizationOnboardingCreate(ctx context.Context, c client.Reader, onboarding *crmv1alpha1.OrganizationOnboarding) field.ErrorList {
	var allErrs field.ErrorList
	spec := field.NewPath("spec")
	company := onboarding.Spec.CompanyRef.Name

	_, errs := getCompany(ctx, c, spec.Child("companyRef"), onboarding.Spec.CompanyRef)
	allErrs = append(allErrs, errs...)
	allErrs = append(allErrs, validateOwnedOnlyBy(onboarding.OwnerReferences, "Company", company)...)

	path := spec.Child("opportunityRef")
	opportunity, errs := getOpportunity(ctx, c, path, onboarding.Spec.OpportunityRef)
	if len(errs) > 0 {
		return append(allErrs, errs...)
	}
	if opportunity.Spec.CompanyRef.Name != company {
		allErrs = append(allErrs, field.Invalid(path.Child("name"), opportunity.Name, fmt.Sprintf(
			"Opportunity is with Company %q, not the onboarding's Company %q", opportunity.Spec.CompanyRef.Name, company)))
	}
	return allErrs
}

// ValidateOrganizationOnboardingUpdate keeps the owner references and checks
// the references CRM records in status. Who can change the details is up to
// IAM; the schema keeps the Company, the Opportunity, and, once the
// onboarding is sent, its currency and payment arrangement.
func ValidateOrganizationOnboardingUpdate(ctx context.Context, c client.Reader, oldOnboarding, newOnboarding *crmv1alpha1.OrganizationOnboarding) field.ErrorList {
	allErrs := validateOwnersUnchanged(oldOnboarding.OwnerReferences, newOnboarding.OwnerReferences)
	return append(allErrs, validateOnboardingStatus(ctx, c, &oldOnboarding.Status, newOnboarding)...)
}

// validateOnboardingStatus checks the status references that changed: the
// current invitation must exist and be for this onboarding, the organization
// must exist and be linked to the onboarding's Company, the billing account
// must exist in that organization's namespace, and the User who authorized
// invoicing must exist.
func validateOnboardingStatus(ctx context.Context, c client.Reader, old *crmv1alpha1.OrganizationOnboardingStatus, onboarding *crmv1alpha1.OrganizationOnboarding) field.ErrorList {
	var allErrs field.ErrorList
	status := &onboarding.Status
	path := field.NewPath("status")

	if ref := status.CurrentInvitationRef; ref != nil && !equality.Semantic.DeepEqual(ref, old.CurrentInvitationRef) {
		refPath := path.Child("currentInvitationRef", "name")
		invitation := &crmv1alpha1.OnboardingInvitation{}
		errs := get(ctx, c, refPath, "OnboardingInvitation", ref.Name, invitation)
		allErrs = append(allErrs, errs...)
		if len(errs) == 0 && invitation.Spec.OnboardingRef.Name != onboarding.Name {
			allErrs = append(allErrs, field.Invalid(refPath, ref.Name, fmt.Sprintf(
				"OnboardingInvitation is for onboarding %q, not %q", invitation.Spec.OnboardingRef.Name, onboarding.Name)))
		}
	}

	organizationChanged := !equality.Semantic.DeepEqual(status.OrganizationRef, old.OrganizationRef)
	if ref := status.OrganizationRef; ref != nil && organizationChanged {
		refPath := path.Child("organizationRef")
		errs := getOrganization(ctx, c, refPath, *ref)
		allErrs = append(allErrs, errs...)
		if len(errs) == 0 {
			allErrs = append(allErrs, organizationOfCompany(ctx, c, refPath.Child("name"), ref.Name, onboarding.Spec.CompanyRef)...)
		}
	}

	billingChanged := !equality.Semantic.DeepEqual(status.BillingAccountRef, old.BillingAccountRef)
	if ref := status.BillingAccountRef; ref != nil && (billingChanged || organizationChanged) {
		refPath := path.Child("billingAccountRef")
		if status.OrganizationRef == nil {
			allErrs = append(allErrs, field.Invalid(refPath.Child("name"), ref.Name,
				"a billing account needs status.organizationRef, whose namespace it lives in"))
		} else {
			allErrs = append(allErrs, getBillingAccount(ctx, c, refPath, status.OrganizationRef.Name, *ref)...)
		}
	}

	if auth := status.InvoiceAuthorization; auth != nil && (old.InvoiceAuthorization == nil || auth.UserRef != old.InvoiceAuthorization.UserRef) {
		_, errs := getUser(ctx, c, path.Child("invoiceAuthorization", "userRef"), auth.UserRef)
		allErrs = append(allErrs, errs...)
	}

	return allErrs
}

// organizationOfCompany reports an Organization the onboarding's Company
// doesn't link in spec.organizationRefs, so an onboarding can never record
// another customer's Organization. CRM links the Organization to the Company
// before recording it on the onboarding.
func organizationOfCompany(ctx context.Context, c client.Reader, path *field.Path, organization string, ref crmv1alpha1.CompanyReference) field.ErrorList {
	company, errs := getCompany(ctx, c, field.NewPath("spec", "companyRef"), ref)
	if len(errs) > 0 {
		return errs
	}
	for _, linked := range company.Spec.OrganizationRefs {
		if linked.Name == organization {
			return nil
		}
	}
	return field.ErrorList{field.Invalid(path, organization, fmt.Sprintf(
		"Organization isn't linked to the onboarding's Company %q; link it in the Company's organizationRefs first", ref.Name))}
}
