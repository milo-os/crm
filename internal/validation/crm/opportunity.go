// SPDX-License-Identifier: AGPL-3.0-only

package crm

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

// ValidateOpportunityCreate rejects an Opportunity whose Company or owner
// doesn't exist, or that names a Person who doesn't exist or works for
// another Company.
func ValidateOpportunityCreate(ctx context.Context, c client.Reader, opportunity *crmv1alpha1.Opportunity) field.ErrorList {
	return validateOpportunity(ctx, c, nil, opportunity)
}

// ValidateOpportunityUpdate applies the same checks to what changed: the
// owner, if it changed, and the People added. The Company can't change,
// which the schema enforces.
func ValidateOpportunityUpdate(ctx context.Context, c client.Reader, oldOpportunity, newOpportunity *crmv1alpha1.Opportunity) field.ErrorList {
	return validateOpportunity(ctx, c, oldOpportunity, newOpportunity)
}

func validateOpportunity(ctx context.Context, c client.Reader, oldOpportunity, opportunity *crmv1alpha1.Opportunity) field.ErrorList {
	var allErrs field.ErrorList
	spec := field.NewPath("spec")
	company := opportunity.Spec.CompanyRef.Name

	// Each reference is checked only when this request sets or changes it
	// (oldOpportunity is nil on create, so everything is checked then). A
	// record referenced earlier may have been deleted since; checking it again
	// would block every later edit, even ones unrelated to it.
	if oldOpportunity == nil {
		_, errs := getCompany(ctx, c, spec.Child("companyRef"), opportunity.Spec.CompanyRef)
		allErrs = append(allErrs, errs...)
	}

	if oldOpportunity == nil || oldOpportunity.Spec.OwnerRef != opportunity.Spec.OwnerRef {
		_, errs := getUser(ctx, c, spec.Child("ownerRef"), opportunity.Spec.OwnerRef)
		allErrs = append(allErrs, errs...)
	}

	// The People already on the Opportunity before this request; only the
	// ones it adds are checked.
	known := map[string]bool{}
	if oldOpportunity != nil {
		for _, ref := range oldOpportunity.Spec.PersonRefs {
			known[ref.Name] = true
		}
	}
	for i, ref := range opportunity.Spec.PersonRefs {
		if known[ref.Name] {
			continue
		}
		path := spec.Child("personRefs").Index(i)
		person, errs := getPerson(ctx, c, path, ref)
		if len(errs) > 0 {
			allErrs = append(allErrs, errs...)
			continue
		}
		allErrs = append(allErrs, personOfCompany(path.Child("name"), person, company, "Opportunity")...)
	}

	return allErrs
}
