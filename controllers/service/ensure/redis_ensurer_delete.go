package ensure

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/zhizuqiu/redis-operator/controllers/service/element"
	"github.com/zhizuqiu/redis-operator/controllers/service/k8s"
	"github.com/zhizuqiu/redis-operator/controllers/util"
)

type RedisDeleteEnsure interface {
	DeleteEnsureSentinelStatefulSets(el element.Element) (element.Element, error)
	DeleteEnsureRedisStatefulSets(el element.Element) (element.Element, error)
	DeleteEnsureSentinelPods(el element.Element) (element.Element, error)
	DeleteEnsureRedisPods(el element.Element) (element.Element, error)
	DeleteEnsureSentinelPvcs(el element.Element) (element.Element, error)
	DeleteEnsureRedisPvcs(el element.Element) (element.Element, error)
}

type RedisDeleteEnsurer struct {
	K8SService k8s.Services
	Log        logr.Logger
}

func NewRedisDeleteEnsurer(k8sService k8s.Services, log logr.Logger) *RedisDeleteEnsurer {
	log = log.WithValues("ensure", "RedisDeleteEnsurer")
	return &RedisDeleteEnsurer{
		K8SService: k8sService,
		Log:        log,
	}
}

// --- DeleteEnsureSentinelStatefulSets ---
func (r RedisDeleteEnsurer) DeleteEnsureSentinelStatefulSets(el element.Element) (element.Element, error) {
	err := util.NilError()

	labels := util.GetSentinelLabels(el.Redis)

	if err = r.deleteListStatefulSets(el.Redis.Namespace, labels); err != nil {
		return el, err
	}

	return el, nil
}

// --- DeleteEnsureRedisStatefulSets ---
func (r RedisDeleteEnsurer) DeleteEnsureRedisStatefulSets(el element.Element) (element.Element, error) {
	err := util.NilError()

	labels := util.GetRedisLabels(el.Redis)

	if err = r.deleteListStatefulSets(el.Redis.Namespace, labels); err != nil {
		return el, err
	}

	return el, nil
}

func (r RedisDeleteEnsurer) deleteListStatefulSets(namespace string, labels map[string]string) error {
	statefulsetList, err := r.K8SService.ListStatefulSets(namespace, labels)
	if err != nil {
		return err
	}

	if len(statefulsetList.Items) < 1 {
		return nil
	}

	for _, statefulSet := range statefulsetList.Items {
		if err = r.K8SService.Delete(context.Background(), &statefulSet); err != nil {
			return err
		}
	}
	return nil
}

// --- DeleteEnsureSentinelPods ---
func (r RedisDeleteEnsurer) DeleteEnsureSentinelPods(el element.Element) (element.Element, error) {
	err := util.NilError()

	labels := util.GetSentinelLabels(el.Redis)

	if err = r.deleteListPods(el.Redis.Namespace, labels); err != nil {
		return el, err
	}

	return el, nil
}

// --- DeleteEnsureRedisPods ---
func (r RedisDeleteEnsurer) DeleteEnsureRedisPods(el element.Element) (element.Element, error) {
	err := util.NilError()

	labels := util.GetRedisLabels(el.Redis)

	if err = r.deleteListPods(el.Redis.Namespace, labels); err != nil {
		return el, err
	}

	return el, nil
}

func (r RedisDeleteEnsurer) deleteListPods(namespace string, labels map[string]string) error {
	podList, err := r.K8SService.ListPods(namespace, labels)
	if err != nil {
		return err
	}

	if len(podList.Items) < 1 {
		return nil
	}

	for _, pod := range podList.Items {
		if err = r.K8SService.Delete(context.Background(), &pod); err != nil {
			return err
		}
	}
	return nil
}

// --- DeleteEnsureSentinelPvcs ---
func (r RedisDeleteEnsurer) DeleteEnsureSentinelPvcs(el element.Element) (element.Element, error) {
	err := util.NilError()

	labels := util.GetSentinelLabels(el.Redis)

	if err = r.deleteListPvc(el.Redis.Namespace, labels); err != nil {
		return el, err
	}

	return el, nil
}

// --- DeleteEnsureRedisPvcs ---
func (r RedisDeleteEnsurer) DeleteEnsureRedisPvcs(el element.Element) (element.Element, error) {
	err := util.NilError()

	labels := util.GetRedisLabels(el.Redis)

	if err = r.deleteListPvc(el.Redis.Namespace, labels); err != nil {
		return el, err
	}

	return el, nil
}

func (r RedisDeleteEnsurer) deleteListPvc(namespace string, labels map[string]string) error {
	pvcList, err := r.K8SService.ListPvc(namespace, labels)
	if err != nil {
		return err
	}

	if len(pvcList.Items) < 1 {
		return nil
	}

	for _, pvc := range pvcList.Items {
		if err = r.K8SService.Delete(context.Background(), &pvc); err != nil {
			return err
		}
	}

	return nil
}

func (r RedisDeleteEnsurer) deleteListPv(labels map[string]string) error {
	pvList, err := r.K8SService.ListPv(labels)
	if err != nil {
		return err
	}

	if len(pvList.Items) < 1 {
		return nil
	}

	for _, pv := range pvList.Items {
		if err = r.K8SService.Delete(context.Background(), &pv); err != nil {
			return err
		}
	}

	return nil
}
