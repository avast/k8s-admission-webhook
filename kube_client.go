package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

var clientset *kubernetes.Clientset

func InitKubeClientSet(inCluster bool) {
	clientset = clients(inCluster)
}

func IngressClientAllNamespaces() v1beta1.IngressInterface {
	return IngressClient(metav1.NamespaceAll)
}

func IngressClient(namespace string) (ingresses v1beta1.IngressInterface) {
	ingresses = clientset.ExtensionsV1beta1().Ingresses(namespace)
	return
}

func clients(inCluster bool) *kubernetes.Clientset {

	var config *rest.Config

	if inCluster {
		c, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		config = c
	} else {
		kubeconfig := filepath.Join(homeDir(), ".kube", "config")
		exists, existsErr := fileExists(kubeconfig)
		if existsErr != nil {
			panic(existsErr)
		}
		if exists {
			c, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				panic(err.Error())
			}
			config = c
		} else {
			panic(kubeconfig + " does not exist")
		}
	}

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

func fileExists(name string) (bool, error) {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}
