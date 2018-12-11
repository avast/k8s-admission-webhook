// +build e2e

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIngressE2E(t *testing.T) {
	t.Run("Reject", func(t *testing.T) {
		err := applyManifest("test/manifests/ingress-valid.yaml", true)
		if assert.Nil(t, err) {
			err := applyManifest("test/manifests/ingress-collision-tls.yaml", true)
			assert.NotNil(t, err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		err := applyManifest("test/manifests/ingress-valid.yaml", true)
		if assert.Nil(t, err) {
			err := applyManifest("test/manifests/ingress-valid.yaml", true)
			if assert.Nil(t, err) {
				err := applyManifest("test/manifests/ingress-no-collision-tls.yaml", true)
				assert.Nil(t, err)
			}
		}
	})

}
