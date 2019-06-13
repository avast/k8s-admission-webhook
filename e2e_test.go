// +build e2e

package main

import (
	"fmt"
	"testing"
	"time"
	"os"
	"os/exec"

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
var shouldFailWithWritableRootFilesystemError = []string{
	"test/manifests/pod-readonly-rootfs-false.yaml",
	"test/manifests/pod-readonly-rootfs-missing.yaml",
	"test/manifests/pod-readonly-rootfs-annotation-missing.yaml",
	"test/manifests/pod-readonly-rootfs-annotation-false.yaml",
}
var shouldSucceed = []string{
	"test/manifests/deployment-complete.yaml",
	"test/manifests/pod-complete.yaml",
	"test/manifests/job-complete.yaml",
	"test/manifests/cronjob-complete.yaml",
	"test/manifests/statefulset-complete.yaml",
	"test/manifests/pod-readonly-rootfs-annotation-whitelist.yaml",
	"test/manifests/deployment-complete-annotation-whitelist.yaml",
	"test/manifests/cronjob-complete-annotation-whitelist.yaml",
	"test/manifests/job-complete-annotation-whitelist.yaml",
	"test/manifests/statefulset-complete-annotation-whitelist.yaml",
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
		for _, p := range shouldFailWithWritableRootFilesystemError {
			t.Run(fmt.Sprintf("%s should fail because security context is missing or root filesystem is not readonly", p), func(t *testing.T) {
				err = applyManifest(p, true)
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), "'securityContext' with 'readOnlyRootFilesystem: true' must be specified")
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

func TestAnnotationPrefix(t *testing.T) {
	t.Run("should succeed with default annotation prefix used", func(t *testing.T) {
		err := applyManifest("test/manifests/pod-custom-annot-prefix-default.yaml", true)
		assert.Nil(t, err)
	})
	t.Run("should succeed with custom string annotation prefix used", func(t *testing.T) {
		err := deleteManifest("test/webhook.yaml")
		if assert.Nil(t, err, "Could not disable webhook") {
			time.Sleep(5 * time.Second)
			
			file, _ := os.Create("test/webhook-prefix.yaml")
			sed := exec.Command("sed", "s/#ANNOTATION_PREFIX_NAME_PLACEHOLDER/- name: ANNOTATIONS_PREFIX/; s/#ANNOTATION_PREFIX_VALUE_PLACEHOLDER/value: \"custom.test.prefix\"/", "test/webhook.yaml")
			sed.Stdout = file
			sed.Start()

			err = applyManifest("test/webhook-prefix.yaml", false)
			time.Sleep(5 * time.Second)
			if assert.Nil(t, err, "Could not apply webhook") {
				err = applyManifest("test/manifests/pod-custom-annot-prefix-set.yaml", true)
				assert.Nil(t, err)
			}
		}
	})
	t.Run("should fail because different annotation prefix is used", func(t *testing.T) {
		err := applyManifest("test/manifests/pod-custom-annot-prefix-wrong.yaml", true)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "'securityContext' with 'readOnlyRootFilesystem: true' must be specified")
		}
	})

	// Clean up after test
	err := deleteManifest("test/webhook-prefix.yaml")
	if assert.Nil(t, err, "Could not disable webhook") {
		time.Sleep(5 * time.Second)
		err = applyManifest("test/webhook.yaml", false)
		assert.Nil(t, err)
		time.Sleep(5 * time.Second)
	}
}
