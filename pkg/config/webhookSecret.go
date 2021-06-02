package config

import (
	"os"

	"github.com/sethvargo/go-password/password"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

const WebhookSecretEnvVar = "WEBHOOK_SECRET"

func GetWebhookSecret() (string, error) {

	if secret := getWebhookSecretFromEnv(); secret != "" {
		klog.Infof("Webhook secret retrieved from env var %s", WebhookSecretEnvVar)
		return secret, nil
	}

	secret, err := getWebhookSecretFromK8sSecret()
	if err != nil {
		return "", err
	}

	klog.Infof("Webhook secret retrieved from k8s secret")
	return secret, nil
}

func getWebhookSecretFromEnv() string {

	return os.Getenv(WebhookSecretEnvVar)
}

const webhookSecretEntry = "webhookSecret"

func getWebhookSecretFromK8sSecret() (string, error) {

	secret, err := getK8sSecret()

	if err != nil {
		return "", err
	}

	val, ok := secret.Data[webhookSecretEntry]
	if !ok {
		klog.Warningf("secret %s does not contain file %s", secretName, webhookSecretEntry)
		pwd, err := createWebhookSecretInK8sSecret(secret)
		if err == nil {
			klog.Warningf("Webhook password stored successfully in secret %s", secretName)
			return pwd, nil
		} else {
			klog.Fatalf("An error happened while trying to store webhook password in secret %s", secretName)
			return "", nil
		}
	}
	return string(val), nil
}

func createWebhookSecretInK8sSecret(secret *v1.Secret) (string, error) {

	pwd, err := password.Generate(64, 10, 10, false, true)
	if err != nil {
		klog.Errorf("Something happened while trying to generate a webhook password: %s", err)
		return "", err
	}
	secret.Data[webhookSecretEntry] = []byte(pwd)

	klog.Infof("Webhook secret created: %s", pwd)

	return pwd, updateK8sSecret(secret)
}
