// SPDX-License-Identifier: AGPL-3.0-only

package crm

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

// ValidatePersonCreate rejects a Person whose Company doesn't exist, or that
// is owned by anything but its Company.
func ValidatePersonCreate(ctx context.Context, c client.Reader, person *crmv1alpha1.Person) field.ErrorList {
	_, errs := getCompany(ctx, c, field.NewPath("spec", "companyRef"), person.Spec.CompanyRef)
	return append(errs, validateOwnedOnlyBy(person.OwnerReferences, "Company", person.Spec.CompanyRef.Name)...)
}

// ValidatePersonUpdate rejects linking a Person to a User that doesn't
// exist, and changing its owner references. CRM sets the link through the
// status subresource. The Company can't change, which the schema enforces.
func ValidatePersonUpdate(ctx context.Context, c client.Reader, oldPerson, newPerson *crmv1alpha1.Person) field.ErrorList {
	allErrs := validateOwnersUnchanged(oldPerson.OwnerReferences, newPerson.OwnerReferences)
	if ref := newPerson.Status.UserRef; ref != nil && !sameUser(ref, oldPerson.Status.UserRef) {
		_, errs := getUser(ctx, c, field.NewPath("status", "userRef"), *ref)
		allErrs = append(allErrs, errs...)
	}
	return allErrs
}

func sameUser(a, b *crmv1alpha1.UserReference) bool {
	return a != nil && b != nil && *a == *b
}
