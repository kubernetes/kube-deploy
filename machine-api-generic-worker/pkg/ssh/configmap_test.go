package ssh

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"k8s.io/client-go/kubernetes"
)

func generateByteSlice(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}

func generateValidPEM() []byte {
	b := &bytes.Buffer{}
	pk, _ := NewPrivateKey()

	_ = pem.Encode(b, &pem.Block{Type: rsaPrivateKey, Bytes: x509.MarshalPKCS1PrivateKey(pk)})
	return b.Bytes()
}

func generateSecretWithCustomNameAndIndex(name, index string, b []byte) runtime.Object {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string][]byte{
			index: b,
		},
	}
}

func generateSecretWithCustomName(name string, b []byte) runtime.Object {
	return generateSecretWithCustomNameAndIndex(name, privateKeyDataIndex, b)
}

func generateSecretWithKey(b []byte) runtime.Object {
	return generateSecretWithCustomName(secretName, b)
}

func fakeClientFactory(objs ...runtime.Object) kubernetes.Interface {
	return fake.NewSimpleClientset(objs...)
}

func TestEnsureSSHKeypairSecret(t *testing.T) {
	type args struct {
		client kubernetes.Interface
	}
	tests := []struct {
		name    string
		args    args
		want    *rsa.PrivateKey
		wantErr bool
	}{
		{
			name: "nil client",
			args: args{
				client: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "fake basis client without a key",
			args: args{
				client: fakeClientFactory(generateSecretWithKey(nil)),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "fake basis client with a malformed key",
			args: args{
				client: fakeClientFactory(generateSecretWithKey(generateByteSlice(2048))),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "fake basis client with a valid key",
			args: args{
				client: fakeClientFactory(generateSecretWithKey(generateValidPEM())),
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "fake basis client with a valid key, but the wrong secrete name",
			args: args{
				client: fakeClientFactory(generateSecretWithCustomName("blah", generateValidPEM())),
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "fake basis client with a valid key, but the wrong index",
			args: args{
				client: fakeClientFactory(generateSecretWithCustomNameAndIndex(secretName, "blah", generateValidPEM())),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EnsureSSHKeypairSecret(tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureSSHKeypairSecret() error = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
		})
	}
}
