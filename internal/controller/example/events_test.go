// SPDX-License-Identifier: AGPL-3.0-only

package example

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	examplev1alpha1 "go.miloapis.com/crm/api/example/v1alpha1"
)

type capturedEvent struct {
	eventType, reason, message string
	annotations                map[string]string
}

// captureRecorder records events in memory, including their annotations,
// which record.FakeRecorder doesn't expose.
type captureRecorder struct {
	events []capturedEvent
}

func (c *captureRecorder) Event(object runtime.Object, eventType, reason, message string) {
	c.AnnotatedEventf(object, nil, eventType, reason, "%s", message)
}

func (c *captureRecorder) Eventf(object runtime.Object, eventType, reason, messageFmt string, args ...any) {
	c.AnnotatedEventf(object, nil, eventType, reason, messageFmt, args...)
}

func (c *captureRecorder) AnnotatedEventf(_ runtime.Object, annotations map[string]string, eventType, reason, messageFmt string, args ...any) {
	c.events = append(c.events, capturedEvent{
		eventType:   eventType,
		reason:      reason,
		message:     fmt.Sprintf(messageFmt, args...),
		annotations: annotations,
	})
}

func readyResource(namespace, description string) *examplev1alpha1.Resource {
	return &examplev1alpha1.Resource{
		ObjectMeta: metav1.ObjectMeta{Name: "demo", Namespace: namespace},
		Spec:       examplev1alpha1.ResourceSpec{Description: description},
		Status: examplev1alpha1.ResourceStatus{
			Conditions: []metav1.Condition{{Type: ConditionTypeReady, Status: metav1.ConditionTrue}},
		},
	}
}

func TestRecordReadyTransition(t *testing.T) {
	ready := &metav1.Condition{Type: ConditionTypeReady, Status: metav1.ConditionTrue}
	notReady := &metav1.Condition{Type: ConditionTypeReady, Status: metav1.ConditionFalse}

	tests := []struct {
		name            string
		resource        *examplev1alpha1.Resource
		before          *metav1.Condition
		wantEvent       bool
		wantAnnotations map[string]string
	}{
		{
			name:      "first time ready in an organization namespace",
			resource:  readyResource("organization-acme", "Demo resource."),
			before:    nil,
			wantEvent: true,
			wantAnnotations: map[string]string{
				examplev1alpha1.ScopeTypeAnnotation:        "Organization",
				examplev1alpha1.ScopeNameAnnotation:        "acme",
				examplev1alpha1.EventAnnotationDescription: "Demo resource.",
			},
		},
		{
			name:      "ready again after not ready",
			resource:  readyResource("organization-acme", "Demo resource."),
			before:    notReady,
			wantEvent: true,
			wantAnnotations: map[string]string{
				examplev1alpha1.ScopeTypeAnnotation:        "Organization",
				examplev1alpha1.ScopeNameAnnotation:        "acme",
				examplev1alpha1.EventAnnotationDescription: "Demo resource.",
			},
		},
		{
			name:            "outside an organization namespace has no scope",
			resource:        readyResource("default", "Demo resource."),
			before:          nil,
			wantEvent:       true,
			wantAnnotations: map[string]string{examplev1alpha1.EventAnnotationDescription: "Demo resource."},
		},
		{
			name:     "already ready records nothing",
			resource: readyResource("organization-acme", "Demo resource."),
			before:   ready,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := &captureRecorder{}
			recordReadyTransition(recorder, tt.resource, tt.before)

			if !tt.wantEvent {
				if len(recorder.events) != 0 {
					t.Fatalf("expected no event, got %+v", recorder.events)
				}
				return
			}
			if len(recorder.events) != 1 {
				t.Fatalf("expected one event, got %d", len(recorder.events))
			}
			got := recorder.events[0]
			if got.eventType != corev1.EventTypeNormal || got.reason != examplev1alpha1.EventReasonResourceReady {
				t.Fatalf("event = %s/%s, want Normal/%s", got.eventType, got.reason, examplev1alpha1.EventReasonResourceReady)
			}
			if len(got.annotations) != len(tt.wantAnnotations) {
				t.Fatalf("annotations = %v, want %v", got.annotations, tt.wantAnnotations)
			}
			for k, v := range tt.wantAnnotations {
				if got.annotations[k] != v {
					t.Fatalf("annotation %s = %q, want %q", k, got.annotations[k], v)
				}
			}
		})
	}
}

func TestRecordReadyTransitionNilRecorder(t *testing.T) {
	recordReadyTransition(nil, readyResource("organization-acme", "Demo resource."), nil)
}
