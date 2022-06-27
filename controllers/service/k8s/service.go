package k8s

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Service interface {
	GetService(namespace, name string) (*v1.Service, error)
}

type ServiceService struct {
	KubeClient client.Client
	Scheme     *runtime.Scheme
	Log        logr.Logger
}

func NewServiceService(kubeClient client.Client, log logr.Logger, scheme *runtime.Scheme) *ServiceService {
	log = log.WithValues("service", "k8s.ServiceService")
	return &ServiceService{
		KubeClient: kubeClient,
		Scheme:     scheme,
		Log:        log,
	}
}

func (s ServiceService) GetService(namespace, name string) (*v1.Service, error) {
	var service = &v1.Service{}
	if err := s.KubeClient.Get(context.Background(),
		types.NamespacedName{Namespace: namespace, Name: name},
		service,
	); err != nil {
		return nil, err
	}

	return service, nil
}
