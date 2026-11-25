// SPDX-License-Identifier: AGPL-3.0-only

package crm

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

// ValidatePersonCreate rejects a Person whose Company doesn't exist.
func ValidatePersonCreate(ctx context.Context, c client.Reader, person *crmv1alpha1.Person) field.ErrorList {
	_, errs := getCompany(ctx, c, field.NewPath("spec", "companyRef"), person.Spec.CompanyRef)
	return errs
}

// ValidatePersonUpdate rejects linking a Person to a User that doesn't
// exist. CRM sets the link through the status subresource. The Company can't
// change, which the schema enforces.
func ValidatePersonUpdate(ctx context.Context, c client.Reader, oldPerson, newPerson *crmv1alpha1.Person) field.ErrorList {
	var allErrs field.ErrorList
	if ref := newPerson.Status.UserRef; ref != nil && !sameUser(ref, oldPerson.Status.UserRef) {
		_, errs := getUser(ctx, c, field.NewPath("status", "userRef"), *ref)
		allErrs = append(allErrs, errs...)
	}
	return allErrs
}

func sameUser(a, b *crmv1alpha1.UserReference) bool {
	return a != nil && b != nil && *a == *b
}
