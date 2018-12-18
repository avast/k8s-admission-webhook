package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultName = "defaultIngressName"
var defaultNamespace = "defaultIngressNamespace"
var defaultSecret = "defaultSecret"
var collisionName = "collisionName"

var defaultPaths = []PathDefinition{
	{"service1.avast.com", "/", "service1", "80", defaultName, defaultNamespace},
	{"service2.avast.com", "/app", "service2", "80", defaultName, defaultNamespace},
	{"service3.avast.com", "/service3/", "service3", "80", defaultName, defaultNamespace},
	{"service4.avast.com", "/", "service4", "80", defaultName, defaultNamespace},
	{"service5.avast-stage.avast.com", "/", "service5", "80", defaultName, defaultNamespace},
}

var invalidHosts = []PathDefinition{
	{"service..avast.com", "/", "service1", "80", defaultName, defaultNamespace},
	{"service*.avast.com", "/", "service2", "80", defaultName, defaultNamespace},
	{"service${}.avast.com", "/", "service3", "80", defaultName, defaultNamespace},
}

var invalidPaths = []PathDefinition{
	{"service1.avast.com", "/*.**", "service1", "80", defaultName, defaultNamespace},
	{"service2.avast.com", "/${}", "service2", "80", defaultName, defaultNamespace},
	{"service3.avast.com", "/>", "service3", "80", defaultName, defaultNamespace},
}

var updatePaths = []PathDefinition{
	//changing port to 8080
	{"service1.avast.com", "/", "service1", "8080", defaultName, defaultNamespace},
}

var collisionPaths = []PathDefinition{
	//same definition as in defaultPaths, but name differs
	{"service1.avast.com", "/", "service2", "80", collisionName, defaultNamespace},
	//changed service
	{"service2.avast.com", "/app", "service3", "80", collisionName, defaultNamespace},
	//changed port
	{"service3.avast.com", "/service3/", "service3", "8080", collisionName, defaultNamespace},
}

var defaultTls = []TlsDefinition{
	{"service1.avast.com", defaultSecret, defaultName, defaultNamespace},
	{"service2.avast.com", defaultSecret, defaultName, defaultNamespace},
}

var collisionTls = []TlsDefinition{
	//changed secret
	{"service1.avast.com", "notDefaultSecret", defaultName, defaultNamespace},
	//changed ingress name
	{"service2.avast.com", "notDefaultSecret", collisionName, defaultNamespace},
	//changed ingress namespace
	{"service2.avast.com", defaultSecret, defaultName, collisionName},
}

var targetDescription = "test"

func TestIngress(t *testing.T) {
	initLogger()
	t.Run("Path Validation	", func(t *testing.T) {
		t.Run("Regex", func(t *testing.T) {
			t.Run("should pass path regex validation", func(t *testing.T) {
				validation := &objectValidation{"Ingress", &metav1.ObjectMeta{}, &validationViolationSet{}}
				ValidatePathDataRegex(defaultPaths, validation, targetDescription)
				assert.Empty(t, validation.Violations)
			})

			t.Run("should not pass path regex validation - invalid path", func(t *testing.T) {
				validation := &objectValidation{"Ingress", &metav1.ObjectMeta{}, &validationViolationSet{}}
				ValidatePathDataRegex(invalidPaths, validation, targetDescription)
				assert.Len(t, validation.Violations.Violations, 3)
			})

			t.Run("should not pass path regex validation - invalid host", func(t *testing.T) {
				validation := &objectValidation{"Ingress", &metav1.ObjectMeta{}, &validationViolationSet{}}
				ValidatePathDataRegex(invalidHosts, validation, targetDescription)
				assert.Len(t, validation.Violations.Violations, 3)
			})
		})

		t.Run("Collision", func(t *testing.T) {
			t.Run("should not pass path collision validation - collision paths", func(t *testing.T) {
				validation := &objectValidation{"Ingress", &metav1.ObjectMeta{}, &validationViolationSet{}}
				ValidatePathDataCollision(collisionPaths, defaultPaths, validation, targetDescription)
				assert.Len(t, validation.Violations.Violations, 3)
			})
			t.Run("should pass path collision validation - twice defaultPaths", func(t *testing.T) {
				validation := &objectValidation{"Ingress", &metav1.ObjectMeta{}, &validationViolationSet{}}
				ValidatePathDataCollision(defaultPaths, defaultPaths, validation, targetDescription)
				assert.Len(t, validation.Violations.Violations, 0)
			})
			t.Run("update should pass", func(t *testing.T) {
				validation := &objectValidation{"Ingress", &metav1.ObjectMeta{}, &validationViolationSet{}}
				ValidatePathDataCollision(updatePaths, defaultPaths, validation, targetDescription)
				assert.Len(t, validation.Violations.Violations, 0)
			})
		})
	})

	t.Run("TLS Validation", func(t *testing.T) {
		t.Run("Regex", func(t *testing.T) {

		})
		t.Run("Collision", func(t *testing.T) {
			t.Run("should not pass tls collision validation - collision tls", func(t *testing.T) {
				validation := &objectValidation{"Ingress", &metav1.ObjectMeta{}, &validationViolationSet{}}
				ValidateTlsDataCollision(collisionTls, defaultTls, validation, targetDescription)
				assert.Len(t, validation.Violations.Violations, 3)
			})

			t.Run("should pass tls collision validation - twice default tls", func(t *testing.T) {
				validation := &objectValidation{"Ingress", &metav1.ObjectMeta{}, &validationViolationSet{}}
				ValidateTlsDataCollision(defaultTls, defaultTls, validation, targetDescription)
				assert.Len(t, validation.Violations.Violations, 0)
			})
		})
	})
}
