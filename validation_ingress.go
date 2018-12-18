package main

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var ingressHostRegExp = regexp.MustCompile(`^([a-zA-Z0-9_][a-zA-Z0-9_-]{0,62})(\.[a-zA-Z0-9_][a-zA-Z0-9_-]{0,62})*[._]?$`)
var ingressPathRegExp = regexp.MustCompile(`^[A-zA-Z0-9_/.\-]*$`)

type TlsDefinition struct {
	host             string
	secretName       string
	ingressName      string
	ingressNamespace string
}

type PathDefinition struct {
	host             string
	path             string
	serviceName      string
	servicePort      string
	ingressName      string
	ingressNamespace string
}

func (pathDefinition *PathDefinition) toUri() string {
	return pathDefinition.host + pathDefinition.path
}

func ValidateIngress(validation *objectValidation, ingress *extv1beta1.Ingress, config *config, clientSet *kubernetes.Clientset) error {

	if config.RuleIngressCollision {
		targetDesc := fmt.Sprintf("Ingress %s.%s: ", ingress.Name, ingress.Namespace)

		existingIngresses, err := IngressClientAllNamespaces(clientSet).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		log.Debugf("There are %d ingresses in the cluster to be compared", len(existingIngresses.Items))

		newIngressPathData := ingressPath(ingress)
		newIngressTlsData := ingressTls(ingress)

		ValidatePathDataRegex(newIngressPathData, validation, targetDesc)
		ValidateTlsDataRegex(newIngressTlsData, validation, targetDesc)

		var existingIngressesPathData []PathDefinition
		for _, existingIngress := range existingIngresses.Items {
			existingIngressesPathData = append(existingIngressesPathData, ingressPath(&existingIngress)...)
		}
		ValidatePathDataCollision(newIngressPathData, existingIngressesPathData, validation, targetDesc)

		var existingIngressesTlsData []TlsDefinition
		for _, existingIngress := range existingIngresses.Items {
			existingIngressesTlsData = append(existingIngressesTlsData, ingressTls(&existingIngress)...)
		}
		ValidateTlsDataCollision(newIngressTlsData, existingIngressesTlsData, validation, targetDesc)
	}
	return nil
}

func ValidateTlsDataRegex(tlsDefinition []TlsDefinition, validation *objectValidation, targetDesc string) {
	for _, tls := range tlsDefinition {
		validateHost(tls.host, validation, targetDesc)
	}
}

func ValidateTlsDataCollision(newIngressTlsData []TlsDefinition, existingIngressesTlsData []TlsDefinition, validation *objectValidation, targetDesc string) {
	for _, newTls := range newIngressTlsData {
		for _, existingTls := range existingIngressesTlsData {
			if newTls.host == existingTls.host {
				//if hosts are identical then also secret name and namespace has to match
				if !(newTls.ingressNamespace == existingTls.ingressNamespace && newTls.secretName == existingTls.secretName) {
					validation.Violations.add(
						validationViolation{
							targetDesc,
							fmt.Sprintf("TLS collision with '%s.%s' on '%s'", existingTls.ingressName, existingTls.ingressNamespace, existingTls.host),
						},
					)
				}
			}
		}
	}
}

func ValidatePathDataRegex(pathData []PathDefinition, validation *objectValidation, targetDesc string) {
	for _, path := range pathData {
		validatePath(path.path, validation, targetDesc)
		validateHost(path.host, validation, targetDesc)
	}
}

func ValidatePathDataCollision(newIngressPathData []PathDefinition, existingIngressPathData []PathDefinition, validation *objectValidation, targetDesc string) {
	for _, newIngressPath := range newIngressPathData {
		for _, existingIngressPath := range existingIngressPathData {
			// only other ingresses are considered - when updating it's not a collision
			if nameWithNamespace(newIngressPath) != nameWithNamespace(existingIngressPath) {
				if newIngressPath.toUri() == existingIngressPath.toUri() {
					if newIngressPath.serviceName != existingIngressPath.serviceName || newIngressPath.servicePort != existingIngressPath.servicePort {
						violation := validationViolation{
							targetDesc,
							fmt.Sprintf("Path collision with '%s' -> '%s:%s'", existingIngressPath.toUri(), existingIngressPath.serviceName, existingIngressPath.servicePort),
						}
						validation.Violations.add(violation)
					}
				}
			}
		}
	}
}

func nameWithNamespace(pathDefinition PathDefinition) string {
	return pathDefinition.ingressName + "." + pathDefinition.ingressNamespace
}

func ingressPath(ingress *extv1beta1.Ingress) (result []PathDefinition) {
	for _, rule := range ingress.Spec.Rules {
		host := rule.Host
		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {

				pathValue := path.Path
				if pathValue == "" {
					pathValue = "/"
				}

				pathDefinition := PathDefinition{
					host:             host,
					path:             pathValue,
					serviceName:      path.Backend.ServiceName,
					servicePort:      path.Backend.ServicePort.String(),
					ingressName:      ingress.Name,
					ingressNamespace: ingress.Namespace,
				}
				result = append(result, pathDefinition)
			}
		} else {
			log.Warnf("No http definition for %s.%s in rule %v", ingress.Name, ingress.Namespace, rule)
		}
	}
	return
}

func ingressTls(ingress *extv1beta1.Ingress) (result []TlsDefinition) {
	for _, tls := range ingress.Spec.TLS {
		for _, host := range tls.Hosts {
			tlsDefinition := TlsDefinition{host, tls.SecretName, ingress.Name, ingress.Namespace}
			result = append(result, tlsDefinition)
		}
	}
	return
}

func validateHost(host string, validation *objectValidation, targetDesc string) {
	if !ingressHostRegExp.MatchString(host) {
		validation.Violations.add(validationViolation{targetDesc, fmt.Sprintf("Host '%s' is not valid", host)})
	}
}

func validatePath(path string, validation *objectValidation, targetDesc string) {
	valid := strings.HasPrefix(path, "/")
	valid = valid && ingressPathRegExp.MatchString(path)
	if !valid {
		validation.Violations.add(validationViolation{targetDesc, fmt.Sprintf("Path '%s' is not valid", path)})

	}
}
