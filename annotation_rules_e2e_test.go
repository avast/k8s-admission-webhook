// +build e2e

package main

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestAnnotationRulesE2E(t *testing.T) {
	err := applyManifest("test/manifests/namespace.yaml", false)

	if assert.Nil(t, err) {
		err = applyManifest("test/webhook.yaml", false)
		time.Sleep(5 * time.Second)
		if assert.Nil(t, err) {
			t.Run("Reject Pod", func(t *testing.T) {
				err := applyManifest("test/manifests/annotation-rules-pod-valid.yaml", true)
				if assert.Nil(t, err, "annotation-rules-pod-valid.yaml should be applied") {
					err := applyManifest("test/manifests/annotation-rules-pod-invalid.yaml", true)
					assert.Error(t, err, "annotation-rules-pod-invalid should not be applied")
				}

			})
			t.Run("Accept Pod", func(t *testing.T) {
				err := applyManifest("test/manifests/annotation-rules-pod-valid.yaml", true)
				if assert.Nil(t, err) {
					err := applyManifest("test/manifests/annotation-rules-pod-valid.yaml", true)
					assert.Nil(t, err)
				}
			})
			t.Run("Reject Ingress", func(t *testing.T) {
				err := applyManifest("test/manifests/annotation-rules-ingress-valid.yaml", true)
				if assert.Nil(t, err, "annotation-rules-ingress-valid.yaml should be applied") {
					err := applyManifest("test/manifests/annotation-rules-ingress-invalid.yaml", true)
					assert.Error(t, err, "annotation-rules-ingress-invalid should not be applied")
				}

			})
			t.Run("Accept Ingress", func(t *testing.T) {
				err := applyManifest("test/manifests/annotation-rules-ingress-valid.yaml", true)
				if assert.Nil(t, err) {
					err := applyManifest("test/manifests/annotation-rules-ingress-valid.yaml", true)
					assert.Nil(t, err)
				}
			})
		}
	}
}
