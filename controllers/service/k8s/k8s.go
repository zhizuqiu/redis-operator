package k8s

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Service is the K8s service entrypoint.
type Services interface {
	All
	CRD
	Pod
	Secret
	Deployment
	ConfigMap
	StatefulSet
	Service
	Pv
	Pvc
}

type services struct {
	All
	CRD
	Pod
	Secret
	Deployment
	ConfigMap
	StatefulSet
	Service
	Pv
	Pvc
}

func New(kubeClient client.Client, log logr.Logger, scheme *runtime.Scheme) Services {
	return &services{
		All:         NewAllService(kubeClient, log),
		CRD:         NewCRDService(kubeClient, log),
		Pod:         NewPodService(kubeClient, log, scheme),
		Secret:      NewSecretService(kubeClient, log),
		Deployment:  NewDeploymentService(kubeClient, log, scheme),
		ConfigMap:   NewConfigMapService(kubeClient, log, scheme),
		StatefulSet: NewStatefulSetService(kubeClient, log, scheme),
		Service:     NewServiceService(kubeClient, log, scheme),
		Pv:          NewPvService(kubeClient, log),
		Pvc:         NewPvcService(kubeClient, log),
	}
}
