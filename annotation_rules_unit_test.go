package main

import (
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func createObject(name string, namespace string, annotationNames []string) *metav1.ObjectMeta {
	annotations := make(map[string]string)
	for _,annotationName := range annotationNames {
		annotations[annotationName] = ""
	}
	return &metav1.ObjectMeta{
		Name:                       name,
		GenerateName:               name,
		Namespace:                  namespace,
		Annotations:                annotations,
	}
}

func TestAnnotationRules(t *testing.T) {
	initLogger()
	t.Run("Annotation Rules", func(t *testing.T) {
		t.Run("Matching Kind", func(t *testing.T) {
			t.Run("AllowAll must allow any prefix", func(t *testing.T) {
				validation := &objectValidation{"Pod", &metav1.ObjectMeta{}, &validationViolationSet{}}
				annotationRules := &AnnotationRules {
					AnnotationRules: map[string]AnnotationRule{
						"Pod": {
							Policy:     "AllowAll",
							Exceptions: nil,
						},
					},
				}
				validateAnnotationsByRules(validation, createObject("pod-1", "test", []string{"annotation1"}), "Pod", annotationRules)
				assert.Empty(t, validation.Violations)
			})
			t.Run("DenyAll must deny any prefix", func(t *testing.T) {
				validation := &objectValidation{"Pod", &metav1.ObjectMeta{}, &validationViolationSet{}}
				annotationRules := &AnnotationRules {
					AnnotationRules: map[string]AnnotationRule{
						"Pod": {
							Policy:     "DenyAll",
							Exceptions: nil,
						},
					},
				}
				validateAnnotationsByRules(validation, createObject("pod-1", "test", []string{"annotation1"}), "Pod", annotationRules)
				assert.Len(t, validation.Violations.Violations, 1)
			})
			t.Run("AllowAll with exception must deny exception prefix", func(t *testing.T) {
				validation := &objectValidation{"Pod", &metav1.ObjectMeta{}, &validationViolationSet{}}
				annotationRules := &AnnotationRules {
					AnnotationRules: map[string]AnnotationRule{
						"Pod": {
							Policy:     "AllowAll",
							Exceptions: []string{"annotation1"},
						},
					},
				}
				validateAnnotationsByRules(validation, createObject("pod-1", "test", []string{"annotation1"}), "Pod", annotationRules)
				assert.Len(t, validation.Violations.Violations, 1)
			})
			t.Run("DenyAll with exception must allow exception prefix", func(t *testing.T) {
				validation := &objectValidation{"Pod", &metav1.ObjectMeta{}, &validationViolationSet{}}
				annotationRules := &AnnotationRules {
					AnnotationRules: map[string]AnnotationRule{
						"Pod": {
							Policy:     "DenyAll",
							Exceptions: []string{"annotation1"},
						},
					},
				}
				validateAnnotationsByRules(validation, createObject("pod-1", "test", []string{"annotation1"}), "Pod", annotationRules)
				assert.Empty(t, validation.Violations)
			})

		})
		t.Run("NOT Matching Kind", func(t *testing.T) {
			t.Run("AllowAll must allow any prefix", func(t *testing.T) {
				validation := &objectValidation{"Pod", &metav1.ObjectMeta{}, &validationViolationSet{}}
				annotationRules := &AnnotationRules{
					AnnotationRules: map[string]AnnotationRule{
						"Deployment": {
							Policy:     "AllowAll",
							Exceptions: nil,
						},
					},
				}
				validateAnnotationsByRules(validation, createObject("pod-1", "test", []string{"annotation1"}), "Pod", annotationRules)
				assert.Empty(t, validation.Violations)
			})
			t.Run("DenyAll must allow any prefix", func(t *testing.T) {
				validation := &objectValidation{"Pod", &metav1.ObjectMeta{}, &validationViolationSet{}}
				annotationRules := &AnnotationRules{
					AnnotationRules: map[string]AnnotationRule{
						"Deployment": {
							Policy:     "DenyAll",
							Exceptions: nil,
						},
					},
				}
				validateAnnotationsByRules(validation, createObject("pod-1", "test", []string{"annotation1"}), "Pod", annotationRules)
				assert.Empty(t, validation.Violations)
			})
		})
	})
}