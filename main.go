package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command {
	Use:  "k8s-admission-webhook",
	Long: "Kubernetes Admission Webhook",
}

func main() {
	initLogger()

	if err := rootCmd.Execute(); err != nil {
		errorWithUsage(err)
	}
}

func initLogger() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func errorWithUsage(err error) {
	log.Error(rootCmd.UsageString())
	log.Fatal(err)
}
