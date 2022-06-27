package k8s

import (
	"context"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apl "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ref "k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// StatefulSet the StatefulSet service that knows how to interact with k8s to manage them
type StatefulSet interface {
	GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error)
	ListStatefulSets(namespace string, labels map[string]string) (*appsv1.StatefulSetList, error)
	GetStatefulSetObjectReference(statefulSet *appsv1.StatefulSet) corev1.ObjectReference
}

type StatefulSetService struct {
	KubeClient client.Client
	Scheme     *runtime.Scheme
	Log        logr.Logger
}

func NewStatefulSetService(kubeClient client.Client, log logr.Logger, scheme *runtime.Scheme) *StatefulSetService {
	log = log.WithValues("service", "k8s.StatefulSetService")
	return &StatefulSetService{
		KubeClient: kubeClient,
		Scheme:     scheme,
		Log:        log,
	}
}

func (s *StatefulSetService) ListStatefulSets(namespace string, labels map[string]string) (*appsv1.StatefulSetList, error) {
	var statefulSetList = &appsv1.StatefulSetList{}
	if err := s.KubeClient.List(context.Background(),
		statefulSetList,
		&client.ListOptions{
			Namespace:     namespace,
			LabelSelector: apl.SelectorFromSet(labels),
		},
	); err != nil {
		return nil, err
	}
	return statefulSetList, nil
}

func (s *StatefulSetService) GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error) {
	var statefulSet = &appsv1.StatefulSet{}
	if err := s.KubeClient.Get(context.Background(),
		types.NamespacedName{Namespace: namespace, Name: name},
		statefulSet,
	); err != nil {
		return nil, err
	}
	return statefulSet, nil
}

func (s *StatefulSetService) GetStatefulSetObjectReference(statefulSet *appsv1.StatefulSet) corev1.ObjectReference {
	if statefulSet == nil {
		return corev1.ObjectReference{}
	}

	if statefulSet.GetObjectMeta().GetDeletionTimestamp() != nil {
		// return nil, errors.NewNotFound(appsv1.Resource("statefulset"), statefulSet.Name)
		return corev1.ObjectReference{}
	}

	referenceRef, err := ref.GetReference(s.Scheme, statefulSet)
	if err != nil {
		return corev1.ObjectReference{}
	}

	return *referenceRef
}
