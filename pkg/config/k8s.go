package config

import (
	"io/ioutil"

	v1meta "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

func getK8sClientSet() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

const secretName = "pr-brancher-secrets"

func getK8sSecret() (*v1meta.Secret, error) {
	clientSet, err := getK8sClientSet()
	if err != nil {
		return nil, err
	}
	namespace, err := getMyNamespace()
	if err != nil {
		return nil, err
	}

	secret, err := clientSet.CoreV1().Secrets(namespace).Get(secretName, v1.GetOptions{})
	if err != nil {
		klog.Error("Error looking up for %s secret in namespace %s : %s", secretName, namespace, err)
		return nil, err
	}

	return secret, nil
}

func updateK8sSecret(secret *v1meta.Secret) error {
	clientSet, err := getK8sClientSet()
	if err != nil {
		return err
	}
	namespace, err := getMyNamespace()
	if err != nil {
		return err
	}
	_, err = clientSet.CoreV1().Secrets(namespace).Update(secret)
	return err
}

func getMyNamespace() (string, error) {
	bytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
