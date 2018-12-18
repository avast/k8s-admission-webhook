package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/client-go/kubernetes"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type admitFunc func(v1beta1.AdmissionReview, *config, *kubernetes.Clientset) *v1beta1.AdmissionResponse

func serve(w http.ResponseWriter, r *http.Request, admit admitFunc, config *config, clientSet *kubernetes.Clientset) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	//	glog.V(2).Info(fmt.Sprintf("handling request: %v", body))
	var reviewResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		log.Error(err)
		reviewResponse = toAdmissionResponse(err)
	}

	response := v1beta1.AdmissionReview{}

	if ar.Request != nil {
		reviewResponse = admit(ar, config, clientSet)
		log.Infof("sending response: %v", reviewResponse)

		if reviewResponse != nil {
			response.Response = reviewResponse
			response.Response.UID = ar.Request.UID
		}
		// reset the Object and OldObject, they are not needed in a response.
		ar.Request.Object = runtime.RawExtension{}
		ar.Request.OldObject = runtime.RawExtension{}
	} else {
		response.Response = toAdmissionResponse(fmt.Errorf("Invalid admission request"))
	}

	resp, err := json.Marshal(response)
	if err != nil {
		log.Error(err)
	}
	if _, err := w.Write(resp); err != nil {
		log.Error(err)
	}
}

func toAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func (fn admitFunc) serve(config *config, clientSet *kubernetes.Clientset) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		serve(w, r, fn, config, clientSet)
	}
}
