package k8s

import (
	"context"
	"github.com/go-logr/logr"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ref "k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ConfigMap interface {
	GetConfigMap(namespace string, name string) (*corev1.ConfigMap, error)
	GetConfigMapObjectReference(configMap *corev1.ConfigMap) corev1.ObjectReference
	UpdateSentinelState(redis *roav1.Redis, currentStatus roav1.SentinelState) error
	UpdateRedisState(redis *roav1.Redis, currentStatus roav1.RedisState) error
}

type ConfigMapService struct {
	KubeClient client.Client
	Scheme     *runtime.Scheme
	Log        logr.Logger
}

func NewConfigMapService(kubeClient client.Client, log logr.Logger, scheme *runtime.Scheme) *ConfigMapService {
	log = log.WithValues("service", "k8s.ConfigMapService")
	return &ConfigMapService{
		KubeClient: kubeClient,
		Log:        log,
		Scheme:     scheme,
	}
}

func (c *ConfigMapService) GetConfigMap(namespace string, name string) (*corev1.ConfigMap, error) {

	var configMap = &corev1.ConfigMap{}
	if err := c.KubeClient.Get(context.Background(),
		types.NamespacedName{Namespace: namespace, Name: name},
		configMap,
	); err != nil {
		return nil, err
	}

	return configMap, nil
}

func (c *ConfigMapService) GetConfigMapObjectReference(configMap *corev1.ConfigMap) corev1.ObjectReference {

	if configMap == nil {
		return corev1.ObjectReference{}
	}

	if configMap.GetObjectMeta().GetDeletionTimestamp() != nil {
		return corev1.ObjectReference{}
	}

	referenceRef, err := ref.GetReference(c.Scheme, configMap)
	if err != nil {
		return corev1.ObjectReference{}
	}

	return *referenceRef
}

func (c *ConfigMapService) UpdateSentinelState(redis *roav1.Redis, currentStatus roav1.SentinelState) error {
	redis.Status.Sentinel = currentStatus
	if err := c.KubeClient.Status().Update(context.Background(), redis); err != nil {
		return err
	}
	return nil
}

func (c *ConfigMapService) UpdateRedisState(redis *roav1.Redis, currentStatus roav1.RedisState) error {
	redis.Status.Redis = currentStatus
	if err := c.KubeClient.Status().Update(context.Background(), redis); err != nil {
		return err
	}
	return nil
}
