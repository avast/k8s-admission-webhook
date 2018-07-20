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

func TestManifests(t *testing.T) {
	err := applyManifest("test/manifests/namespace.yaml", false)
	if assert.Nil(t, err) {
		t.Run("A deployment without resource limits and requests fails", func(t *testing.T) {
			err = applyManifest("test/manifests/sleep-deployment-incomplete.yaml", true)
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), "'cpu' resource limit must be specified")
				assert.Contains(t, err.Error(), "'memory' resource limit must be specified")
				assert.Contains(t, err.Error(), "'cpu' resource request must be specified")
				assert.Contains(t, err.Error(), "'memory' resource request must be specified")
			}
		})
		t.Run("A deployment with resource limits and requests succeeds", func(t *testing.T) {
			err = applyManifest("test/manifests/sleep-deployment-complete.yaml", true)
			assert.Nil(t, err)
		})
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
