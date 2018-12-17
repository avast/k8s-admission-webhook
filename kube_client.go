package main

import (
	"errors"
	"github.com/JaSei/pathutil-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func IngressClientAllNamespaces(clientset *kubernetes.Clientset) v1beta1.IngressInterface {
	return IngressClient(metav1.NamespaceAll, clientset)
}

func IngressClient(namespace string, clientset *kubernetes.Clientset) (ingresses v1beta1.IngressInterface) {
	ingresses = clientset.ExtensionsV1beta1().Ingresses(namespace)
	return
}

func KubeClientSet(inCluster bool) (*kubernetes.Clientset, error) {

	var config *rest.Config

	if inCluster {
		c, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		config = c
	} else {
		kubeconfig, err := pathutil.Home(".kube", "config")
		if err != nil {
			return nil, err
		}

		if kubeconfig.IsFile() {
			c, err := clientcmd.BuildConfigFromFlags("", kubeconfig.String())
			if err != nil {
				return nil, err
			}
			config = c
		} else {
			return nil, errors.New(kubeconfig.String() + " does not exist")
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
