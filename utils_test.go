package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func applyManifest(name string, deleteFirst bool) error {
	if deleteFirst {
		err := deleteManifest(name)
		if err != nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}

	apply := exec.Command(kubectl(), "apply", "-f", name)
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

func deleteManifest(name string) error {
	delete := exec.Command(kubectl(), "delete", "-f", name)
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
	}
	fmt.Print(string(output))
	return nil
}

func kubectl() string {
	kubectl := os.Getenv("KUBECTL")
	if kubectl == "" {
		kubectl = "kubectl"
	}
	return kubectl
}
