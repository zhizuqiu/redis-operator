package k8s

import (
	"context"
	"github.com/go-logr/logr"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ref "k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Deployment interface {
	GetDeployment(namespace, name string) (*appsv1.Deployment, error)
	GetDeploymentObjectReference(deployment *appsv1.Deployment) corev1.ObjectReference
	GetDeploymentPods(namespace, name string) (*corev1.PodList, error)
	UpdateExporterState(redis *roav1.Redis, currentStatus roav1.ExporterState) error
}

type DeploymentService struct {
	KubeClient client.Client
	Scheme     *runtime.Scheme
	Log        logr.Logger
}

func NewDeploymentService(kubeClient client.Client, log logr.Logger, sheme *runtime.Scheme) *DeploymentService {
	log = log.WithValues("service", "k8s.DeploymentService")
	return &DeploymentService{
		KubeClient: kubeClient,
		Scheme:     sheme,
		Log:        log,
	}
}

func (d DeploymentService) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	var deployment = &appsv1.Deployment{}
	if err := d.KubeClient.Get(context.Background(),
		types.NamespacedName{Namespace: namespace, Name: name},
		deployment,
	); err != nil {
		return nil, err
	}
	return deployment, nil
}

func (d DeploymentService) GetDeploymentObjectReference(deployment *appsv1.Deployment) corev1.ObjectReference {
	if deployment == nil {
		return corev1.ObjectReference{}
	}

	if deployment.GetObjectMeta().GetDeletionTimestamp() != nil {
		return corev1.ObjectReference{}
	}

	referenceRef, err := ref.GetReference(d.Scheme, deployment)
	if err != nil {
		return corev1.ObjectReference{}
	}

	return *referenceRef
}

func (d DeploymentService) GetDeploymentPods(namespace, name string) (*corev1.PodList, error) {
	deployment, err := d.GetDeployment(namespace, name)
	if err != nil {
		return nil, err
	}

	labels := make(map[string]string)
	for k, v := range deployment.Spec.Selector.MatchLabels {
		labels[k] = v
	}
	return ListPods(d.KubeClient, namespace, labels)
}

func (d DeploymentService) UpdateExporterState(redis *roav1.Redis, currentStatus roav1.ExporterState) error {
	redis.Status.Exporter = currentStatus
	if err := d.KubeClient.Status().Update(context.Background(), redis); err != nil {
		return err
	}
	return nil
}
