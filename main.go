package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"

	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validate(ar v1beta1.AdmissionReview, config *config) *v1beta1.AdmissionResponse {
	validation := &objectValidation{ar.Request.Kind.Kind, nil, &validationViolationSet{}}
	deserializer := codecs.UniversalDeserializer()

	if ar.Request.Kind.Kind == "Deployment" {
		deployment := appsv1.Deployment{}

		raw := ar.Request.Object.Raw
		if _, _, err := deserializer.Decode(raw, nil, &deployment); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debugf("Admitting deployment: %+v", deployment)
		validation.ObjMeta = &deployment.ObjectMeta

		for _, container := range deployment.Spec.Template.Spec.Containers {
			validateContainerResources(validation, fmt.Sprintf("Container %s", container.Name), &container, config)
		}

		for _, container := range deployment.Spec.Template.Spec.InitContainers {
			validateContainerResources(validation, fmt.Sprintf("Init container %s", container.Name), &container, config)
		}
	} else {
		log.Warnf("Admitted an unexpected resource: %v", ar.Request.Kind)
	}

	reviewResponse := v1beta1.AdmissionResponse{}
	message := validation.message(config)
	if len(message) > 0 {
		reviewResponse.Allowed = false
		reviewResponse.Result = &metav1.Status{Message: message}
	} else {
		reviewResponse.Allowed = true
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
