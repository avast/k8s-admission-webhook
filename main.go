package main

import (
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func validate(ar v1beta1.AdmissionReview, config *config, clientSet *kubernetes.Clientset) *v1beta1.AdmissionResponse {
	validation := &objectValidation{ar.Request.Kind.Kind, nil, &validationViolationSet{}}
	deserializer := codecs.UniversalDeserializer()

	raw := ar.Request.Object.Raw
	var configMessage string
	switch ar.Request.Kind.Kind {
	case "Pod":
		configMessage = config.RuleResourceViolationMessage
		pod := corev1.Pod{}
		if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debugf("Admitting Pod: %+v", pod)
		validation.ObjMeta = &pod.ObjectMeta
		validatePodSpec(validation, &pod.ObjectMeta, &pod.Spec, config)

	case "ReplicaSet":
		configMessage = config.RuleResourceViolationMessage
		replicaSet := appsv1.ReplicaSet{}
		if _, _, err := deserializer.Decode(raw, nil, &replicaSet); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debugf("Admitting ReplicaSet: %+v", replicaSet)
		validation.ObjMeta = &replicaSet.ObjectMeta
		validatePodSpec(validation, &replicaSet.Spec.Template.ObjectMeta, &replicaSet.Spec.Template.Spec, config)

	case "Deployment":
		configMessage = config.RuleResourceViolationMessage
		deployment := appsv1.Deployment{}
		if _, _, err := deserializer.Decode(raw, nil, &deployment); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debugf("Admitting deployment: %+v", deployment)
		validation.ObjMeta = &deployment.ObjectMeta
		validatePodSpec(validation, &deployment.Spec.Template.ObjectMeta, &deployment.Spec.Template.Spec, config)

	case "DaemonSet":
		configMessage = config.RuleResourceViolationMessage
		daemonSet := appsv1.DaemonSet{}
		if _, _, err := deserializer.Decode(raw, nil, &daemonSet); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debugf("Admitting DaemonSet: %+v", daemonSet)
		validation.ObjMeta = &daemonSet.ObjectMeta
		validatePodSpec(validation, &daemonSet.Spec.Template.ObjectMeta, &daemonSet.Spec.Template.Spec, config)

	case "Job":
		configMessage = config.RuleResourceViolationMessage
		job := batchv1.Job{}
		if _, _, err := deserializer.Decode(raw, nil, &job); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debugf("Admitting Job: %+v", job)
		validation.ObjMeta = &job.ObjectMeta
		validatePodSpec(validation, &job.Spec.Template.ObjectMeta, &job.Spec.Template.Spec, config)

	case "CronJob":
		configMessage = config.RuleResourceViolationMessage
		cronJob := batchv1beta1.CronJob{}
		if _, _, err := deserializer.Decode(raw, nil, &cronJob); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debugf("Admitting CronJob: %+v", cronJob)
		validation.ObjMeta = &cronJob.ObjectMeta
		validatePodSpec(validation, &cronJob.Spec.JobTemplate.Spec.Template.ObjectMeta, &cronJob.Spec.JobTemplate.Spec.Template.Spec, config)

	case "Ingress":
		configMessage = config.RuleIngressViolationMessage
		ingress := extv1beta1.Ingress{}
		if _, _, err := deserializer.Decode(raw, nil, &ingress); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debugf("Admitting Ingress: %+v", ingress)
		validation.ObjMeta = &ingress.ObjectMeta
		err := ValidateIngress(validation, &ingress, config, clientSet)
		if err != nil {
			return toAdmissionResponse(err)
		}

	case "StatefulSet":
		configMessage = config.RuleResourceViolationMessage
		statefulSet := appsv1.StatefulSet{}
		if _, _, err := deserializer.Decode(raw, nil, &statefulSet); err != nil {
			log.Error(err)
			return toAdmissionResponse(err)
		}

		log.Debugf("Admitting stateful set: %+v", statefulSet)
		validation.ObjMeta = &statefulSet.ObjectMeta
		validatePodSpec(validation, &statefulSet.Spec.Template.ObjectMeta, &statefulSet.Spec.Template.Spec, config)

	default:
		log.Warnf("Admitted an unexpected resource: %v", ar.Request.Kind)
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

func main() {
	initLogger()

	//parse command line arguments
	config, configErr := initialize()
	if configErr != nil {
		log.Fatal(configErr)
	}
	log.Debugf("Configuration is: %+v", config)

	//initialize kube client
	kubeClientSet, kubeClientSetErr := KubeClientSet(true)
	if kubeClientSetErr != nil {
		log.Fatal(kubeClientSetErr)
	}

	http.HandleFunc("/validate", admitFunc(validate).serve(config, kubeClientSet))

	addr := fmt.Sprintf(":%v", config.ListenPort)
	var httpErr error
	if config.NoTLS {
		log.Infof("Starting webserver at %v (no TLS)", addr)
		httpErr = http.ListenAndServe(addr, nil)
	} else {
		log.Infof("Starting webserver at %v (TLS)", addr)
		httpErr = http.ListenAndServeTLS(addr, config.TLSCertFile, config.TLSPrivateKeyFile, nil)
	}

	if httpErr != nil {
		log.Fatal(httpErr)
	} else {
		log.Info("Finished")
	}

}

func initLogger() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}
