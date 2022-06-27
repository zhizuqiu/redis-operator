package check

import (
	"errors"
	"github.com/go-logr/logr"
	"github.com/zhizuqiu/redis-operator/controllers/service/element"
	"github.com/zhizuqiu/redis-operator/controllers/service/k8s"
	"github.com/zhizuqiu/redis-operator/controllers/util"
	"strconv"
)

type RedisDeleteChecke interface {
	DeleteCheckSentinelPods(el element.Element) (element.Element, error)
	DeleteCheckRedisPods(el element.Element) (element.Element, error)
	DeleteCheckSentinelPvcs(el element.Element) (element.Element, error)
	DeleteCheckRedisPvcs(el element.Element) (element.Element, error)
}

type RedisDeleteChecker struct {
	K8SService k8s.Services
	Log        logr.Logger
}

func NewRedisDeleteChecker(k8sService k8s.Services, log logr.Logger) *RedisDeleteChecker {
	log = log.WithValues("check", "RedisDeleteChecker")
	return &RedisDeleteChecker{
		K8SService: k8sService,
		Log:        log,
	}
}

func (r RedisDeleteChecker) DeleteCheckSentinelPods(el element.Element) (element.Element, error) {
	podList, err := r.K8SService.ListPods(el.Redis.Namespace, util.GetSentinelLabels(el.Redis))
	if err != nil {
		return el, err
	}
	if len(podList.Items) > 0 {
		return el, errors.New("sentinel pod's size = " + strconv.Itoa(len(podList.Items)))
	}
	return el, nil
}

func (r RedisDeleteChecker) DeleteCheckRedisPods(el element.Element) (element.Element, error) {
	podList, err := r.K8SService.ListPods(el.Redis.Namespace, util.GetRedisLabels(el.Redis))
	if err != nil {
		return el, err
	}
	if len(podList.Items) > 0 {
		return el, errors.New("redis pod's size = " + strconv.Itoa(len(podList.Items)))
	}
	return el, nil
}

func (r RedisDeleteChecker) DeleteCheckSentinelPvcs(el element.Element) (element.Element, error) {
	pvcList, err := r.K8SService.ListPvc(el.Redis.Namespace, util.GetSentinelLabels(el.Redis))
	if err != nil {
		return el, err
	}
	if len(pvcList.Items) > 0 {
		return el, errors.New("sentinel pvc's size = " + strconv.Itoa(len(pvcList.Items)))
	}
	return el, nil
}

func (r RedisDeleteChecker) DeleteCheckRedisPvcs(el element.Element) (element.Element, error) {
	pvcList, err := r.K8SService.ListPvc(el.Redis.Namespace, util.GetRedisLabels(el.Redis))
	if err != nil {
		return el, err
	}
	if len(pvcList.Items) > 0 {
		return el, errors.New("redis pvc's size = " + strconv.Itoa(len(pvcList.Items)))
	}
	return el, nil
}
