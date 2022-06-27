package k8s

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apl "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Pvc interface {
	GetPvc(namespace, name string) (*v1.PersistentVolumeClaim, error)
	ListPvc(namespace string, labels map[string]string) (*v1.PersistentVolumeClaimList, error)
}

type PvcService struct {
	KubeClient client.Client
	Log        logr.Logger
}

func NewPvcService(kubeClient client.Client, log logr.Logger) *PvcService {
	log = log.WithValues("service", "k8s.PvcService")
	return &PvcService{
		KubeClient: kubeClient,
		Log:        log,
	}
}

func (p PvcService) GetPvc(namespace, name string) (*v1.PersistentVolumeClaim, error) {
	var pvc = &v1.PersistentVolumeClaim{}
	if err := p.KubeClient.Get(context.Background(),
		types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		},
		pvc,
	); err != nil {
		return nil, err
	}
	return pvc, nil
}

func (p PvcService) ListPvc(namespace string, labels map[string]string) (*v1.PersistentVolumeClaimList, error) {
	var pvcList = &v1.PersistentVolumeClaimList{}
	if err := p.KubeClient.List(context.Background(),
		pvcList,
		// client.InNamespace(req.Namespace),
		&client.ListOptions{
			Namespace:     namespace,
			LabelSelector: apl.SelectorFromSet(labels),
		},
		// client.MatchingFields{ownerKey: req.Name},
	); err != nil {
		Error3(p.Log, err, "unable to list Pvc")
		return nil, err
	}
	return pvcList, nil
}
