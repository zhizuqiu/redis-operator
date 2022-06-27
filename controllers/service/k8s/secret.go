package k8s

import (
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Secret interface {
	GetSecret(namespace, name string) (*corev1.Secret, error)
}

type SecretService struct {
	KubeClient client.Client
	Log        logr.Logger
}

func NewSecretService(kubeClient client.Client, log logr.Logger) *SecretService {
	log = log.WithValues("service", "k8s.SecretService")
	return &SecretService{
		KubeClient: kubeClient,
		Log:        log,
	}
}

func (s SecretService) GetSecret(namespace, name string) (*corev1.Secret, error) {
	var secret = &corev1.Secret{}
	if err := s.KubeClient.Get(context.Background(),
		types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		},
		secret,
	); err != nil {
		return nil, err
	}
	return secret, nil
}
