package config

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/klog"
)

func GetGithubToken() (string, error) {

	token := getGithubTokenFromEnv()
	if token != "" {
		return token, nil
	}

	token, err := getGithubTokenFromK8sSecret()
	if err != nil {
		return "", err
	}

	klog.Info("github token obtained from k8s secret")
	return token, nil
}

func getGithubTokenFromEnv() string {
	return os.Getenv("GITHUB_TOKEN")
}

func getGithubTokenFromK8sSecret() (string, error) {

	secret, err := getK8sSecret()

	if err != nil {
		return "", err
	}
	const githubTokenEntry = "githubToken"
	val, ok := secret.Data[githubTokenEntry]
	if !ok {
		err := fmt.Errorf("secret %s does not contain file %s", secretName, githubTokenEntry)
		klog.Error(err.Error())
		return "", err
	}
	return strings.TrimRight(string(val), "\n"), nil
}
