// +build e2e

package main

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var shouldFailWithResourcesNotSpecifiedErrors = []string{
	"test/manifests/deployment-incomplete.yaml",
	"test/manifests/pod-incomplete.yaml",
	"test/manifests/job-incomplete.yaml",
	"test/manifests/cronjob-incomplete.yaml",
}
var shouldFailWithResourcesMustBeNonZeroErrors = []string{
	"test/manifests/deployment-zero.yaml",
	"test/manifests/pod-zero.yaml",
	"test/manifests/job-zero.yaml",
	"test/manifests/cronjob-zero.yaml",
}
var shouldSucceed = []string{
	"test/manifests/deployment-complete.yaml",
	"test/manifests/pod-complete.yaml",
	"test/manifests/job-complete.yaml",
	"test/manifests/cronjob-complete.yaml",
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

func applyManifest(name string, deleteFirst bool) error {
	kubectl := os.Getenv("KUBECTL")
	if kubectl == "" {
		kubectl = "kubectl"
	}
	if deleteFirst {
		delete := exec.Command(kubectl, "delete", "-f", name)
		output, err := delete.Output()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				stdErr := string(exitError.Stderr)
				if !strings.Contains(stdErr, "NotFound") {
					fmt.Printf("Unexpected error while deleting, stderr: %s", stdErr)
					return errors.New(stdErr)
				}
			} else {
				return err
			}
		} else {
			fmt.Print(string(output))
		}
	}

	apply := exec.Command(kubectl, "apply", "-f", name)
	output, err := apply.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			stdErr := string(exitError.Stderr)
			fmt.Printf("Non-zero exit code while applying, stderr: %s", stdErr)
			return errors.New(stdErr)
		} else {
			return err
		}
	}
	fmt.Print(string(output))
	return nil
}
