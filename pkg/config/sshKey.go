package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
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
	secret, err := getK8sSecret()
	if err != nil {
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
