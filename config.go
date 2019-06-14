package main

import (
	"github.com/spf13/cobra"
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
	RuleSecurityReadonlyRootFilesystemRequired                  bool   `mapstructure:"rule-security-readonly-rootfs-required"`
	RuleSecurityReadonlyRootFilesystemRequiredWhitelistEnabled  bool   `mapstructure:"rule-security-readonly-rootfs-required-whitelist-enabled"`
	RuleIngressCollision                                        bool   `mapstructure:"rule-ingress-collision"`
	RuleIngressViolationMessage                                 string `mapstructure:"rule-ingress-violation-message"`
	AnnotationsPrefix                                           string `mapstructure:"annotations-prefix"`
	Namespace                                                   string `mapstructure:"namespace"`
}

func initCommonFlags(cmd *cobra.Command) {
	//pod
	cmd.Flags().String("rule-resource-violation-message", "",
		"Additional message to be included whenever any of the resource-related rules are violated.")
	cmd.Flags().Bool("rule-resource-limit-cpu-required", false,
		"Whether 'cpu' limit in resource specifications is required.")
	cmd.Flags().Bool("rule-resource-limit-cpu-must-be-nonzero", false,
		"Whether 'cpu' limit in resource specifications must be a nonzero value.")
	cmd.Flags().Bool("rule-resource-limit-memory-required", false,
		"Whether 'memory' limit in resource specifications is required.")
	cmd.Flags().Bool("rule-resource-limit-memory-must-be-nonzero", false,
		"Whether 'memory' limit in resource specifications must be a nonzero value.")
	cmd.Flags().Bool("rule-resource-request-cpu-required", false,
		"Whether 'cpu' request in resource specifications is required.")
	cmd.Flags().Bool("rule-resource-request-cpu-must-be-nonzero", false,
		"Whether 'cpu' request in resource specifications must be a nonzero value.")
	cmd.Flags().Bool("rule-resource-request-memory-required", false,
		"Whether 'memory' request in resource specifications is required.")
	cmd.Flags().Bool("rule-resource-request-memory-must-be-nonzero", false,
		"Whether 'memory' request in resource specifications must be a nonzero value.")
	cmd.Flags().Bool("rule-security-readonly-rootfs-required", false,
		"Whether 'readOnlyRootFilesystem' in security context specifications is required.")
	cmd.Flags().Bool("rule-security-readonly-rootfs-required-whitelist-enabled", false,
		"Whether rule 'readOnlyRootFilesystem' in security context can be ignored by container whitelisting.")

	//ingress
	cmd.Flags().String("rule-ingress-violation-message", "",
		"Additional message to be included whenever any of the ingress-related rules are violated.")
	cmd.Flags().Bool("rule-ingress-collision", false,
		"Whether ingress tls and host collision should be checked")

	//customizations
	cmd.Flags().String("annotations-prefix", "admission.validation.avast.com",
		"What prefix should be used for admission validation annotations.")
}
