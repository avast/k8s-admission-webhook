package main

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var scannerCmd = &cobra.Command {
	Use:  "scanner",
	Short:"Scans cluster (from current context) for objects vialoting rules",
	Long: "Scans cluster (from current context) for objects violating rules spciefied by flags",
	Run:  scanCluster,
}

var scannerViper = viper.New()

func init() {
	rootCmd.AddCommand(scannerCmd)

	scannerCmd.Flags().String("namespace", "",
		"Whether specific namespace should be scanned. If omitted, all namespaces are scanned.")

	initCommonFlags(scannerCmd)

	if err := scannerViper.BindPFlags(scannerCmd.Flags()); err != nil {
		errorWithUsage(err)
	}

	scannerViper.AutomaticEnv()
	scannerViper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}

func scanCluster(cmd *cobra.Command, args []string) {
	config := &config{}
	if err := scannerViper.Unmarshal(config); err != nil {
		errorWithUsage(err)
	}

	log.Debugf("Configuration is: %+v", config)

	kubeClientSet, err := KubeClientSet(false)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Debugf("Init finished!")
	
	validatePods(kubeClientSet, config)
	validateIngresses(kubeClientSet, config)	

	log.Debugf("Check completed!")
}

func validatePods(clientset *kubernetes.Clientset, config *config) {
	log.Debugf("Check Pods...")

	namespaceToScan := config.Namespace
	pods, err := clientset.CoreV1().Pods(namespaceToScan).List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
	}

	if namespaceToScan == "" {
		log.Debugf("There are %d pods in all namespaces", len(pods.Items))
	} else {
		log.Debugf("There are %d pods in the namespace '%s'", len(pods.Items), namespaceToScan)
	}

	for _, pod := range pods.Items {
		validation := &objectValidation{"Pod", nil, &validationViolationSet{}}
		validatePodSpec(validation, &pod.ObjectMeta, &pod.Spec, config)
		if len(validation.Violations.Violations) > 0 {
			log.Debugf("Pod from namespace '%s' with name '%s' has following violations:", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
			for _, v := range validation.Violations.Violations {
				log.Debugf("   %s", v.Message)
			}
		}
	}
}

func validateIngresses(clientset *kubernetes.Clientset, config *config) {
	log.Debugf("Check Ingresses...")

	namespaceToScan := config.Namespace
	ingresses, err := clientset.ExtensionsV1beta1().Ingresses(namespaceToScan).List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	if namespaceToScan == "" {
		log.Debugf("There are %d ingresses in all namespaces", len(ingresses.Items))
	} else {
		log.Debugf("There are %d ingresses in the namespace '%s'", len(ingresses.Items), namespaceToScan)
	}

	for _, ingress := range ingresses.Items {
		validation := &objectValidation{"Ingress", nil, &validationViolationSet{}}
		ValidateIngress(validation, &ingress, config, clientset)
		if len(validation.Violations.Violations) > 0 {
			log.Debugf("Ingress from namespace '%s' with name '%s' has following violations:", ingress.ObjectMeta.Namespace, ingress.ObjectMeta.Name)
			for _, v := range validation.Violations.Violations {
				log.Debugf("   %s", v.Message)
			}
		}
	}
}
