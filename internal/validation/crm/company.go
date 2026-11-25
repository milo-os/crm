// SPDX-License-Identifier: AGPL-3.0-only

package crm

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

// ValidateCompanyCreate rejects a Company that links an organization that
// doesn't exist.
func ValidateCompanyCreate(ctx context.Context, c client.Reader, company *crmv1alpha1.Company) field.ErrorList {
	return validateCompany(ctx, c, nil, company)
}

// ValidateCompanyUpdate checks only the organizations newly linked.
func ValidateCompanyUpdate(ctx context.Context, c client.Reader, oldCompany, newCompany *crmv1alpha1.Company) field.ErrorList {
	return validateCompany(ctx, c, oldCompany, newCompany)
}

func validateCompany(ctx context.Context, c client.Reader, oldCompany, company *crmv1alpha1.Company) field.ErrorList {
	var allErrs field.ErrorList

	// The organizations the Company already linked before this request. On
	// create there are none, so every organization is checked.
	linked := map[string]bool{}
	if oldCompany != nil {
		for _, ref := range oldCompany.Spec.OrganizationRefs {
			linked[ref.Name] = true
		}
	}

	for i, ref := range company.Spec.OrganizationRefs {
		// Only check organizations this request adds. One linked earlier may
		// have been deleted since; checking it again would block every later
		// edit to the Company, even ones unrelated to it.
		if linked[ref.Name] {
			continue
		}
		// The index makes the error point at the exact entry, such as
		// spec.organizationRefs[2].name.
		allErrs = append(allErrs, getOrganization(ctx, c, field.NewPath("spec", "organizationRefs").Index(i), ref)...)
	}
	return allErrs
}
