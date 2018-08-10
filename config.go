package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type config struct {
	NoTLS                             bool   `mapstructure:"no-tls"`
	TLSCertFile                       string `mapstructure:"tls-cert-file"`
	TLSPrivateKeyFile                 string `mapstructure:"tls-private-key-file"`
	ListenPort                        int    `mapstructure:"listen-port"`
	RuleResourceViolationMessage      string `mapstructure:"rule-resource-violation-message"`
	RuleResourceLimitCPURequired      bool   `mapstructure:"rule-resource-limit-cpu-required"`
	RuleResourceLimitMemoryRequired   bool   `mapstructure:"rule-resource-limit-memory-required"`
	RuleResourceRequestCPURequired    bool   `mapstructure:"rule-resource-request-cpu-required"`
	RuleResourceRequestMemoryRequired bool   `mapstructure:"rule-resource-request-memory-required"`
}

var rootCmd = &cobra.Command{
	Use:  "k8s-admission-webhook",
	Long: "Kubernetes Admission Webhook",
	Run:  func(cmd *cobra.Command, args []string) {},
}

func initialize() config {
	rootCmd.Flags().String("tls-cert-file", "",
		"Path to the certificate file. Required, unless --no-tls is set.")
	rootCmd.Flags().Bool("no-tls", false,
		"Do not use TLS.")
	rootCmd.Flags().String("tls-private-key-file", "",
		"Path to the certificate key file. Required, unless --no-tls is set.")
	rootCmd.Flags().Int32("listen-port", 443,
		"Port to listen on.")
	rootCmd.Flags().String("rule-resource-violation-message", "",
		"Additional message to be included whenever any of the resource-related rules are violated.")
	rootCmd.Flags().Bool("rule-resource-limit-cpu-required", false,
		"Whether 'cpu' limit is required in resource specifications.")
	rootCmd.Flags().Bool("rule-resource-limit-memory-required", false,
		"Whether 'memory' limit is required in resource specifications.")
	rootCmd.Flags().Bool("rule-resource-request-cpu-required", false,
		"Whether 'cpu' request is required in resource specifications.")
	rootCmd.Flags().Bool("rule-resource-request-memory-required", false,
		"Whether 'memory' request is required in resource specifications.")

	viper.BindPFlags(rootCmd.Flags())
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	c := config{}
	viper.Unmarshal(&c)
	if !c.NoTLS && (c.TLSPrivateKeyFile == "" || c.TLSCertFile == "") {
		fmt.Println("Both --tls-cert-file and --tls-private-key-file are required (unless TLS is disabled by setting --no-tls)")
		rootCmd.Usage()
		os.Exit(1)
	}

	return c
}
