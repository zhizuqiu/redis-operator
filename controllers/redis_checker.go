package controllers

import (
	"encoding/json"
	"errors"
	"github.com/zhizuqiu/redis-operator/controllers/service/element"
	"github.com/zhizuqiu/redis-operator/controllers/service/k8s"
	"github.com/zhizuqiu/redis-operator/controllers/service/redis_client"
	"github.com/zhizuqiu/redis-operator/controllers/util"
	"reflect"
	"strconv"
	"time"
)

const (
	// timeToPrepare = 2 * time.Minute
	timeToPrepare = 30 * time.Second
)

func (r *RedisReconciler) CheckAndHeal(el element.Element) (element.Element, error) {
	log := r.Log.WithValues("controller", "CheckAndHeal")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	// Number of redis is equal as the set on the RF spec
	// Number of sentinel is equal as the set on the RF spec
	// Check only one master
	// Number of redis master is 1
	// All redis slaves have the same master
	// All sentinels points to the same redis master
	// Sentinel has not death nodes
	// Sentinel knows the correct slave number

	err := util.NilError()
	el, err = r.checkState(el)
	if err != nil {
		return el, err
	}

	el, err, needCheckAndHealCustomConfig := r.needCheckAndHealCustomConfig(el)
	if err != nil {
		return el, err
	}

	el, err, needCheckAndHealPassword := r.needCheckAndHealPassword(el)
	if err != nil {
		return el, err
	}

	if !needCheckAndHealCustomConfig && !needCheckAndHealPassword && el.Redis.Status.State.Cluster {
		Info(log, "cluster = true, skip CheckAndHeal()", el.Redis)
		return el, nil
	}

	if !needCheckAndHealCustomConfig && !needCheckAndHealPassword && !util.IsNeedAutoFailover(el.Redis) {
		Info(log, "no need to auto failover, skip CheckAndHeal()", el.Redis)
		return el, nil
	}

	err = r.checkNumber(el)
	if err != nil {
		return el, err
	}

	el, err = r.checkMaster(el)
	if err != nil {
		return el, err
	}

	el, err = r.checkAndHeal(el)
	if err != nil {
		return el, err
	}

	/*
		err = r.UpdateRedisesPods(el.Redis)
		if err != nil {
			return err
		}
	*/

	el, err = r.checkAndHealCustomConfig(el)
	if err != nil {
		return el, err
	}

	el, err = r.checkAndHealPassword(el)
	if err != nil {
		return el, err
	}

	return el, nil
}

// --- checkState ---
func (r *RedisReconciler) checkState(el element.Element) (element.Element, error) {
	// log := r.Log.WithValues("controller", "checkState")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	previousStatus := el.Redis.Status.State
	currentStatus := *el.Redis.Status.State.DeepCopy()

	podList, err := r.RedisHandler.K8sServices.ListPods(el.Redis.Namespace, util.GetInstanceLabels(el.Redis.Name))
	if err != nil {
		return el, err
	}
	currentStatus.Pods = r.RedisHandler.getPodStates(podList)
	currentStatus.Phase = util.GetGlobalPhase(el.Redis, podList)
	currentStatus.Ready = util.GetGlobalReady(el.Redis, podList)

	if !reflect.DeepEqual(previousStatus, currentStatus) {
		Info(r.Log, "State Status not equal", el.Redis)
		if err := r.RedisHandler.Healer.UpdatePodStatus(el.Redis, currentStatus); err != nil {
			return el, err
		}
		el.NeedReLoad = true
	} else {
		Info(r.Log, "State Status equal", el.Redis)
	}
	return el, nil
}

// --- checkNumber ---
func (r *RedisReconciler) checkNumber(el element.Element) error {
	log := r.Log.WithValues("controller", "checkNumber")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	err := util.NilError()
	err = r.RedisHandler.Checker.CheckRedisNumber(el)
	if err != nil {
		Error(log, err, "Number of redis mismatch, this could be for a change on the statefulset", el.Redis)
		return err
	}
	if err = r.RedisHandler.Checker.CheckSentinelNumber(el); err != nil {
		Error(log, err, "Number of sentinel mismatch, this could be for a change on the deployment", el.Redis)
		return err
	}

	return nil
}

// --- checkMaster ---
func (r *RedisReconciler) checkMaster(el element.Element) (element.Element, error) {
	log := r.Log.WithValues("controller", "checkMaster")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	nMasters, err := r.RedisHandler.Checker.GetNumberMasters(el)
	if err != nil {
		return el, err
	}
	Info(log, "Master Number:"+strconv.Itoa(nMasters), el.Redis)

	switch nMasters {
	case 0:
		el.NeedReCheckError = append(el.NeedReCheckError, errors.New("Master Number = 0"))
		redisePods, err := r.RedisHandler.Checker.GetRedisPods(el)
		if err != nil {
			return el, err
		}
		if len(redisePods) == 1 {
			if err = r.RedisHandler.Healer.MakeMaster(redisePods[0], el.Redis); err != nil {
				return el, err
			}
			break
		}
		minTime, err2 := r.RedisHandler.Checker.GetMinimumRedisPodTime(el)
		if err2 != nil {
			return el, err2
		}
		Info(log, "minTime:"+minTime.String(), el.Redis)
		if minTime > timeToPrepare {
			el.NeedReCheckError = append(el.NeedReCheckError, errors.New("time "+util.Floadt64ToString(minTime.Round(time.Second).Seconds())+" more than expected. Not even one master, fixing..."))
			Info(log, "time "+util.Floadt64ToString(minTime.Round(time.Second).Seconds())+" more than expected. Not even one master, fixing...", el.Redis)
			// We can consider there's an error
			if err2 = r.RedisHandler.Healer.SetOldestAsMaster(el.Redis); err2 != nil {
				return el, err2
			}
		} else {
			// We'll wait until failover is done
			el.NeedReCheckError = append(el.NeedReCheckError, errors.New("No master found, wait until failover"))
			Info(log, "No master found, wait until failover", el.Redis)
			return el, nil
		}
	case 1:
		break
	default:
		Info(log, "More than one master, fix manually", el.Redis)
		return el, nil
	}
	return el, nil
}

// --- checkAndHeal ---
func (r *RedisReconciler) checkAndHeal(el element.Element) (element.Element, error) {
	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	masterPod, err := r.RedisHandler.Checker.GetMasterPod(el)
	if err != nil {
		return el, err
	}

	el, err = r.checkAndHealRedis(el, masterPod)
	if err != nil {
		return el, err
	}

	el, err = r.checkAndHealSentinels(el, masterPod)
	if err != nil {
		return el, err
	}

	return el, nil
}

func (r *RedisReconciler) checkAndHealRedis(el element.Element, masterPod redis_client.RedisParam) (element.Element, error) {
	log := r.Log.WithValues("controller", "checkAndHealRedis")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	if err2 := r.RedisHandler.Checker.CheckAllSlavesFromMaster(masterPod, el); err2 != nil {
		Info(log, "Not all slaves have the same master", el.Redis)
		if err3 := r.RedisHandler.Healer.SetMasterOnAll(masterPod.Ip, el.Redis); err3 != nil {
			return el, err3
		}
	}
	return el, nil
}

func (r *RedisReconciler) checkAndHealSentinels(el element.Element, masterPod redis_client.RedisParam) (element.Element, error) {
	log := r.Log.WithValues("controller", "checkAndHealSentinels")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	sentinels, err := r.RedisHandler.Checker.GetSentinelsPods(el)
	if err != nil {
		return el, err
	}

	for _, sip := range sentinels {
		if err = r.RedisHandler.Checker.CheckSentinelMonitor(sip, masterPod.Ip); err != nil {
			el.NeedReCheckError = append(el.NeedReCheckError, errors.New("Sentinel is not monitoring the correct master"))
			Info(log, "Sentinel is not monitoring the correct master", el.Redis)
			if err = r.RedisHandler.Healer.NewSentinelMonitor(sip, masterPod.Ip, el.Redis); err != nil {
				return el, err
			}
		}
	}

	for _, sip := range sentinels {
		if err := r.RedisHandler.Checker.CheckSentinelNumberInMemory(sip, el); err != nil {
			el.NeedReCheckError = append(el.NeedReCheckError, errors.New(sip.Name+": Sentinel has more sentinel in memory than spected"))
			Error(log, err, sip.Name+": Sentinel has more sentinel in memory than spected", el.Redis)
			if err = r.RedisHandler.Healer.RestoreSentinel(sip); err != nil {
				return el, err
			}
		}
	}
	for _, sip := range sentinels {
		if err := r.RedisHandler.Checker.CheckSentinelSlavesNumberInMemory(sip, el); err != nil {
			el.NeedReCheckError = append(el.NeedReCheckError, errors.New(sip.Name+": Sentinel has more slaves in memory than spected"))
			Error(log, err, sip.Name+": Sentinel has more slaves in memory than spected", el.Redis)
			if err = r.RedisHandler.Healer.RestoreSentinel(sip); err != nil {
				return el, err
			}
		}
	}

	return el, nil
}

// --- checkAndHealCustomConfig ---
func (r *RedisReconciler) checkAndHealCustomConfig(el element.Element) (element.Element, error) {

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	err := util.NilError()
	el, err = r.checkAndHealRedisCustomConfig(el)
	if err != nil {
		return el, err
	}

	el, err = r.checkAndHealSentinelCustomConfig(el)
	if err != nil {
		return el, err
	}

	return el, nil
}

func (r *RedisReconciler) checkAndHealRedisCustomConfig(el element.Element) (element.Element, error) {
	log := r.Log.WithValues("controller", "checkAndHealRedisCustomConfig")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	if el.Redis.Spec.Redis.CustomConfig != nil {
		previousStatus := el.Redis.Status.Redis.RedisCustomConfig
		currentStatus := *el.Redis.Status.Redis.RedisCustomConfig.DeepCopy()

		configJsonByte, err := json.Marshal(el.Redis.Spec.Redis.CustomConfig)
		if err != nil {
			return el, err
		}
		currentStatus.Md5 = util.MD5(string(configJsonByte))

		if !reflect.DeepEqual(previousStatus, currentStatus) {
			el.NeedReCheckError = append(el.NeedReCheckError, errors.New("RedisCustomConfig Status not equal"))
			Info(log, "RedisCustomConfig Status not equal", el.Redis)
			if err = r.applyRedisCustomConfig(el); err != nil {
				return el, err
			}
			if err = r.RedisHandler.Healer.UpdateRedisConfigStatus(el.Redis, currentStatus); err != nil {
				return el, err
			}
			el.NeedReLoad = true
		} else {
			Info(log, "RedisCustomConfig Status equal", el.Redis)
		}
	}
	return el, nil
}

func (r *RedisReconciler) applyRedisCustomConfig(el element.Element) error {

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	redises, err := r.RedisHandler.Checker.GetRedisPods(el)
	if err != nil {
		return err
	}
	for _, rip := range redises {
		if err := r.RedisHandler.Healer.SetRedisCustomConfig(rip, el.Redis); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisReconciler) checkAndHealSentinelCustomConfig(el element.Element) (element.Element, error) {
	log := r.Log.WithValues("controller", "checkAndHealSentinelCustomConfig")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	if el.Redis.Spec.Sentinel.CustomConfig != nil {
		previousStatus := el.Redis.Status.Sentinel.SentinelCustomConfig
		currentStatus := *el.Redis.Status.Sentinel.SentinelCustomConfig.DeepCopy()

		configJsonByte, err := json.Marshal(el.Redis.Spec.Sentinel.CustomConfig)
		if err != nil {
			return el, err
		}
		currentStatus.Md5 = util.MD5(string(configJsonByte))

		if !reflect.DeepEqual(previousStatus, currentStatus) {
			el.NeedReCheckError = append(el.NeedReCheckError, errors.New("SentinelCustomConfig Status not equal"))
			Info(log, "SentinelCustomConfig Status not equal", el.Redis)

			sentinels, err := r.RedisHandler.Checker.GetSentinelsPods(el)
			if err != nil {
				return el, err
			}
			for _, sip := range sentinels {
				if err = r.RedisHandler.Healer.SetSentinelCustomConfig(sip, el.Redis); err != nil {
					return el, err
				}
			}
			if err = r.RedisHandler.Healer.UpdateSentinelConfigStatus(el.Redis, currentStatus); err != nil {
				return el, err
			}
			el.NeedReLoad = true
		} else {
			Info(log, "CustomConfig Status equal", el.Redis)
		}
	}

	return el, nil
}

// --- needCheckAndHealCustomConfig ---
func (r *RedisReconciler) needCheckAndHealCustomConfig(el element.Element) (element.Element, error, bool) {
	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err, false
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	el, err, need := r.needCheckAndHealRedisCustomConfig(el)
	if err != nil {
		return el, err, need
	}
	if need {
		return el, nil, true
	}

	el, err, need = r.needCheckAndHealSentinelCustomConfig(el)
	if err != nil {
		return el, err, need
	}
	if need {
		return el, nil, true
	}

	return el, nil, false
}

func (r *RedisReconciler) needCheckAndHealRedisCustomConfig(el element.Element) (element.Element, error, bool) {
	log := r.Log.WithValues("controller", "needCheckAndHealRedisCustomConfig")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err, false
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	if el.Redis.Spec.Redis.CustomConfig != nil {
		previousStatus := el.Redis.Status.Redis.RedisCustomConfig
		currentStatus := *el.Redis.Status.Redis.RedisCustomConfig.DeepCopy()

		configJsonByte, err := json.Marshal(el.Redis.Spec.Redis.CustomConfig)
		if err != nil {
			return el, err, false
		}
		currentStatus.Md5 = util.MD5(string(configJsonByte))

		if !reflect.DeepEqual(previousStatus, currentStatus) {
			Info(log, "need check and heal redis custom config", el.Redis)
			return el, nil, true
		} else {
			Info(log, "no need to check and heal redis custom config", el.Redis)
			return el, nil, false
		}
	}

	return el, nil, false
}

func (r *RedisReconciler) needCheckAndHealSentinelCustomConfig(el element.Element) (element.Element, error, bool) {
	log := r.Log.WithValues("controller", "needCheckAndHealSentinelCustomConfig")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err, false
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	if el.Redis.Spec.Sentinel.CustomConfig != nil {
		previousStatus := el.Redis.Status.Sentinel.SentinelCustomConfig
		currentStatus := *el.Redis.Status.Sentinel.SentinelCustomConfig.DeepCopy()

		configJsonByte, err := json.Marshal(el.Redis.Spec.Sentinel.CustomConfig)
		if err != nil {
			return el, err, false
		}
		currentStatus.Md5 = util.MD5(string(configJsonByte))

		if !reflect.DeepEqual(previousStatus, currentStatus) {
			Info(log, "need check and heal sentinel custom config", el.Redis)
			return el, nil, true
		} else {
			Info(log, "no need to check and heal sentinel custom config", el.Redis)
			return el, nil, false
		}
	}

	return el, nil, false
}

// --- checkAndHealPassword ---
func (r *RedisReconciler) checkAndHealPassword(el element.Element) (element.Element, error) {
	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	err := util.NilError()
	el, err = r.checkAndHealSentinelPassword(el)
	if err != nil {
		return el, err
	}

	el, err = r.checkAndHealRedisPassword(el)
	if err != nil {
		return el, err
	}

	return el, nil
}

func (r *RedisReconciler) checkAndHealRedisPassword(el element.Element) (element.Element, error) {
	log := r.Log.WithValues("controller", "checkAndHealRedisPassword")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	previousStatus := el.Redis.Status.Redis.RedisPassword
	currentStatus := *el.Redis.Status.Redis.RedisPassword.DeepCopy()

	password, err := k8s.GetSpecRedisPassword(r.RedisHandler.K8sServices, el.Redis)
	if err != nil {
		return el, err
	}
	currentStatus.Md5 = util.MD5(password)

	if !reflect.DeepEqual(previousStatus, currentStatus) {
		el.NeedReCheckError = append(el.NeedReCheckError, errors.New("RedisPassword Status not equal"))
		Info(log, "RedisPassword Status not equal", el.Redis)
		if err = r.applyRedisPassword(el); err != nil {
			return el, err
		}
		if err := r.RedisHandler.Healer.UpdateRedisPasswordStatus(el.Redis, currentStatus); err != nil {
			return el, err
		}
		el.NeedReLoad = true
	} else {
		Info(log, "RedisPassword Status equal", el.Redis)
	}
	return el, nil
}

func (r *RedisReconciler) applyRedisPassword(el element.Element) error {
	log := r.Log.WithValues("controller", "applyRedisPassword")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	masterPod, err := r.RedisHandler.Checker.GetMasterPod(el)
	if err != nil {
		return err
	}
	redises, err := r.RedisHandler.Checker.GetRedisPods(el)
	if err != nil {
		return err
	}
	success := false
	for _, redisPod := range redises {
		if reflect.DeepEqual(masterPod, redisPod) {
			Info(log, "starting set redis master "+redisPod.Name+" new password...", el.Redis)
			if err := r.RedisHandler.Healer.SetRedisPassword(redisPod, el.Redis); err != nil {
				return err
			}
			success = true
		}
	}
	if success {
		for _, redisPod := range redises {
			if !reflect.DeepEqual(masterPod, redisPod) {
				Info(log, "starting set redis slave "+redisPod.Name+" new password...", el.Redis)
				if err := r.RedisHandler.Healer.SetRedisPassword(redisPod, el.Redis); err != nil {
					return err
				}
			}
		}
	} else {
		return errors.New("starting set redis master new password error")
	}

	return nil
}

func (r *RedisReconciler) checkAndHealSentinelPassword(el element.Element) (element.Element, error) {
	log := r.Log.WithValues("controller", "checkAndHealSentinelPassword")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	previousStatus := el.Redis.Status.Sentinel.SentinelPassword
	currentStatus := *el.Redis.Status.Sentinel.SentinelPassword.DeepCopy()
	password, err := k8s.GetSpecRedisPassword(r.RedisHandler.K8sServices, el.Redis)
	if err != nil {
		return el, err
	}
	currentStatus.Md5 = util.MD5(password)

	if !reflect.DeepEqual(previousStatus, currentStatus) {
		el.NeedReCheckError = append(el.NeedReCheckError, errors.New("SentinelPassword Status not equal"))
		Info(log, "SentinelPassword Status not equal", el.Redis)
		if err = r.applySentinelPassword(el); err != nil {
			return el, err
		}

		if err = r.RedisHandler.Healer.UpdateSentinelPasswordStatus(el.Redis, currentStatus); err != nil {
			return el, err
		}
		el.NeedReLoad = true
	} else {
		Info(log, "SentinelPassword Status equal", el.Redis)
	}
	return el, nil
}

func (r *RedisReconciler) applySentinelPassword(el element.Element) error {
	log := r.Log.WithValues("controller", "applySentinelPassword")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	redises, err := r.RedisHandler.Checker.GetSentinelsPods(el)
	if err != nil {
		return err
	}
	for _, rip := range redises {
		Info(log, "starting set sentinel "+rip.Name+" new password...", el.Redis)
		if err := r.RedisHandler.Healer.SetSentinelPassword(rip, el.Redis); err != nil {
			return err
		}
	}
	return nil
}

// --- needCheckAndHealPassword ---
func (r *RedisReconciler) needCheckAndHealPassword(el element.Element) (element.Element, error, bool) {
	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err, false
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	el, err, need := r.needCheckAndHealRedisPassword(el)
	if err != nil {
		return el, err, need
	}
	if need {
		return el, nil, true
	}

	el, err, need = r.needCheckAndHealSentinelPassword(el)
	if err != nil {
		return el, err, need
	}
	if need {
		return el, nil, true
	}

	return el, nil, false
}

func (r *RedisReconciler) needCheckAndHealRedisPassword(el element.Element) (element.Element, error, bool) {
	log := r.Log.WithValues("controller", "needCheckAndHealRedisPassword")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err, false
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	previousStatus := el.Redis.Status.Redis.RedisPassword
	currentStatus := *el.Redis.Status.Redis.RedisPassword.DeepCopy()

	password, err := k8s.GetSpecRedisPassword(r.RedisHandler.K8sServices, el.Redis)
	if err != nil {
		return el, err, false
	}
	currentStatus.Md5 = util.MD5(password)

	if !reflect.DeepEqual(previousStatus, currentStatus) {
		Info(log, "need check and heal redis password", el.Redis)
		return el, nil, true
	} else {
		Info(log, "no need to check and heal redis password", el.Redis)
		return el, nil, false
	}
}

func (r *RedisReconciler) needCheckAndHealSentinelPassword(el element.Element) (element.Element, error, bool) {
	log := r.Log.WithValues("controller", "needCheckAndHealSentinelPassword")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err, false
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	previousStatus := el.Redis.Status.Sentinel.SentinelPassword
	currentStatus := *el.Redis.Status.Sentinel.SentinelPassword.DeepCopy()
	password, err := k8s.GetSpecRedisPassword(r.RedisHandler.K8sServices, el.Redis)
	if err != nil {
		return el, err, false
	}
	currentStatus.Md5 = util.MD5(password)

	if !reflect.DeepEqual(previousStatus, currentStatus) {
		Info(log, "need check and heal sentinel password", el.Redis)
		return el, nil, true
	} else {
		Info(log, "no need to check and heal sentinel password", el.Redis)
		return el, nil, false
	}
}

// --- CheckCluster ---
func (r *RedisReconciler) CheckCluster(el element.Element, init bool) (element.Element, error) {
	log := r.Log.WithValues("controller", "CheckCluster")

	if el.NeedReLoad {
		redisNew, err := r.RedisHandler.K8sServices.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	previousStatus := el.Redis.Status.State
	currentStatus := *el.Redis.Status.State.DeepCopy()

	currentStatus.Cluster = init

	if !reflect.DeepEqual(previousStatus, currentStatus) {
		Info(log, "State Cluster Status not equal", el.Redis)
		if err := r.RedisHandler.Healer.UpdatePodStatus(el.Redis, currentStatus); err != nil {
			return el, err
		}
		el.NeedReLoad = true
	} else {
		Info(log, "State Cluster Status equal", el.Redis)
	}

	return el, nil
}
