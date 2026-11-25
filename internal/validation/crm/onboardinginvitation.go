// SPDX-License-Identifier: AGPL-3.0-only

package crm

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	iamv1alpha1 "go.miloapis.com/milo/pkg/apis/iam/v1alpha1"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

// ValidateOnboardingInvitationCreate rejects an invitation whose onboarding
// or Person doesn't exist, whose Person doesn't work for the onboarding's
// Company, or whose recipientEmail isn't the Person's email. The spec can't
// be changed afterwards, so it is only checked here; the Person's email may
// change later, and the invitation keeps the address it was sent to.
func ValidateOnboardingInvitationCreate(ctx context.Context, c client.Reader, invitation *crmv1alpha1.OnboardingInvitation) field.ErrorList {
	var allErrs field.ErrorList
	spec := field.NewPath("spec")

	onboarding, onboardingErrs := getOnboarding(ctx, c, spec.Child("onboardingRef"), invitation.Spec.OnboardingRef)
	allErrs = append(allErrs, onboardingErrs...)

	path := spec.Child("personRef")
	person, personErrs := getPerson(ctx, c, path, invitation.Spec.PersonRef)
	allErrs = append(allErrs, personErrs...)

	if len(personErrs) == 0 && !sameEmail(invitation.Spec.RecipientEmail, person.Spec.Email) {
		allErrs = append(allErrs, field.Invalid(spec.Child("recipientEmail"), invitation.Spec.RecipientEmail,
			"must be the Person's email, "+person.Spec.Email))
	}

	if len(onboardingErrs) == 0 && len(personErrs) == 0 {
		allErrs = append(allErrs, personOfCompany(path.Child("name"), person, onboarding.Spec.CompanyRef.Name, "onboarding")...)
	}
	return allErrs
}

// ValidateOnboardingInvitationUpdate checks the Users CRM records in status
// when they change. The inviter must exist. The User who redeemed the
// invitation must exist, have recipientEmail as their email, and have
// verified it: only they may redeem it.
func ValidateOnboardingInvitationUpdate(ctx context.Context, c client.Reader, oldInvitation, newInvitation *crmv1alpha1.OnboardingInvitation) field.ErrorList {
	var allErrs field.ErrorList
	status := field.NewPath("status")

	if ref := newInvitation.Status.InviterRef; ref != nil && !sameUser(ref, oldInvitation.Status.InviterRef) {
		_, errs := getUser(ctx, c, status.Child("inviterRef"), *ref)
		allErrs = append(allErrs, errs...)
	}

	if ref := newInvitation.Status.RedeemedUserRef; ref != nil && !sameUser(ref, oldInvitation.Status.RedeemedUserRef) {
		path := status.Child("redeemedUserRef", "name")
		user, errs := getUser(ctx, c, status.Child("redeemedUserRef"), *ref)
		if len(errs) > 0 {
			return append(allErrs, errs...)
		}
		allErrs = append(allErrs, redeemerMatches(path, user, newInvitation.Spec.RecipientEmail)...)
	}

	return allErrs
}

// redeemerMatches reports a User whose email isn't the invitation's
// recipient email, or who hasn't verified it.
func redeemerMatches(path *field.Path, user *iamv1alpha1.User, recipientEmail string) field.ErrorList {
	if !sameEmail(user.Spec.Email, recipientEmail) {
		return field.ErrorList{field.Invalid(path, user.Name,
			"User's email isn't the invitation's recipientEmail")}
	}
	if user.Status.EmailVerification != iamv1alpha1.EmailVerificationStateVerified {
		return field.ErrorList{field.Invalid(path, user.Name,
			"User hasn't verified their email")}
	}
	return nil
}

// sameEmail compares two email addresses, ignoring case and surrounding
// spaces.
func sameEmail(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}
