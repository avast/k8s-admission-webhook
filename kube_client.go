package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
	"path/filepath"
)

var clientset *kubernetes.Clientset

func initKubeClientSet() {
	clientset = clients()
}

func ingressClient() (ingresses v1beta1.IngressInterface) {
	ingresses = clientset.ExtensionsV1beta1().Ingresses(metav1.NamespaceAll)
	return

}

func clients() *kubernetes.Clientset {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	config.WrapTransport = httpWrapper
	logger.Infof("Config: %+v\n", config)
	if err != nil {
		panic(err.Error())
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func httpWrapper(rt http.RoundTripper) http.RoundTripper {
	roundTripper := RT{rt}
	return roundTripper
}

type RT struct {
	original http.RoundTripper
}

func (roundTripper RT) RoundTrip(req *http.Request) (*http.Response, error) {
	logger.Debugf("%+v\n", req)
	response, err := roundTripper.original.RoundTrip(req)
	logger.Debugf("%+v\n", response)
	if err != nil {
		return response, err
	} else {
		body, bodyErr := ioutil.ReadAll(response.Body)
		if bodyErr != nil {
			return response, err
		}
		bodyCloseErr := response.Body.Close()
		if bodyCloseErr != nil {
			return response, err
		}
		logger.Debugf(string(body))
		response.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}

	return response, err
}
