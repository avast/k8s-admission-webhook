// +build e2e

package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var shouldFailWithResourcesNotSpecifiedErrors = []string{
	"test/manifests/deployment-incomplete.yaml",
	"test/manifests/pod-incomplete.yaml",
	"test/manifests/job-incomplete.yaml",
	"test/manifests/cronjob-incomplete.yaml",
	"test/manifests/statefulset-incomplete.yaml",
}
var shouldFailWithResourcesMustBeNonZeroErrors = []string{
	"test/manifests/deployment-zero.yaml",
	"test/manifests/pod-zero.yaml",
	"test/manifests/job-zero.yaml",
	"test/manifests/cronjob-zero.yaml",
	"test/manifests/statefulset-zero.yaml",
}
var shouldSucceed = []string{
	"test/manifests/deployment-complete.yaml",
	"test/manifests/pod-complete.yaml",
	"test/manifests/job-complete.yaml",
	"test/manifests/cronjob-complete.yaml",
	"test/manifests/statefulset-complete.yaml",
}

func TestManifests(t *testing.T) {
	err := applyManifest("test/manifests/namespace.yaml", false)
	if assert.Nil(t, err) {
		for _, p := range shouldFailWithResourcesNotSpecifiedErrors {
			t.Run(fmt.Sprintf("%s should fail because resource limits and requests are not specified", p), func(t *testing.T) {
				err = applyManifest(p, true)
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), "'cpu' resource limit must be specified")
					assert.Contains(t, err.Error(), "'memory' resource limit must be specified")
					assert.Contains(t, err.Error(), "'cpu' resource request must be specified")
					assert.Contains(t, err.Error(), "'memory' resource request must be specified")
				}
			})
		}
		for _, p := range shouldFailWithResourcesMustBeNonZeroErrors {
			t.Run(fmt.Sprintf("%s should fail because resource limits and requests are set to zero", p), func(t *testing.T) {
				err = applyManifest(p, true)
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), "'cpu' resource limit must be a nonzero value")
					assert.Contains(t, err.Error(), "'memory' resource limit must be a nonzero value")
					assert.Contains(t, err.Error(), "'cpu' resource request must be a nonzero value")
					assert.Contains(t, err.Error(), "'memory' resource request must be a nonzero value")
				}
			})
		}
		for _, p := range shouldSucceed {
			t.Run(fmt.Sprintf("%s should succeed", p), func(t *testing.T) {
				err = applyManifest(p, true)
				assert.Nil(t, err)
			})
		}
	}
}

func TestUpdate(t *testing.T) {
	err := applyManifest("test/manifests/namespace.yaml", false)
	if assert.Nil(t, err) {
		t.Run("should allow fixing pre-existing invalid resource spec", func(t *testing.T) {
			err = deleteManifest("test/webhook.yaml")
			if assert.Nil(t, err, "Could not disable webhook") {
				time.Sleep(5 * time.Second)
				err = applyManifest("test/manifests/invalid-deployment-update-01-zero.yaml", false)
				if assert.Nil(t, err, "Could not apply deployment-zero") {
					err = applyManifest("test/webhook.yaml", false)
					time.Sleep(5 * time.Second)
					if assert.Nil(t, err) {
						err = applyManifest("test/manifests/invalid-deployment-update-02-complete.yaml", false)
						assert.Nil(t, err)
					}
				}
			}
		})
	}
}
