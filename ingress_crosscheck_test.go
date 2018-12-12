// +build crosscheck

package main

import (
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestClusterCollisions(t *testing.T) {
	initLogger()
	initKubeClientSet(false)

	t.Run("Cross cluster validation", func(t *testing.T) {
		remoteIngresses, err := ingressClient().List(metav1.ListOptions{})
		if assert.Nil(t, err) {
			for _, ingress := range remoteIngresses.Items {
				logger.Debugf("Processing ingress %s", ingress.Name)

				validation := &objectValidation{ingress.Kind, nil, &validationViolationSet{}}
				config := &config{RuleIngressCollision: true}
				err := ValidateIngress(validation, &ingress, config)
				if assert.Nil(t, err) {
					assert.Len(t, validation.Violations.Violations, 0)
				}
			}
		}

	})

}
