package k8s

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apl "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Pv interface {
	GetPv(name string) (*v1.PersistentVolume, error)
	ListPv(labels map[string]string) (*v1.PersistentVolumeList, error)
}

type PvService struct {
	KubeClient client.Client
	Log        logr.Logger
}

func NewPvService(kubeClient client.Client, log logr.Logger) *PvService {
	log = log.WithValues("service", "k8s.PvService")
	return &PvService{
		KubeClient: kubeClient,
		Log:        log,
	}
}

func (c *PvService) GetPv(name string) (*v1.PersistentVolume, error) {
	var pv = &v1.PersistentVolume{}
	if err := c.KubeClient.Get(context.Background(),
		types.NamespacedName{
			Name: name,
		},
		pv,
	); err != nil {
		return nil, err
	}
	return pv, nil
}

func (c *PvService) ListPv(labels map[string]string) (*v1.PersistentVolumeList, error) {
	var pvList = &v1.PersistentVolumeList{}
	if err := c.KubeClient.List(context.Background(),
		pvList,
		// client.InNamespace(req.Namespace),
		&client.ListOptions{
			LabelSelector: apl.SelectorFromSet(labels),
		},
		// client.MatchingFields{ownerKey: req.Name},
	); err != nil {
		Error3(c.Log, err, "unable to list Pv")
		return nil, err
	}
	return pvList, nil
}
