package check

import (
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/zhizuqiu/redis-operator/controllers/service/element"
	"github.com/zhizuqiu/redis-operator/controllers/service/k8s"
	"github.com/zhizuqiu/redis-operator/controllers/service/redis_client"
	"github.com/zhizuqiu/redis-operator/controllers/util"
	corev1 "k8s.io/api/core/v1"
	"strconv"
	"time"
)

type RedisCheck interface {
	CheckRedisNumber(el element.Element) error
	CheckSentinelNumber(el element.Element) error
	CheckAllSlavesFromMaster(master redis_client.RedisParam, el element.Element) error
	CheckSentinelNumberInMemory(sentinel redis_client.RedisParam, el element.Element) error
	CheckSentinelSlavesNumberInMemory(sentinel redis_client.RedisParam, el element.Element) error
	CheckSentinelMonitor(sentinel redis_client.RedisParam, monitor ...string) error
	GetMasterPod(el element.Element) (redis_client.RedisParam, error)
	GetNumberMasters(el element.Element) (int, error)
	GetRedisPods(el element.Element) ([]redis_client.RedisParam, error)
	GetSentinelsPods(el element.Element) ([]redis_client.RedisParam, error)
	GetMinimumRedisPodTime(el element.Element) (time.Duration, error)
}

type RedisChecker struct {
	K8sService  k8s.Services
	RedisClient redis_client.RedisClient
	Log         logr.Logger
}

func NewRedisChecker(k8sService k8s.Services, redisClient redis_client.RedisClient, log logr.Logger) *RedisChecker {
	log = log.WithValues("check", "RedisChecker")
	return &RedisChecker{
		K8sService:  k8sService,
		RedisClient: redisClient,
		Log:         log,
	}
}

func (rc *RedisChecker) CheckRedisNumber(el element.Element) error {
	labels := util.GetRedisLabels(el.Redis)
	ss, err := rc.K8sService.ListStatefulSets(el.Redis.Namespace, labels)
	if err != nil {
		return err
	}
	has := true
	for i := 0; i < int(el.Redis.Spec.Redis.Replicas); i++ {
		name := util.GetRedisNameByIndex(el.Redis, i)
		if util.SearchStatefulSetByName(name, ss) == nil {
			has = false
			break
		}
	}
	if !has {
		return errors.New("number of redis pods differ from specification")
	}
	return nil
}

func (rc *RedisChecker) CheckSentinelNumber(el element.Element) error {
	labels := util.GetSentinelLabels(el.Redis)
	d, err := rc.K8sService.ListStatefulSets(el.Redis.Namespace, labels)
	if err != nil {
		return err
	}
	has := true
	for i := 0; i < int(el.Redis.Spec.Redis.Replicas); i++ {
		name := util.GetSentinelNameByIndex(el.Redis, i)
		if util.SearchStatefulSetByName(name, d) == nil {
			has = false
			break
		}
	}
	if !has {
		return errors.New("number of sentinel pods differ from specification")
	}
	return nil
}

func (rc *RedisChecker) CheckAllSlavesFromMaster(master redis_client.RedisParam, el element.Element) error {
	redisPods, err := rc.GetRedisPods(el)
	if err != nil {
		return err
	}

	for _, redisPod := range redisPods {
		password, err := rc.RedisClient.GetRedisPassword(redisPod)
		if err != nil {
			return err
		}
		slave, err := rc.RedisClient.GetSlaveOf(redisPod, password)
		if err != nil {
			return err
		}
		if slave != "" && slave != master.Ip {
			return fmt.Errorf("slave %s don't have the master %s, has %s", redisPod.Name, master, slave)
		}
	}
	return nil
}

func (rc *RedisChecker) CheckSentinelNumberInMemory(sentinel redis_client.RedisParam, el element.Element) error {
	nSentinels, err := rc.RedisClient.GetNumberSentinelsInMemory(sentinel)
	if err != nil {
		return err
	} else if nSentinels != el.Redis.Spec.Sentinel.Replicas {
		return errors.New("sentinels in memory mismatch")
	}
	return nil
}

func (rc *RedisChecker) CheckSentinelSlavesNumberInMemory(sentinel redis_client.RedisParam, el element.Element) error {
	nSlaves, err := rc.RedisClient.GetNumberSentinelSlavesInMemory(sentinel)
	if err != nil {
		return err
	} else if nSlaves != el.Redis.Spec.Redis.Replicas-1 {
		return errors.New("redis slaves in sentinel memory mismatch")
	}
	return nil
}

func (rc *RedisChecker) CheckSentinelMonitor(sentinel redis_client.RedisParam, monitor ...string) error {

	monitorIP := monitor[0]
	monitorPort := ""
	if len(monitor) > 1 {
		monitorPort = monitor[1]
	}
	actualMonitorIP, actualMonitorPort, err := rc.RedisClient.GetSentinelMonitor(sentinel)
	if err != nil {
		return err
	}
	if actualMonitorIP != monitorIP || (monitorPort != "" && monitorPort != actualMonitorPort) {
		return errors.New("the monitor on the sentinel config does not match with the expected one")
	}
	return nil
}

func (rc *RedisChecker) GetMasterPod(el element.Element) (redis_client.RedisParam, error) {
	redisPods, err := rc.GetRedisPods(el)
	if err != nil {
		return redis_client.RedisParam{}, err
	}

	masterExecPods := []redis_client.RedisParam{}
	for _, redisPod := range redisPods {
		password, err := rc.RedisClient.GetRedisPassword(
			redis_client.RedisParam{
				NameSpace: redisPod.NameSpace,
				Name:      redisPod.Name,
			})
		if err != nil {
			return redis_client.RedisParam{}, err
		}
		master, err := rc.RedisClient.IsMaster(redisPod, password)
		if err != nil {
			return redis_client.RedisParam{}, err
		}
		if master {
			masterExecPods = append(masterExecPods, redisPod)
		}
	}

	if len(masterExecPods) != 1 {
		return redis_client.RedisParam{}, errors.New("number of redis nodes known as master is different than 1")
	}
	return masterExecPods[0], nil
}

func getStatefulSetPodNames(statefulSetName string, replicas int32) []string {
	names := make([]string, 0)
	for i := int32(0); i < replicas; i++ {
		names = append(names, statefulSetName+"-"+strconv.Itoa(int(i))+"-0")
	}
	return names
}

func (rc *RedisChecker) GetNumberMasters(el element.Element) (int, error) {
	nMasters := 0

	podNames := getStatefulSetPodNames(util.GetRedisRootName(el.Redis), el.Redis.Spec.Redis.Replicas)

	for _, podName := range podNames {
		password, err := rc.RedisClient.GetRedisPassword(
			redis_client.RedisParam{
				NameSpace: el.Redis.Namespace,
				Name:      podName,
			})
		if err != nil {
			return nMasters, err
		}

		master, err := rc.RedisClient.IsMaster(
			redis_client.RedisParam{
				NameSpace: el.Redis.Namespace,
				Name:      podName,
			},
			password,
		)
		if err != nil {
			return nMasters, err
		}
		if master {
			nMasters++
		}
	}
	return nMasters, nil
}

func (rc *RedisChecker) GetRedisPods(el element.Element) ([]redis_client.RedisParam, error) {
	var redises []redis_client.RedisParam
	podList, err := rc.K8sService.ListPods(el.Redis.Namespace, util.GetRedisLabels(el.Redis))
	if err != nil {
		return nil, err
	}
	for _, rp := range podList.Items {
		if rp.Status.Phase == corev1.PodRunning && rp.DeletionTimestamp == nil { // Only work with running pods
			redises = append(redises, redis_client.RedisParam{
				Ip:        rp.Status.PodIP,
				NameSpace: rp.Namespace,
				Name:      rp.Name,
			})
		}
	}
	return redises, nil
}

func (rc *RedisChecker) GetSentinelsPods(el element.Element) ([]redis_client.RedisParam, error) {
	sentinels := []redis_client.RedisParam{}
	rps, err := rc.K8sService.ListPods(el.Redis.Namespace, util.GetSentinelLabels(el.Redis))
	if err != nil {
		return nil, err
	}
	for _, sp := range rps.Items {
		if sp.Status.Phase == corev1.PodRunning && sp.DeletionTimestamp == nil { // Only work with running pods
			sentinels = append(sentinels, redis_client.RedisParam{
				NameSpace: sp.Namespace,
				Name:      sp.Name,
				Ip:        sp.Status.PodIP,
			})
		}
	}
	return sentinels, nil
}

func (rc *RedisChecker) GetMinimumRedisPodTime(el element.Element) (time.Duration, error) {
	minTime := 100000 * time.Hour // More than ten years
	rps, err := rc.K8sService.ListPods(el.Redis.Namespace, util.GetRedisLabels(el.Redis))
	if err != nil {
		return minTime, err
	}
	for _, redisNode := range rps.Items {
		if redisNode.Status.StartTime == nil {
			continue
		}
		start := redisNode.Status.StartTime.Round(time.Second)
		alive := time.Now().Sub(start)
		rc.Log.Info("State " + redisNode.Status.PodIP + " has been alive for " + util.Floadt64ToString(alive.Seconds()) + " seconds")
		if alive < minTime {
			minTime = alive
		}
	}
	return minTime, nil
}
