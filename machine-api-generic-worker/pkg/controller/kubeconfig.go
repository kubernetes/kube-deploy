package controller

import (
	"fmt"

	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	secretTypeBootstrapToken v1.SecretType = "bootstrap.kubernetes.io/token"
)

func (c *Controller) getClusterInfoKubeconfig() (*clientcmdapi.Config, error) {
	cm, err := c.kubeClient.CoreV1().ConfigMaps(metav1.NamespacePublic).Get("cluster-info", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	data, found := cm.Data["kubeconfig"]
	if !found {
		return nil, fmt.Errorf("no kubeconfig found in cluster-info configmap")
	}
	return clientcmd.Load([]byte(data))
}

func (c *Controller) createBootstrapToken(name string) (string, error) {
	tokenID := rand.String(6)
	tokenSecret := rand.String(16)

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("bootstrap-token-%s", tokenID),
		},
		Type: secretTypeBootstrapToken,
		StringData: map[string]string{
			"description":                    "bootstrap token for " + name,
			"token-id":                       tokenID,
			"token-secret":                   tokenSecret,
			"expiration":                     metav1.Now().Add(24 * time.Hour).Format(time.RFC3339),
			"usage-bootstrap-authentication": "true",
			"usage-bootstrap-signing":        "true",
		},
	}

	_, err := c.kubeClient.CoreV1().Secrets(metav1.NamespaceSystem).Create(&secret)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s", tokenID, tokenSecret), nil
}

func (c *Controller) createBootstrapKubeconfig(name string) (string, error) {
	token, err := c.createBootstrapToken(name)
	if err != nil {
		return "", err
	}

	infoKubeconfig, err := c.getClusterInfoKubeconfig()
	if err != nil {
		return "", err
	}

	outConfig := infoKubeconfig.DeepCopy()

	outConfig.AuthInfos = map[string]*clientcmdapi.AuthInfo{
		"": {
			Token: token,
		},
	}

	bytes, err := clientcmd.Write(*outConfig)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
