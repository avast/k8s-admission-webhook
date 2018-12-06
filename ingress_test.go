package main

import (
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var validPaths = []PathDefinition{
	{"service1.avast.com", "/", "service1", "80"},
	{"service2.avast.com", "/", "service2", "80"},
}

var targetDescription = "test"

func TestIngress(t *testing.T) {
	initLogger()
	initKubeClientSet()

	t.Run("should pass path validation", func(t *testing.T) {
		validation := &objectValidation{"Ingress", &metav1.ObjectMeta{}, &validationViolationSet{}}
		validatePathDataRegex(validPaths, validation, targetDescription)
		assert.Empty(t, validation.Violations)
	})

	t.Run("E2E test fail", func(t *testing.T) {

	})
	//ingress()
}

//func ingress() {
//	content, err := ioutil.ReadFile("test/manifests/ingress-collision-tls.yaml") // just pass the file name
//	if err != nil {
//		panic(err)
//	}
//
//	ingress := &v1beta1.Ingress{}
//	deserializer := codecs.UniversalDeserializer()
//	if _, _, err := deserializer.Decode(content, nil, ingress); err != nil {
//		panic(err)
//	}
//
//	validation := &objectValidation{ingress.Kind, nil, &validationViolationSet{}}
//	config := &config{}
//	err := validateIngress(validation, ingress, config)
//
//	logger.Info(validation)
//
//}
