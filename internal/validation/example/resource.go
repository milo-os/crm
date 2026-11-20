// SPDX-License-Identifier: AGPL-3.0-only

// Package example holds admission validation for the example.miloapis.com
// API group. The webhooks call these functions and turn the returned field
// errors into API errors, so the rules stay testable without a webhook.
package example

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"

	examplev1alpha1 "go.miloapis.com/crm/api/example/v1alpha1"
)

// ValidateResourceCreate validates a Resource on creation.
func ValidateResourceCreate(resource *examplev1alpha1.Resource) field.ErrorList {
	return validateResourceSpec(resource)
}

// ValidateResourceUpdate validates a Resource on update.
func ValidateResourceUpdate(_, newResource *examplev1alpha1.Resource) field.ErrorList {
	return validateResourceSpec(newResource)
}

// ValidateResourceDelete refuses deletion while deletion protection is
// enabled on the Resource.
func ValidateResourceDelete(resource *examplev1alpha1.Resource) field.ErrorList {
	var allErrs field.ErrorList

	if resource.GetAnnotations()[examplev1alpha1.ResourceDeletionProtectionAnnotation] == "true" {
		allErrs = append(allErrs, field.Forbidden(
			field.NewPath("metadata", "annotations").Key(examplev1alpha1.ResourceDeletionProtectionAnnotation),
			"deletion protection is enabled; remove the annotation to delete this resource",
		))
	}

	return allErrs
}

// validateResourceSpec rejects a blank description. The defaulting webhook
// fills in an empty description, so only whitespace-only values reach here.
func validateResourceSpec(resource *examplev1alpha1.Resource) field.ErrorList {
	var allErrs field.ErrorList

	if strings.TrimSpace(resource.Spec.Description) == "" {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec", "description"),
			resource.Spec.Description,
			"must not be blank",
		))
	}

	return allErrs
}
