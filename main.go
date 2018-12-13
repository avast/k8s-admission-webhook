package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"
)

func validate(ar v1beta1.AdmissionReview, config *config) *v1beta1.AdmissionResponse {
	validation := &objectValidation{ar.Request.Kind.Kind, nil, &validationViolationSet{}}
	deserializer := codecs.UniversalDeserializer()

	raw := ar.Request.Object.Raw
	var configMessage string
	switch ar.Request.Kind.Kind {
	case "Pod":
		configMessage = config.RuleResourceViolationMessage
		pod := corev1.Pod{}
		if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
			logger.Error(err)
			return toAdmissionResponse(err)
		}

		logger.Debugf("Admitting Pod: %+v", pod)
		validation.ObjMeta = &pod.ObjectMeta
		validatePodSpec(validation, &pod.Spec, config)

	case "ReplicaSet":
		configMessage = config.RuleResourceViolationMessage
		replicaSet := appsv1.ReplicaSet{}
		if _, _, err := deserializer.Decode(raw, nil, &replicaSet); err != nil {
			logger.Error(err)
			return toAdmissionResponse(err)
		}

		logger.Debugf("Admitting ReplicaSet: %+v", replicaSet)
		validation.ObjMeta = &replicaSet.ObjectMeta
		validatePodSpec(validation, &replicaSet.Spec.Template.Spec, config)

	case "Deployment":
		configMessage = config.RuleResourceViolationMessage
		deployment := appsv1.Deployment{}
		if _, _, err := deserializer.Decode(raw, nil, &deployment); err != nil {
			logger.Error(err)
			return toAdmissionResponse(err)
		}

		logger.Debugf("Admitting deployment: %+v", deployment)
		validation.ObjMeta = &deployment.ObjectMeta
		validatePodSpec(validation, &deployment.Spec.Template.Spec, config)

	case "DaemonSet":
		configMessage = config.RuleResourceViolationMessage
		daemonSet := appsv1.DaemonSet{}
		if _, _, err := deserializer.Decode(raw, nil, &daemonSet); err != nil {
			logger.Error(err)
			return toAdmissionResponse(err)
		}

		logger.Debugf("Admitting DaemonSet: %+v", daemonSet)
		validation.ObjMeta = &daemonSet.ObjectMeta
		validatePodSpec(validation, &daemonSet.Spec.Template.Spec, config)

	case "Job":
		configMessage = config.RuleResourceViolationMessage
		job := batchv1.Job{}
		if _, _, err := deserializer.Decode(raw, nil, &job); err != nil {
			logger.Error(err)
			return toAdmissionResponse(err)
		}

		logger.Debugf("Admitting Job: %+v", job)
		validation.ObjMeta = &job.ObjectMeta
		validatePodSpec(validation, &job.Spec.Template.Spec, config)

	case "CronJob":
		configMessage = config.RuleResourceViolationMessage
		cronJob := batchv1beta1.CronJob{}
		if _, _, err := deserializer.Decode(raw, nil, &cronJob); err != nil {
			logger.Error(err)
			return toAdmissionResponse(err)
		}

		logger.Debugf("Admitting CronJob: %+v", cronJob)
		validation.ObjMeta = &cronJob.ObjectMeta
		validatePodSpec(validation, &cronJob.Spec.JobTemplate.Spec.Template.Spec, config)

	case "Ingress":
		configMessage = config.RuleIngressViolationMessage
		ingress := extv1beta1.Ingress{}
		if _, _, err := deserializer.Decode(raw, nil, &ingress); err != nil {
			logger.Error(err)
			return toAdmissionResponse(err)
		}

		logger.Debugf("Admitting Ingress: %+v", ingress)
		validation.ObjMeta = &ingress.ObjectMeta
		err := ValidateIngress(validation, &ingress, config)
		if err != nil {
			return toAdmissionResponse(err)
		}

	default:
		logger.Warnf("Admitted an unexpected resource: %v", ar.Request.Kind)
	}

	reviewResponse := v1beta1.AdmissionResponse{}

	message := validation.message(configMessage)
	if len(message) > 0 {
		reviewResponse.Allowed = false
		reviewResponse.Result = &metav1.Status{Message: message}
	} else {
		reviewResponse.Allowed = true
	}

	return &reviewResponse
}

var logger *logrus.Logger

func main() {
	config := initialize()
	initLogger()
	InitKubeClientSet(true)

	// parse and validate arguments
	logger.Debugf("Configuration is: %+v", config)

	http.HandleFunc("/validate", admitFunc(validate).serve(&config))
	addr := fmt.Sprintf(":%v", config.ListenPort)

	var err error

	if config.NoTLS {
		logger.Infof("Starting webserver at %v (no TLS)", addr)
		err = http.ListenAndServe(addr, nil)
	} else {
		logger.Infof("Starting webserver at %v (TLS)", addr)
		err = http.ListenAndServeTLS(addr, config.TLSCertFile, config.TLSPrivateKeyFile, nil)
	}

	logger.Fatal(err)
}

func initLogger() {
	logger = logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)
}
