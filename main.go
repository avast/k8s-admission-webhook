package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"

	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// https://github.com/kubernetes/kubernetes/tree/release-1.9/test/images/webhook

func isResourceSet(resList corev1.ResourceList, name corev1.ResourceName) bool {
	var missing = resList == nil
	if !missing {
		if _, ok := resList[name]; !ok {
			missing = true
		}
	}
	return !missing
}

func validate(ar v1beta1.AdmissionReview, config *config) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}

	var validationMsg string
	reviewResponse.Allowed = true

	deserializer := codecs.UniversalDeserializer()

	if ar.Request.Kind.Kind == "Deployment" {
		deployment := appsv1.Deployment{}

		raw := ar.Request.Object.Raw
		if _, _, err := deserializer.Decode(raw, nil, &deployment); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debugf("Admitting deployment: %+v", deployment)

		var containerMsg string
		var resourcesValidationError = false
		checkResourceIsSet := func(resList corev1.ResourceList, name corev1.ResourceName, enabled bool, message string) {
			if enabled && !isResourceSet(resList, name) {
				resourcesValidationError = true
				containerMsg = containerMsg + message + " "
			}
		}

		for _, container := range deployment.Spec.Template.Spec.Containers {
			checkResourceIsSet(container.Resources.Limits, corev1.ResourceCPU,
				config.RuleResourceLimitCPURequired, config.RuleResourceLimitCPURequiredMessage)
			checkResourceIsSet(container.Resources.Limits, corev1.ResourceMemory,
				config.RuleResourceLimitMemoryRequired, config.RuleResourceLimitMemoryRequiredMessage)
			checkResourceIsSet(container.Resources.Requests, corev1.ResourceCPU,
				config.RuleResourceRequestCPURequired, config.RuleResourceRequestCPURequiredMessage)
			checkResourceIsSet(container.Resources.Requests, corev1.ResourceCPU,
				config.RuleResourceRequestMemoryRequired, config.RuleResourceRequestMemoryRequiredMessage)

			if resourcesValidationError {
				reviewResponse.Allowed = false
				validationMsg = fmt.Sprintf("%s Container '%s' validation errors: %s", validationMsg, container.Name, containerMsg)
			}
		}
	} else {
		log.Warnf("Admitted an unexpected resource: %v", ar.Request.Kind)
	}

	if !reviewResponse.Allowed {
		reviewResponse.Result = &metav1.Status{Message: strings.TrimSpace(validationMsg)}
	}
	return &reviewResponse
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	// parse and validate arguments
	config := initialize()
	log.Debugf("Configuration is: %+v", config)

	http.HandleFunc("/validate", admitFunc(validate).serve(&config))
	addr := fmt.Sprintf(":%v", config.ListenPort)

	var err error

	if config.NoTLS {
		log.Infof("Starting webserver at %v (no TLS)", addr)
		err = http.ListenAndServe(addr, nil)
	} else {
		log.Infof("Starting webserver at %v (TLS)", addr)
		err = http.ListenAndServeTLS(addr, config.TLSCertFile, config.TLSPrivateKeyFile, nil)
	}

	log.Fatal(err)
}
