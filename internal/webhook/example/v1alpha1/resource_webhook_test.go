// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import (
	"context"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	examplev1alpha1 "go.miloapis.com/crm/api/example/v1alpha1"
)

func TestResourceDefault(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{name: "empty description is defaulted", description: "", want: ResourceDefaultDescription},
		{name: "existing description is kept", description: "Mine.", want: "Mine."},
	}

	webhook := &resourceWebhook{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &examplev1alpha1.Resource{Spec: examplev1alpha1.ResourceSpec{Description: tt.description}}
			if err := webhook.Default(context.Background(), resource); err != nil {
				t.Fatalf("Default() error = %v", err)
			}
			if resource.Spec.Description != tt.want {
				t.Fatalf("description = %q, want %q", resource.Spec.Description, tt.want)
			}
		})
	}
}

func TestResourceValidationReturnsInvalid(t *testing.T) {
	webhook := &resourceWebhook{}
	blank := &examplev1alpha1.Resource{
		ObjectMeta: metav1.ObjectMeta{Name: "example"},
		Spec:       examplev1alpha1.ResourceSpec{Description: " "},
	}
	protected := &examplev1alpha1.Resource{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "example",
			Annotations: map[string]string{examplev1alpha1.ResourceDeletionProtectionAnnotation: "true"},
		},
		Spec: examplev1alpha1.ResourceSpec{Description: "Protected."},
	}

	_, createErr := webhook.ValidateCreate(context.Background(), blank)
	_, updateErr := webhook.ValidateUpdate(context.Background(), protected, blank)
	_, deleteErr := webhook.ValidateDelete(context.Background(), protected)

	for name, err := range map[string]error{"create": createErr, "update": updateErr, "delete": deleteErr} {
		if !apierrors.IsInvalid(err) {
			t.Errorf("%s: expected an Invalid API error, got %v", name, err)
		}
	}
}
