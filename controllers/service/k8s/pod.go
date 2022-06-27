package k8s

import (
	"context"
	"github.com/go-logr/logr"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	apl "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Pod interface {
	ListPods(namespace string, labels map[string]string) (*v1.PodList, error)
	UpdatePodStatus(redis *roav1.Redis, currentStatus roav1.State) error
}

type PodService struct {
	KubeClient client.Client
	Scheme     *runtime.Scheme
	Log        logr.Logger
}

func NewPodService(kubeClient client.Client, log logr.Logger, scheme *runtime.Scheme) *PodService {
	log = log.WithValues("service", "k8s.PodService")
	return &PodService{
		KubeClient: kubeClient,
		Log:        log,
		Scheme:     scheme,
	}
}

func (p PodService) ListPods(namespace string, labels map[string]string) (*v1.PodList, error) {
	var podList = &v1.PodList{}
	if err := p.KubeClient.List(context.Background(),
		podList,
		// client.InNamespace(req.Namespace),
		&client.ListOptions{
			Namespace:     namespace,
			LabelSelector: apl.SelectorFromSet(labels),
		},
		// client.MatchingFields{ownerKey: req.Name},
	); err != nil {
		Error3(p.Log, err, "unable to list Pod")
		return nil, err
	}
	return podList, nil
}

func (p PodService) UpdatePodStatus(redis *roav1.Redis, currentStatus roav1.State) error {
	redis.Status.State = currentStatus
	if err := p.KubeClient.Status().Update(context.Background(), redis); err != nil {
		return err
	}
	return nil
}
