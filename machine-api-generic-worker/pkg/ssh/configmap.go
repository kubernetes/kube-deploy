package ssh

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	privateKeyDataIndex = "id_rsa"
	secretName          = "machine-controller-ssh-key"
	rsaPrivateKey       = "RSA PRIVATE KEY"
)

// EnsureSSHKeypairSecret
func EnsureSSHKeypairSecret(client kubernetes.Interface) (*rsa.PrivateKey, error) {
	if client == nil {
		return nil, fmt.Errorf("got an nil k8s client")
	}
	secret, err := client.CoreV1().Secrets(metav1.NamespaceSystem).Get(secretName, metav1.GetOptions{})
	if err == nil {
		return keyFromSecret(secret)
	}

	if !errors.IsNotFound(err) {
		return nil, err
	}

	glog.V(4).Info("generating master ssh keypair")
	pk, err := NewPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ssh keypair: %v", err)
	}

	privateKeyPEM := &pem.Block{Type: rsaPrivateKey, Bytes: x509.MarshalPKCS1PrivateKey(pk)}
	privBuf := bytes.Buffer{}
	err = pem.Encode(&privBuf, privateKeyPEM)
	if err != nil {
		return nil, err
	}

	secret = &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			privateKeyDataIndex: privBuf.Bytes(),
		},
	}

	_, err = client.CoreV1().Secrets(metav1.NamespaceSystem).Create(secret)
	if err != nil {
		return nil, err
	}
	return pk, nil

}

func keyFromSecret(secret *v1.Secret) (*rsa.PrivateKey, error) {
	b, exists := secret.Data[privateKeyDataIndex]
	if !exists {
		return nil, fmt.Errorf("key data not found in secret '%s/%s' (secret.data['%s']). remove it and a new one will be created", secret.Namespace, secret.Name, privateKeyDataIndex)
	}
	if len(b) == 0 {
		return nil, fmt.Errorf("key data not found in secret '%s/%s' (secret.data['%s']). remove it and a new one will be created", secret.Namespace, secret.Name, privateKeyDataIndex)
	}
	decoded, _ := pem.Decode(b)

	if decoded == nil {
		return nil, fmt.Errorf("invalid PEM in secret '%s/%s'. remove it and a new one will be created", secret.Namespace, secret.Name)
	}

	pk, err := x509.ParsePKCS1PrivateKey(decoded.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	return pk, nil
}
