// +build e2e

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIngressE2E(t *testing.T) {
	t.Run("Reject", func(t *testing.T) {
		err := applyManifest("test/manifests/ingress-valid.yaml", true)
		if assert.Nil(t, err, "ingress-valid.yaml should be applied") {
			err := applyManifest("test/manifests/ingress-collision-tls.yaml", true)
			assert.Error(t, err, "ingress-collision-tls.yaml should not be applied")
			err = applyManifest("test/manifests/ingress-collision-path.yaml", true)
			assert.Error(t, err, "ingress-collision-path.yaml should not be applied")
		}

	})

	t.Run("Success", func(t *testing.T) {
		err := applyManifest("test/manifests/ingress-valid.yaml", true)
		if assert.Nil(t, err) {
			err := applyManifest("test/manifests/ingress-valid.yaml", true)
			if assert.Nil(t, err) {
				err := applyManifest("test/manifests/ingress-no-collision.yaml", true)
				assert.Nil(t, err)
			}
		}
	})

}
