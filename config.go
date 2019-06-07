package main

import (
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type config struct {
	NoTLS                                                       bool   `mapstructure:"no-tls"`
	TLSCertFile                                                 string `mapstructure:"tls-cert-file"`
	TLSPrivateKeyFile                                           string `mapstructure:"tls-private-key-file"`
	ListenPort                                                  int    `mapstructure:"listen-port"`
	RuleResourceViolationMessage                                string `mapstructure:"rule-resource-violation-message"`
	RuleResourceLimitCPURequired                                bool   `mapstructure:"rule-resource-limit-cpu-required"`
	RuleResourceLimitCPUMustBeNonZero                           bool   `mapstructure:"rule-resource-limit-cpu-must-be-nonzero"`
	RuleResourceLimitMemoryRequired                             bool   `mapstructure:"rule-resource-limit-memory-required"`
	RuleResourceLimitMemoryMustBeNonZero                        bool   `mapstructure:"rule-resource-limit-memory-must-be-nonzero"`
	RuleResourceRequestCPURequired                              bool   `mapstructure:"rule-resource-request-cpu-required"`
	RuleResourceRequestCPUMustBeNonZero                         bool   `mapstructure:"rule-resource-request-cpu-must-be-nonzero"`
	RuleResourceRequestMemoryRequired                           bool   `mapstructure:"rule-resource-request-memory-required"`
	RuleResourceRequestMemoryMustBeNonZero                      bool   `mapstructure:"rule-resource-request-memory-must-be-nonzero"`
	RuleSecurityReadonlyRootFilesystemRequired                  bool   `mapstructure:"rule-security-readonly-root-filesystem-required"`
	RuleSecurityReadonlyRootFilesystemRequiredWhitelistEnabled  bool   `mapstructure:"rule-security-readonly-root-filesystem-required-whitelist-enabled"`
	RuleIngressCollision                                        bool   `mapstructure:"rule-ingress-collision"`
	RuleIngressViolationMessage                                 string `mapstructure:"rule-ingress-violation-message"`
	AdmissionValidationAnnotationsPrefix                        string `mapstructure:"admission-validation-annotations-prefix"`
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

	//pod
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
	rootCmd.Flags().Bool("rule-security-readonly-root-filesystem-required", false,
		"Whether 'readOnlyRootFilesystem' in security context specifications is required.")
	rootCmd.Flags().Bool("rule-security-readonly-root-filesystem-required-whitelist-enabled", false,
		"Whether rule 'readOnlyRootFilesystem' in security context can be overriden by container whitelisting.")

	//ingress
	rootCmd.Flags().String("rule-ingress-violation-message", "",
		"Additional message to be included whenever any of the ingress-related rules are violated.")
	rootCmd.Flags().Bool("rule-ingress-collision", false,
		"Whether ingress tls and host collision should be checked")

	//customizations
	rootCmd.Flags().String("admission-validation-annotations-prefix", "",
		"What prefix should be used for admission validation annotations.")

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
