package main

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

type config struct {
	NoTLS                                  bool   `mapstructure:"no-tls"`
	TLSCertFile                            string `mapstructure:"tls-cert-file"`
	TLSPrivateKeyFile                      string `mapstructure:"tls-private-key-file"`
	ListenPort                             int    `mapstructure:"listen-port"`
	RuleResourceViolationMessage           string `mapstructure:"rule-resource-violation-message"`
	RuleResourceLimitCPURequired           bool   `mapstructure:"rule-resource-limit-cpu-required"`
	RuleResourceLimitCPUMustBeNonZero      bool   `mapstructure:"rule-resource-limit-cpu-must-be-nonzero"`
	RuleResourceLimitMemoryRequired        bool   `mapstructure:"rule-resource-limit-memory-required"`
	RuleResourceLimitMemoryMustBeNonZero   bool   `mapstructure:"rule-resource-limit-memory-must-be-nonzero"`
	RuleResourceRequestCPURequired         bool   `mapstructure:"rule-resource-request-cpu-required"`
	RuleResourceRequestCPUMustBeNonZero    bool   `mapstructure:"rule-resource-request-cpu-must-be-nonzero"`
	RuleResourceRequestMemoryRequired      bool   `mapstructure:"rule-resource-request-memory-required"`
	RuleResourceRequestMemoryMustBeNonZero bool   `mapstructure:"rule-resource-request-memory-must-be-nonzero"`
	RuleIngressCollision                   bool   `mapstructure:"rule-ingress-collision"`
	RuleIngressViolationMessage            string `mapstructure:"rule-ingress-violation-message"`
}

var rootCmd = &cobra.Command{
	Use:  "k8s-admission-webhook",
	Long: "Kubernetes Admission Webhook",
	Run:  func(cmd *cobra.Command, args []string) {},
}

func initialize() (*config, error) {
	rootCmd.Flags().String("tls-cert-file", "",
		"Path to the certificate file. Required, unless --no-tls is set.")
	rootCmd.Flags().Bool("no-tls", false,
		"Do not use TLS.")
	rootCmd.Flags().String("tls-private-key-file", "",
		"Path to the certificate key file. Required, unless --no-tls is set.")
	rootCmd.Flags().Int32("listen-port", 443,
		"Port to listen on.")
	//resources
	rootCmd.Flags().String("rule-resource-violation-message", "",
		"Additional message to be included whenever any of the resource-related rules are violated.")
	rootCmd.Flags().Bool("rule-resource-limit-cpu-required", false,
		"Whether 'cpu' limit in resource specifications is required.")
	rootCmd.Flags().Bool("rule-resource-limit-cpu-must-be-nonzero", false,
		"Whether 'cpu' limit in resource specifications must be a nonzero value.")
	rootCmd.Flags().Bool("rule-resource-limit-memory-required", false,
		"Whether 'memory' limit in resource specifications is required.")
	rootCmd.Flags().Bool("rule-resource-limit-memory-must-be-nonzero", false,
		"Whether 'memory' limit in resource specifications must be a nonzero value.")
	rootCmd.Flags().Bool("rule-resource-request-cpu-required", false,
		"Whether 'cpu' request in resource specifications is required.")
	rootCmd.Flags().Bool("rule-resource-request-cpu-must-be-nonzero", false,
		"Whether 'cpu' request in resource specifications must be a nonzero value.")
	rootCmd.Flags().Bool("rule-resource-request-memory-required", false,
		"Whether 'memory' request in resource specifications is required.")
	rootCmd.Flags().Bool("rule-resource-request-memory-must-be-nonzero", false,
		"Whether 'memory' request in resource specifications must be a nonzero value.")

	//ingress
	rootCmd.Flags().String("rule-ingress-violation-message", "",
		"Additional message to be included whenever any of the ingress-related rules are violated.")
	rootCmd.Flags().Bool("rule-ingress-collision", false,
		"Whether ingress tls and host collision should be checked")

	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		return errorWithUsage(err)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	if err := rootCmd.Execute(); err != nil {
		return errorWithUsage(err)
	}

	c := &config{}
	if err := viper.Unmarshal(c); err != nil {
		return errorWithUsage(err)
	}

	if !c.NoTLS && (c.TLSPrivateKeyFile == "" || c.TLSCertFile == "") {
		return errorWithUsage(errors.New("Both --tls-cert-file and --tls-private-key-file are required (unless TLS is disabled by setting --no-tls)"))
	}

	return c, nil
}

func errorWithUsage(err error) (*config, error) {
	log.Error(rootCmd.UsageString())
	return nil, err
}
