// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// CompanySpec is the customer relationship sales manages.
type CompanySpec struct {
	// DisplayName is the Company's human-readable name.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	// +kubebuilder:validation:XValidation:rule="!self.matches('^[[:space:]]*$')",message="must not be blank"
	DisplayName string `json:"displayName"`

	// Website is the Company's public website, an absolute http or https URL.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=2048
	// +kubebuilder:validation:XValidation:rule="isURL(self) && url(self).getScheme() in ['http', 'https'] && url(self).getHostname() != ''",message="must be an absolute http or https URL, such as https://example.com"
	Website string `json:"website,omitempty"`

	// OrganizationRefs links the Company to the Milo organizations created for
	// it, normally one. These are links, not ownership: deleting the Company
	// never deletes an organization.
	//
	// +optional
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=50
	OrganizationRefs []OrganizationReference `json:"organizationRefs,omitempty"`
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

// Company is a customer relationship that sales manages in Milo. It is
// cluster-scoped and managed only by Milo staff.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Display Name",type=string,JSONPath=`.spec.displayName`
// +kubebuilder:printcolumn:name="Website",type=string,JSONPath=`.spec.website`
// +kubebuilder:printcolumn:name="Organizations",type=string,JSONPath=`.spec.organizationRefs[*].name`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Company struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CompanySpec `json:"spec"`
}

// CompanyList contains a list of Companies.
//
// +kubebuilder:object:root=true
type CompanyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Company `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(GroupVersion, &Company{}, &CompanyList{})
		metav1.AddToGroupVersion(scheme, GroupVersion)
		return nil
	})
}
