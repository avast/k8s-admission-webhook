package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"

	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validate(ar v1beta1.AdmissionReview, config *config) *v1beta1.AdmissionResponse {
	validation := &objectValidation{ar.Request.Kind.Kind, nil, &validationViolationSet{}}
	deserializer := codecs.UniversalDeserializer()

	raw := ar.Request.Object.Raw

	switch ar.Request.Kind.Kind {
	case "Pod":
		pod := corev1.Pod{}
		if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debug("Admitting Pod: %+v", pod)
		validation.ObjMeta = &pod.ObjectMeta
		validatePodSpec(validation, &pod.Spec, config)

	case "ReplicaSet":
		replicaSet := appsv1.ReplicaSet{}
		if _, _, err := deserializer.Decode(raw, nil, &replicaSet); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debug("Admitting ReplicaSet: %+v", replicaSet)
		validation.ObjMeta = &replicaSet.ObjectMeta
		validatePodSpec(validation, &replicaSet.Spec.Template.Spec, config)

	case "Deployment":
		deployment := appsv1.Deployment{}
		if _, _, err := deserializer.Decode(raw, nil, &deployment); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debugf("Admitting deployment: %+v", deployment)
		validation.ObjMeta = &deployment.ObjectMeta
		validatePodSpec(validation, &deployment.Spec.Template.Spec, config)

	case "DaemonSet":
		daemonSet := appsv1.DaemonSet{}
		if _, _, err := deserializer.Decode(raw, nil, &daemonSet); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debug("Admitting DaemonSet: %+v", daemonSet)
		validation.ObjMeta = &daemonSet.ObjectMeta
		validatePodSpec(validation, &daemonSet.Spec.Template.Spec, config)

	case "Job":
		job := batchv1.Job{}
		if _, _, err := deserializer.Decode(raw, nil, &job); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debug("Admitting Job: %+v", job)
		validation.ObjMeta = &job.ObjectMeta
		validatePodSpec(validation, &job.Spec.Template.Spec, config)

	case "CronJob":
		cronJob := batchv1beta1.CronJob{}
		if _, _, err := deserializer.Decode(raw, nil, &cronJob); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debug("Admitting CronJob: %+v", cronJob)
		validation.ObjMeta = &cronJob.ObjectMeta
		validatePodSpec(validation, &cronJob.Spec.JobTemplate.Spec.Template.Spec, config)

	default:
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
