package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func GetSSHKey() (ssh.Signer, error) {

	bytes, err := getSSHKeyBytes()

	if err != nil {
		return nil, err
	}

	return ssh.ParsePrivateKey(bytes)
}

func getSSHKeyBytes() ([]byte, error) {

	bytes, err := getSSHKeyFromEnv()
	if err != nil {
		return nil, err
	}

	if bytes != nil {
		klog.Info("SSH private key obtained from SSH_PK env")
		return bytes, nil
	}

	bytes, err = getSSHKeyFromK8sSecret()
	if err != nil {
		return nil, err
	}

	klog.Info("SSH private key obtained from k8s secret")
	return bytes, nil
}

func getSSHKeyFromEnv() ([]byte, error) {
	pkPath := os.Getenv("SSH_PK")
	if pkPath == "" {
		return nil, nil
	}

	bytes, err := ioutil.ReadFile(pkPath)
	return bytes, err
}

func getSSHKeyFromK8sSecret() ([]byte, error) {
	clientset, err := getK8sClientSet()
	if err != nil {
		return nil, err
	}
	namespace, err := getMyNamespace()
	if err != nil {
		return nil, err
	}

	const secretName = "pr-brancher-secrets"
	secret, err := clientset.CoreV1().Secrets(namespace).Get(secretName, v1.GetOptions{})
	if err != nil {
		klog.Error("Error looking up for %s secret in namespace %s : %s", secretName, namespace, err)
		return nil, err
	}

	const sshPkEntry = "ssh_pk"
	val, ok := secret.Data[sshPkEntry]
	if !ok {
		err := fmt.Errorf("secret %s does not contain file %s", secretName, sshPkEntry)
		klog.Error(err.Error())
		return nil, err
	}
	return val, nil
}

func getMyNamespace() (string, error) {
	bytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
