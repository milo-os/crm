// SPDX-License-Identifier: AGPL-3.0-only

package example

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	examplev1alpha1 "go.miloapis.com/crm/api/example/v1alpha1"
)

func newResource(description string, annotations map[string]string) *examplev1alpha1.Resource {
	return &examplev1alpha1.Resource{
		ObjectMeta: metav1.ObjectMeta{Name: "example", Annotations: annotations},
		Spec:       examplev1alpha1.ResourceSpec{Description: description},
	}
}

func TestValidateResourceCreate(t *testing.T) {
	tests := []struct {
		name        string
		description string
		wantErr     string
	}{
		{name: "valid description", description: "An example resource."},
		{name: "empty description", description: "", wantErr: "must not be blank"},
		{name: "whitespace description", description: "   ", wantErr: "must not be blank"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrors(t, ValidateResourceCreate(newResource(tt.description, nil)).ToAggregate(), tt.wantErr)
		})
	}
}

func TestValidateResourceUpdate(t *testing.T) {
	oldResource := newResource("An example resource.", nil)

	tests := []struct {
		name        string
		description string
		wantErr     string
	}{
		{name: "changed description", description: "Updated description."},
		{name: "blank description", description: "\t", wantErr: "must not be blank"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateResourceUpdate(oldResource, newResource(tt.description, nil))
			assertErrors(t, errs.ToAggregate(), tt.wantErr)
		})
	}
}

func TestValidateResourceDelete(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		wantErr     string
	}{
		{name: "no annotations"},
		{
			name:        "protection disabled",
			annotations: map[string]string{examplev1alpha1.ResourceDeletionProtectionAnnotation: "false"},
		},
		{
			name:        "protection enabled",
			annotations: map[string]string{examplev1alpha1.ResourceDeletionProtectionAnnotation: "true"},
			wantErr:     "deletion protection is enabled",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateResourceDelete(newResource("An example resource.", tt.annotations))
			assertErrors(t, errs.ToAggregate(), tt.wantErr)
		})
	}
}

func assertErrors(t *testing.T, err error, wantErr string) {
	t.Helper()
	switch {
	case wantErr == "" && err != nil:
		t.Fatalf("expected no error, got %v", err)
	case wantErr != "" && err == nil:
		t.Fatalf("expected error containing %q, got none", wantErr)
	case wantErr != "" && !strings.Contains(err.Error(), wantErr):
		t.Fatalf("expected error containing %q, got %v", wantErr, err)
	}
}
