// SPDX-License-Identifier: AGPL-3.0-only

// Package v1alpha1 contains API Schema definitions for the crm.miloapis.com v1alpha1 API group.
//
// Every CRM resource is namespaced and lives in the selling organization's
// namespace (organization-<name>). See docs/design/opportunity-led-onboarding.md.
// +kubebuilder:object:generate=true
// +groupName=crm.miloapis.com
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "crm.miloapis.com", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &runtime.SchemeBuilder{}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
