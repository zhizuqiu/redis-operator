package check

import (
	"errors"
	"github.com/go-logr/logr"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	"github.com/zhizuqiu/redis-operator/controllers/service/k8s"
	"github.com/zhizuqiu/redis-operator/controllers/service/redis_client"
	"github.com/zhizuqiu/redis-operator/controllers/util"
	"sort"
	"strconv"
)

type RedisHeal interface {
	MakeMaster(redisPod redis_client.RedisParam, rs *roav1.Redis) error
	SetOldestAsMaster(rs *roav1.Redis) error
	SetMasterOnAll(masterIP string, rs *roav1.Redis) error
	NewSentinelMonitor(sentinel redis_client.RedisParam, monitor string, rs *roav1.Redis) error
	RestoreSentinel(sentinel redis_client.RedisParam) error
	SetSentinelCustomConfig(sentinel redis_client.RedisParam, rs *roav1.Redis) error
	SetRedisCustomConfig(redisPod redis_client.RedisParam, rs *roav1.Redis) error
	UpdateRedisConfigStatus(redis *roav1.Redis, currentStatus roav1.RedisConfig) error
	UpdateSentinelConfigStatus(redis *roav1.Redis, currentStatus roav1.SentinelConfig) error
	UpdateRedisPasswordStatus(redis *roav1.Redis, currentStatus roav1.RedisPassword) error
	UpdatePodStatus(redis *roav1.Redis, currentStatus roav1.State) error
	UpdateSentinelPasswordStatus(redis *roav1.Redis, currentStatus roav1.RedisPassword) error
	SetRedisPassword(redisPod redis_client.RedisParam, rs *roav1.Redis) error
	SetSentinelPassword(redisPod redis_client.RedisParam, rs *roav1.Redis) error
}

type RedisHealer struct {
	K8sService  k8s.Services
	RedisClient redis_client.RedisClient
	Log         logr.Logger
}

func NewRedisHealer(k8sService k8s.Services, redisClient redis_client.RedisClient, log logr.Logger) *RedisHealer {
	log = log.WithValues("check", "RedisHealer")
	return &RedisHealer{
		K8sService:  k8sService,
		RedisClient: redisClient,
		Log:         log,
	}
}

func (r RedisHealer) MakeMaster(redisPod redis_client.RedisParam, rs *roav1.Redis) error {
	password, err := r.RedisClient.GetRedisPassword(redisPod)
	if err != nil {
		return err
	}

	return r.RedisClient.MakeMaster(redisPod, password)
}

func (r RedisHealer) SetOldestAsMaster(rf *roav1.Redis) error {
	ssp, err := r.K8sService.ListPods(rf.Namespace, util.GetRedisLabels(rf))
	if err != nil {
		return err
	}
	if len(ssp.Items) < 1 {
		return errors.New("number of redis pods are 0")
	}

	// Order the pods so we start by the oldest one
	sort.Slice(ssp.Items, func(i, j int) bool {
		return ssp.Items[i].CreationTimestamp.Before(&ssp.Items[j].CreationTimestamp)
	})

	newMasterIP := ""
	for _, pod := range ssp.Items {
		password, err := r.RedisClient.GetRedisPassword(
			redis_client.RedisParam{
				NameSpace: pod.Namespace,
				Name:      pod.Name,
			})
		if err != nil {
			return err
		}

		if newMasterIP == "" {
			newMasterIP = pod.Status.PodIP
			Info(r.Log, "New master is "+pod.Name+" with ip "+newMasterIP, rf)
			if err := r.RedisClient.MakeMaster(
				redis_client.RedisParam{
					NameSpace: pod.Namespace,
					Name:      pod.Name,
				},
				password,
			); err != nil {
				return err
			}
		} else {
			Info(r.Log, "Making pod "+pod.Name+" slave of "+newMasterIP, rf)
			if err := r.RedisClient.MakeSlaveOf(
				redis_client.RedisParam{
					NameSpace: pod.Namespace,
					Name:      pod.Name,
				},
				password,
				newMasterIP,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r RedisHealer) SetMasterOnAll(masterIP string, rf *roav1.Redis) error {
	ssp, err := r.K8sService.ListPods(rf.Namespace, util.GetRedisLabels(rf))
	if err != nil {
		return err
	}

	for _, pod := range ssp.Items {
		password, err := r.RedisClient.GetRedisPassword(
			redis_client.RedisParam{
				NameSpace: pod.Namespace,
				Name:      pod.Name,
			})
		if err != nil {
			return err
		}
		if pod.Status.PodIP == masterIP {
			Info(r.Log, "Ensure pod "+pod.Name+" is master", rf)
			if err := r.RedisClient.MakeMaster(
				redis_client.RedisParam{
					NameSpace: pod.Namespace,
					Name:      pod.Name,
				},
				password,
			); err != nil {
				return err
			}
		} else {
			Info(r.Log, "Making pod "+pod.Name+" slave of "+masterIP, rf)
			if err := r.RedisClient.MakeSlaveOf(
				redis_client.RedisParam{
					NameSpace: pod.Namespace,
					Name:      pod.Name,
				},
				password,
				masterIP,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r RedisHealer) NewSentinelMonitor(sentinel redis_client.RedisParam, monitor string, rf *roav1.Redis) error {
	Info(r.Log, "Sentinel is not monitoring the correct master, changing...", rf)
	quorum := strconv.Itoa(int(util.GetQuorum(rf)))

	password, err := k8s.GetSpecRedisPassword(r.K8sService, rf)
	if err != nil {
		return err
	}

	return r.RedisClient.MonitorRedis(sentinel, monitor, quorum, password)
}

func (r RedisHealer) RestoreSentinel(sentinel redis_client.RedisParam) error {
	Info2(r.Log, "Restoring sentinel "+sentinel.Ip+"...", sentinel)
	return r.RedisClient.ResetSentinel(sentinel)
}

func (r RedisHealer) SetSentinelCustomConfig(sentinel redis_client.RedisParam, rf *roav1.Redis) error {
	Info(r.Log, "Setting the custom config on sentinel "+sentinel.Ip+"...", rf)
	return r.RedisClient.SetCustomSentinelConfig(sentinel, rf.Spec.Sentinel.CustomConfig)
}

func (r RedisHealer) SetRedisCustomConfig(redisPod redis_client.RedisParam, rf *roav1.Redis) error {
	Info(r.Log, "Setting the custom config on redis "+redisPod.Ip+"...", rf)

	password, err := r.RedisClient.GetRedisPassword(redisPod)
	if err != nil {
		return err
	}

	return r.RedisClient.SetCustomRedisConfig(redisPod, rf.Spec.Redis.CustomConfig, password)
}

func (r RedisHealer) UpdateRedisConfigStatus(redis *roav1.Redis, currentStatus roav1.RedisConfig) error {
	return r.K8sService.UpdateRedisConfigStatus(redis, currentStatus)
}

func (r RedisHealer) UpdateSentinelConfigStatus(redis *roav1.Redis, currentStatus roav1.SentinelConfig) error {
	return r.K8sService.UpdateSentinelConfigStatus(redis, currentStatus)
}

func (r RedisHealer) UpdateRedisPasswordStatus(redis *roav1.Redis, currentStatus roav1.RedisPassword) error {
	return r.K8sService.UpdateRedisPasswordStatus(redis, currentStatus)
}

func (r RedisHealer) UpdatePodStatus(redis *roav1.Redis, currentStatus roav1.State) error {
	return r.K8sService.UpdatePodStatus(redis, currentStatus)
}

func (r RedisHealer) UpdateSentinelPasswordStatus(redis *roav1.Redis, currentStatus roav1.RedisPassword) error {
	return r.K8sService.UpdateSentinelPasswordStatus(redis, currentStatus)
}

func (r RedisHealer) SetRedisPassword(redisPod redis_client.RedisParam, rs *roav1.Redis) error {
	newPassword, err := k8s.GetSpecRedisPassword(r.K8sService, rs)
	if err != nil {
		return err
	}
	return r.RedisClient.SetRedisPassword(redisPod, newPassword)
}

func (r RedisHealer) SetSentinelPassword(redisPod redis_client.RedisParam, rs *roav1.Redis) error {
	newPassword, err := k8s.GetSpecRedisPassword(r.K8sService, rs)
	if err != nil {
		return err
	}
	return r.RedisClient.SetSentinelPassword(redisPod, newPassword)
}
