// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

// CompanyReference names a Company. The Company lives in the selling
// organization's namespace, like every CRM resource, so the reference holds
// only its name.
type CompanyReference struct {
	// Name is the Company's name.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:XValidation:rule="self.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?([.][a-z0-9]([-a-z0-9]*[a-z0-9])?)*$')",message="must be a valid Company name: lowercase letters, digits, '-', and '.', starting and ending with a letter or digit"
	Name string `json:"name"`
}

// OrganizationReference names a Milo organization
// (resourcemanager.miloapis.com Organization).
type OrganizationReference struct {
	// Name is the organization's name.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:XValidation:rule="self.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')",message="must be a valid organization name: lowercase letters, digits, and '-', starting and ending with a letter or digit"
	Name string `json:"name"`
}

// UserReference names a Milo User (iam.miloapis.com User).
type UserReference struct {
	// Name is the User's name.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:XValidation:rule="self.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?([.][a-z0-9]([-a-z0-9]*[a-z0-9])?)*$')",message="must be a valid User name: lowercase letters, digits, '-', and '.', starting and ending with a letter or digit"
	Name string `json:"name"`
}
