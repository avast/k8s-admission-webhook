// +build crosscheck

package main

import (
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

/*

This tests loads all ingress definitions and performs the check for all of
them to find collisions in the current kubernetes cluster.

Current .kube/config context is used to initialize the kubernetes client.

This should not be part of any automated tests. It's meant to be executed manually when needed.

*/
func TestClusterCollisions(t *testing.T) {
	initLogger()
	InitKubeClientSet(false)

	t.Run("Cross cluster validation", func(t *testing.T) {
		remoteIngresses, err := IngressClientAllNamespaces().List(metav1.ListOptions{})
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
