package controllers

import (
	"github.com/zhizuqiu/redis-operator/controllers/service/element"
	"github.com/zhizuqiu/redis-operator/controllers/util"
)

func (r *RedisReconciler) Ensure(el element.Element) (element.Element, error) {
	err := util.NilError()

	el, err = r.RedisHandler.Ensurer.EnsureSentinelConfigMaps(el)
	if err != nil {
		return el, err
	}

	el, err = r.RedisHandler.Ensurer.EnsureRedisReadinessConfigMap(el)
	if err != nil {
		return el, err
	}

	el, err = r.RedisHandler.Ensurer.EnsureRedisMasterConfigMap(el)
	if err != nil {
		return el, err
	}

	el, err = r.RedisHandler.Ensurer.EnsureRedisSlaveConfigMaps(el)
	if err != nil {
		return el, err
	}

	el, err = r.RedisHandler.Ensurer.EnsureRedisStatefulSets(el)
	if err != nil {
		return el, err
	}

	el, err = r.RedisHandler.Ensurer.EnsureSentinelStatefulSets(el)
	if err != nil {
		return el, err
	}

	if el.Redis.Spec.Sentinel.Service.Enabled {
		el, err = r.RedisHandler.Ensurer.EnsureSentinelService(el)
		if err != nil {
			return el, err
		}
	}

	if !el.Redis.Spec.Sentinel.HostNetwork {
		el, err = r.RedisHandler.Ensurer.EnsureSentinelHeadlessService(el)
		if err != nil {
			return el, err
		}
	}

	if !el.Redis.Spec.Redis.HostNetwork {
		el, err = r.RedisHandler.Ensurer.EnsureRedisHeadlessService(el)
		if err != nil {
			return el, err
		}
	}

	if el.Redis.Spec.Exporter.Enabled {
		el, err = r.RedisHandler.Ensurer.EnsureExporterDeployment(el)
		if err != nil {
			return el, err
		}
	}

	return el, nil
}

func (r *RedisReconciler) DeleteEnsure(el element.Element) (element.Element, error) {

	err := util.NilError()

	el, err = r.RedisHandler.DeleteEnsurer.DeleteEnsureSentinelStatefulSets(el)
	if err != nil {
		return el, err
	}

	el, err = r.RedisHandler.DeleteEnsurer.DeleteEnsureRedisStatefulSets(el)
	if err != nil {
		return el, err
	}

	el, err = r.RedisHandler.DeleteEnsurer.DeleteEnsureSentinelPods(el)
	if err != nil {
		return el, err
	}

	el, err = r.RedisHandler.DeleteEnsurer.DeleteEnsureRedisPods(el)
	if err != nil {
		return el, err
	}

	el, err = r.RedisHandler.DeleteChecker.DeleteCheckSentinelPods(el)
	if err != nil {
		return el, err
	}

	el, err = r.RedisHandler.DeleteChecker.DeleteCheckRedisPods(el)
	if err != nil {
		return el, err
	}

	el, err = r.RedisHandler.DeleteEnsurer.DeleteEnsureSentinelPvcs(el)
	if err != nil {
		return el, err
	}

	el, err = r.RedisHandler.DeleteEnsurer.DeleteEnsureRedisPvcs(el)
	if err != nil {
		return el, err
	}

	el, err = r.RedisHandler.DeleteChecker.DeleteCheckSentinelPvcs(el)
	if err != nil {
		return el, err
	}

	el, err = r.RedisHandler.DeleteChecker.DeleteCheckRedisPvcs(el)
	if err != nil {
		return el, err
	}

	return el, nil
}
