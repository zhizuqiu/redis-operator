package ensure

import (
	"context"
	"github.com/go-logr/logr"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	"github.com/zhizuqiu/redis-operator/controllers/service/element"
	"github.com/zhizuqiu/redis-operator/controllers/service/k8s"
	"github.com/zhizuqiu/redis-operator/controllers/util"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"reflect"
)

type RedisEnsure interface {
	EnsureSentinelConfigMaps(el element.Element) (element.Element, error)
	EnsureRedisReadinessConfigMap(el element.Element) (element.Element, error)
	EnsureRedisMasterConfigMap(el element.Element) (element.Element, error)
	EnsureRedisSlaveConfigMaps(el element.Element) (element.Element, error)
	EnsureRedisStatefulSets(el element.Element) (element.Element, error)
	EnsureSentinelStatefulSets(el element.Element) (element.Element, error)
	EnsureSentinelService(el element.Element) (element.Element, error)
	EnsureExporterDeployment(el element.Element) (element.Element, error)
	EnsureSentinelHeadlessService(el element.Element) (element.Element, error)
	EnsureRedisHeadlessService(el element.Element) (element.Element, error)
}

type RedisEnsurer struct {
	K8SService k8s.Services
	Log        logr.Logger
}

func NewRedisEnsurer(k8sService k8s.Services, log logr.Logger) *RedisEnsurer {
	log = log.WithValues("ensure", "RedisEnsurer")
	return &RedisEnsurer{
		K8SService: k8sService,
		Log:        log,
	}
}

// --- EnsureSentinelConfigMap ---
func (r *RedisEnsurer) EnsureSentinelConfigMaps(el element.Element) (element.Element, error) {

	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	err := util.NilError()
	for i := 0; i < int(el.Redis.Spec.Sentinel.Replicas); i++ {
		el, err = r.ensureSentinelConfigMap(el, i)
		if err != nil {
			return el, err
		}
	}

	return el, nil
}

func (r *RedisEnsurer) ensureSentinelConfigMap(el element.Element, index int) (element.Element, error) {
	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	password, err := k8s.GetSpecRedisPassword(r.K8SService, el.Redis)
	if err != nil {
		return el, err
	}

	currentSentinelConfigMapStatus := roav1.RedisStatusItem{}
	previousSentinelStatus := el.Redis.Status.Sentinel
	currentSentinelStatus := *el.Redis.Status.Sentinel.DeepCopy()

	exists := true
	sentinelConfigMap, err := r.K8SService.GetConfigMap(el.Redis.Namespace, util.GetSentinelConfigMapNameByIndex(el.Redis, index))
	if err != nil {
		if errors.IsNotFound(err) {
			exists = false
		} else {
			return el, err
		}
	}

	PrintOBJ("get SentinelSlaveConfigMap", el.Redis, sentinelConfigMap)

	var desiredSentinelConfigMap = &v1.ConfigMap{}
	if exists {
		existingSentinelConfigMap := sentinelConfigMap
		desiredSentinelConfigMap = util.CreateSentinelConfigMapObjByExistingObjByIndex(el.Redis, password, existingSentinelConfigMap.DeepCopy(), index)

		currentSentinelConfigMapStatus.Status = roav1.Desired

		if reflect.DeepEqual(desiredSentinelConfigMap.Data, existingSentinelConfigMap.Data) {
			Info(r.Log, "SentinelSlaveConfigMap Data equal", el.Redis)
			currentSentinelConfigMapStatus.Status = roav1.Desired
		} else {
			Info(r.Log, "SentinelSlaveConfigMap Data not equal", el.Redis)
			currentSentinelConfigMapStatus.Status = roav1.Pending
		}

	} else {
		currentSentinelConfigMapStatus.Status = ""
		Info(r.Log, "SentinelState.SentinelConfigMap.Status=\"\" set SentinelState.SentinelPassword.Md5", el.Redis)
		currentSentinelStatus.SentinelPassword.Md5 = util.MD5(password)
	}

	if !reflect.DeepEqual(previousSentinelStatus, currentSentinelStatus) {
		PrintOBJ("currentSentinelStatus", el.Redis, currentSentinelStatus)
		PrintOBJ("previousSentinelStatus", el.Redis, previousSentinelStatus)

		Info(r.Log, "SentinelState Status not equal", el.Redis)
		if err := r.K8SService.UpdateSentinelState(el.Redis, currentSentinelStatus); err != nil {
			return el, err
		}
		el.NeedReLoad = true
	} else {
		Info(r.Log, "SentinelState Status equal", el.Redis)
	}

	if currentSentinelConfigMapStatus.Status == roav1.Desired {
		return el, nil
	} else if currentSentinelConfigMapStatus.Status == roav1.Pending {
		Info(r.Log, "start update SentinelSlaveConfigMap...", el.Redis)

		if err := r.K8SService.Update(context.Background(), desiredSentinelConfigMap); err != nil {
			return el, err
		}
	} else {
		configMap := util.CreateSentinelConfigMapByIndex(el.Redis, el.OwnerRefs, password, index)
		PrintOBJ("create SentinelSlaveConfigMap object", el.Redis, configMap)

		if err := r.K8SService.Create(context.Background(), configMap); err != nil {
			return el, err
		}
	}

	return el, nil
}

// --- EnsureRedisReadinessConfigMap ---
func (r *RedisEnsurer) EnsureRedisReadinessConfigMap(el element.Element) (element.Element, error) {

	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	currentReadinessConfigMapStatus := roav1.RedisStatusItem{}

	exists := true
	readinessConfigMap, err := r.K8SService.GetConfigMap(el.Redis.Namespace, util.GetRedisReadinessConfigMapName(el.Redis))
	if err != nil {
		if errors.IsNotFound(err) {
			exists = false
		} else {
			return el, err
		}
	}

	PrintOBJ("get RedisReadinessConfigMap", el.Redis, readinessConfigMap)

	if exists {
		currentReadinessConfigMapStatus.Status = roav1.Desired
	} else {
		currentReadinessConfigMapStatus.Status = roav1.Pending
	}

	if currentReadinessConfigMapStatus.Status == roav1.Desired {
		return el, nil
	} else {
		configMap := util.CreateReadinessConfigMap(el.Redis, el.OwnerRefs)

		PrintOBJ("create readinessConfigMap object", el.Redis, configMap)

		// ...and create it on the cluster
		if err := r.K8SService.Create(context.Background(), configMap); err != nil {
			return el, err
		}
	}

	return el, nil
}

// --- EnsureRedisMasterConfigMap ---
func (r *RedisEnsurer) EnsureRedisMasterConfigMap(el element.Element) (element.Element, error) {
	log := r.Log.WithValues("RedisEnsurer", "EnsureRedisMasterConfigMap")

	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	password, err := k8s.GetSpecRedisPassword(r.K8SService, el.Redis)
	if err != nil {
		return el, err
	}

	currentRedisConfigMapStatus := roav1.RedisStatusItem{}
	previousRedisStatus := el.Redis.Status.Redis
	currentRedisStatus := *el.Redis.Status.Redis.DeepCopy()

	exists := true
	redisConfigMap, err := r.K8SService.GetConfigMap(el.Redis.Namespace, util.GetRedisConfigMapNameByIndex(el.Redis, 0))
	if err != nil {
		if errors.IsNotFound(err) {
			exists = false
		} else {
			return el, err
		}
	}

	PrintOBJ("get RedisMasterConfigMap", el.Redis, redisConfigMap)

	var desiredRedisConfigMap = &v1.ConfigMap{}
	if exists {
		existingRedisConfigMap := redisConfigMap
		desiredRedisConfigMap = util.CreateRedisMasterConfigMapObjByExistingObj(el.Redis, password, existingRedisConfigMap.DeepCopy())

		currentRedisConfigMapStatus.Status = roav1.Desired

		if reflect.DeepEqual(desiredRedisConfigMap.Data, existingRedisConfigMap.Data) {
			Info(r.Log, "RedisMasterConfigMap Data equal", el.Redis)
			currentRedisConfigMapStatus.Status = roav1.Desired
		} else {
			Info(r.Log, "RedisMasterConfigMap Data not equal", el.Redis)
			currentRedisConfigMapStatus.Status = roav1.Pending
		}

	} else {
		currentRedisConfigMapStatus.Status = ""
		Info(r.Log, "RedisState.RedisPassword.Status=\"\" set RedisState.RedisPassword.Md5", el.Redis)
		currentRedisStatus.RedisPassword.Md5 = util.MD5(password)
	}

	if !reflect.DeepEqual(previousRedisStatus, currentRedisStatus) {
		PrintOBJ("currentRedisStatus", el.Redis, currentRedisStatus)
		PrintOBJ("previousRedisStatus", el.Redis, previousRedisStatus)

		Info(log, "RedisState Status not equal", el.Redis)
		if err := r.K8SService.UpdateRedisState(el.Redis, currentRedisStatus); err != nil {
			return el, err
		}
		el.NeedReLoad = true
	} else {
		Info(log, "RedisState Status equal", el.Redis)
	}

	if currentRedisConfigMapStatus.Status == roav1.Desired {
		return el, nil
	} else if currentRedisConfigMapStatus.Status == roav1.Pending {
		Info(r.Log, "start update RedisMasterConfigMap...", el.Redis)

		if err := r.K8SService.Update(context.Background(), desiredRedisConfigMap); err != nil {
			return el, err
		}
	} else {
		configMap := util.CreateRedisMasterConfigMap(el.Redis, el.OwnerRefs, password)
		PrintOBJ("create RedisMasterConfigMap object", el.Redis, configMap)

		if err := r.K8SService.Create(context.Background(), configMap); err != nil {
			return el, err
		}
	}

	return el, nil
}

// --- EnsureRedisSlaveConfigMap ---
func (r *RedisEnsurer) EnsureRedisSlaveConfigMaps(el element.Element) (element.Element, error) {
	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	err := util.NilError()
	for i := 0; i < int(el.Redis.Spec.Redis.Replicas); i++ {
		if i != 0 {
			el, err = r.ensureRedisSlaveConfigMap(el, i)
			if err != nil {
				return el, err
			}
		}
	}

	return el, nil
}

func (r *RedisEnsurer) ensureRedisSlaveConfigMap(el element.Element, index int) (element.Element, error) {
	log := r.Log.WithValues("RedisEnsurer", "EnsureAutoFailoverRedisConfigMap")

	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	password, err := k8s.GetSpecRedisPassword(r.K8SService, el.Redis)
	if err != nil {
		return el, err
	}

	currentRedisConfigMapStatus := roav1.RedisStatusItem{}
	previousRedisStatus := el.Redis.Status.Redis
	currentRedisStatus := *el.Redis.Status.Redis.DeepCopy()

	exists := true
	redisConfigMap, err := r.K8SService.GetConfigMap(el.Redis.Namespace, util.GetRedisConfigMapNameByIndex(el.Redis, index))
	if err != nil {
		if errors.IsNotFound(err) {
			exists = false
		} else {
			return el, err
		}
	}

	PrintOBJ("get RedisSlaveConfigMap", el.Redis, redisConfigMap)

	var desiredRedisConfigMap = &v1.ConfigMap{}
	if exists {
		existingRedisConfigMap := redisConfigMap
		desiredRedisConfigMap = util.CreateRedisSlaveConfigMapObjByExistingObjByIndex(el.Redis, password, existingRedisConfigMap.DeepCopy(), index)

		currentRedisConfigMapStatus.Status = roav1.Desired

		if reflect.DeepEqual(desiredRedisConfigMap.Data, existingRedisConfigMap.Data) {
			Info(r.Log, "RedisSlaveConfigMap Data equal", el.Redis)
			currentRedisConfigMapStatus.Status = roav1.Desired
		} else {
			Info(r.Log, "RedisSlaveConfigMap Data not equal", el.Redis)
			currentRedisConfigMapStatus.Status = roav1.Pending
		}

	} else {
		currentRedisConfigMapStatus.Status = ""
		Info(r.Log, "RedisState.RedisPassword.Status=\"\" set RedisState.RedisPassword.Md5", el.Redis)
		currentRedisStatus.RedisPassword.Md5 = util.MD5(password)
	}

	if !reflect.DeepEqual(previousRedisStatus, currentRedisStatus) {
		PrintOBJ("currentRedisStatus", el.Redis, currentRedisStatus)
		PrintOBJ("previousRedisStatus", el.Redis, previousRedisStatus)

		Info(log, "RedisState Status not equal", el.Redis)
		if err := r.K8SService.UpdateRedisState(el.Redis, currentRedisStatus); err != nil {
			return el, err
		}
		el.NeedReLoad = true
	} else {
		Info(log, "RedisState Status equal", el.Redis)
	}

	if currentRedisConfigMapStatus.Status == roav1.Desired {
		return el, nil
	} else if currentRedisConfigMapStatus.Status == roav1.Pending {
		Info(r.Log, "start update RedisSlaveConfigMap...", el.Redis)

		if err := r.K8SService.Update(context.Background(), desiredRedisConfigMap); err != nil {
			return el, err
		}
	} else {
		configMap := util.CreateRedisSlaveConfigMapByIndex(el.Redis, el.OwnerRefs, password, index)
		PrintOBJ("create RedisSlaveConfigMap object", el.Redis, configMap)

		if err := r.K8SService.Create(context.Background(), configMap); err != nil {
			return el, err
		}
	}

	return el, nil
}

// --- EnsureRedisStatefulSets ---
func (r *RedisEnsurer) EnsureRedisStatefulSets(el element.Element) (element.Element, error) {

	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	err := util.NilError()
	for i := 0; i < int(el.Redis.Spec.Redis.Replicas); i++ {
		el, err = r.ensureRedisStatefulSet(el, i)
		if err != nil {
			return el, err
		}
	}

	return el, nil
}

func (r *RedisEnsurer) ensureRedisStatefulSet(el element.Element, index int) (element.Element, error) {
	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	statefulSetName := util.GetRedisNameByIndex(el.Redis, index)

	currentRedisStatefulSetStatus := roav1.RedisStatusItem{}

	exists := true
	statefulSet, err := r.K8SService.GetStatefulSet(el.Redis.Namespace, statefulSetName)
	if err != nil {
		if errors.IsNotFound(err) {
			exists = false
		} else {
			return el, err
		}
	}

	PrintOBJ("get RedisStatefulSet", el.Redis, statefulSet)

	var desiredRedisStatefulSet = &appsv1.StatefulSet{}
	if exists {
		existingRedisStatefulSet := statefulSet
		desiredRedisStatefulSet = util.CreateRedisStatefulSetObjByExistingObjByIndex(el.Redis, el.OwnerRefs, existingRedisStatefulSet.DeepCopy(), index)

		PrintOBJ("desiredRedisStatefulSet", el.Redis, desiredRedisStatefulSet.Spec)
		PrintOBJ("existingRedisStatefulSet", el.Redis, existingRedisStatefulSet.Spec)

		currentRedisStatefulSetStatus.Status = roav1.Desired

		DealResource(&desiredRedisStatefulSet.Spec.Template.Spec)
		DealResource(&existingRedisStatefulSet.Spec.Template.Spec)

		if util.RedisStatefulSetEqual(desiredRedisStatefulSet, existingRedisStatefulSet) {
			Info(r.Log, "RedisStatefulSet Spec equal", el.Redis)
			currentRedisStatefulSetStatus.Status = roav1.Desired
		} else {
			Info(r.Log, "RedisStatefulSet Spec not equal", el.Redis)
			currentRedisStatefulSetStatus.Status = roav1.Pending
		}
	} else {
		currentRedisStatefulSetStatus.Status = ""
	}

	if currentRedisStatefulSetStatus.Status == roav1.Desired {
		return el, nil
	} else if currentRedisStatefulSetStatus.Status == roav1.Pending {

		Info(r.Log, "start update RedisStatefulSet...", el.Redis)
		// ...and Update it on the cluster
		if err := r.K8SService.Update(context.Background(), desiredRedisStatefulSet); err != nil {
			return el, err
		}
	} else {
		statefulSet := util.CreateRedisStatefulSetObjByIndex(el.Redis, el.OwnerRefs, index)

		PrintOBJ("create RedisStatefulSet object", el.Redis, statefulSet)

		// ...and create it on the cluster
		if err := r.K8SService.Create(context.Background(), statefulSet); err != nil {
			return el, err
		}
	}

	return el, nil
}

// --- EnsureSentinelStatefulSets ---
func (r *RedisEnsurer) EnsureSentinelStatefulSets(el element.Element) (element.Element, error) {

	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	err := util.NilError()
	for i := 0; i < int(el.Redis.Spec.Sentinel.Replicas); i++ {
		el, err = r.ensureSentinelStatefulSet(el, i)
		if err != nil {
			return el, err
		}
	}

	return el, nil
}

func (r *RedisEnsurer) ensureSentinelStatefulSet(el element.Element, index int) (element.Element, error) {
	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	statefulSetName := util.GetSentinelNameByIndex(el.Redis, index)

	currentSentinelStatefulSetStatus := roav1.RedisStatusItem{}

	exists := true
	statefulSet, err := r.K8SService.GetStatefulSet(el.Redis.Namespace, statefulSetName)
	if err != nil {
		if errors.IsNotFound(err) {
			exists = false
		} else {
			return el, err
		}
	}

	PrintOBJ("get SentinelStatefulSet", el.Redis, statefulSet)

	var desiredSentinelStatefulSet = &appsv1.StatefulSet{}
	if exists {
		existingSentinelStatefulSet := statefulSet
		desiredSentinelStatefulSet = util.CreateSentinelStatefulSetObjByExistingObjByIndex(el.Redis, el.OwnerRefs, existingSentinelStatefulSet.DeepCopy(), index)

		PrintOBJ("desiredSentinelStatefulSet", el.Redis, desiredSentinelStatefulSet.Spec)
		PrintOBJ("existingSentinelStatefulSet", el.Redis, existingSentinelStatefulSet.Spec)

		currentSentinelStatefulSetStatus.Status = roav1.Desired

		DealResource(&desiredSentinelStatefulSet.Spec.Template.Spec)
		DealResource(&existingSentinelStatefulSet.Spec.Template.Spec)
		if util.SentinelStatefulSetEqual(desiredSentinelStatefulSet, existingSentinelStatefulSet) {
			Info(r.Log, "SentinelStatefulSet Spec equal", el.Redis)
			currentSentinelStatefulSetStatus.Status = roav1.Desired
		} else {
			Info(r.Log, "SentinelStatefulSet Spec not equal", el.Redis)
			currentSentinelStatefulSetStatus.Status = roav1.Pending
		}
	} else {
		currentSentinelStatefulSetStatus.Status = ""
	}

	if currentSentinelStatefulSetStatus.Status == roav1.Desired {
		return el, nil
	} else if currentSentinelStatefulSetStatus.Status == roav1.Pending {

		Info(r.Log, "start update SentinelStatefulSet...", el.Redis)
		// ...and Update it on the cluster
		if err := r.K8SService.Update(context.Background(), desiredSentinelStatefulSet); err != nil {
			return el, err
		}
	} else {
		statefulSet := util.CreateSentinelStatefulSetObjByIndex(el.Redis, el.OwnerRefs, index)
		PrintOBJ("create SentinelStatefulSet object", el.Redis, statefulSet)

		// ...and create it on the cluster
		if err := r.K8SService.Create(context.Background(), statefulSet); err != nil {
			return el, err
		}
	}

	return el, nil
}

// --- EnsureSentinelService ---
func (r *RedisEnsurer) EnsureSentinelService(el element.Element) (element.Element, error) {
	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	currentSentinelStatus := roav1.RedisStatusItem{}

	exists := true
	sentinelService, err := r.K8SService.GetService(el.Redis.Namespace, util.GetSentinelRootName(el.Redis))
	if err != nil {
		if errors.IsNotFound(err) {
			exists = false
		} else {
			return el, err
		}
	}

	PrintOBJ("get SentinelService", el.Redis, sentinelService)

	if exists {
		currentSentinelStatus.Status = roav1.Desired
	} else {
		currentSentinelStatus.Status = roav1.Pending
	}

	if currentSentinelStatus.Status == roav1.Desired {
		return el, nil
	} else {

		service := util.CreateSentinelService(el.Redis, el.OwnerRefs)

		PrintOBJ("create sentinelService object", el.Redis, service)

		// ...and create it on the cluster
		if err := r.K8SService.Create(context.Background(), service); err != nil {
			return el, err
		}
	}

	return el, nil
}

// --- EnsureExporterDeployment ---
func (r *RedisEnsurer) EnsureExporterDeployment(el element.Element) (element.Element, error) {
	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	currentExporterStatus := roav1.RedisStatusItem{}

	exists := true
	exporterDeployment, err := r.K8SService.GetDeployment(el.Redis.Namespace, util.GetExporterRootName(el.Redis))
	if err != nil {
		if errors.IsNotFound(err) {
			exists = false
		} else {
			return el, err
		}
	}

	PrintOBJ("get ExporterDeployment", el.Redis, exporterDeployment)

	password, err := k8s.GetSpecRedisPassword(r.K8SService, el.Redis)
	if err != nil {
		return el, err
	}

	var desiredExporterDeployment = &appsv1.Deployment{}
	if exists {

		existingExporterDeployment := exporterDeployment
		desiredExporterDeployment = util.CreateExporterDeploymentObjByExistingObj(el.Redis, password, existingExporterDeployment.DeepCopy())

		PrintOBJ("desiredExporterDeployment", el.Redis, desiredExporterDeployment.Spec)
		PrintOBJ("existingExporterDeployment", el.Redis, existingExporterDeployment.Spec)
		currentExporterStatus.Status = roav1.Desired
		if util.ExporterDeploymentEqual(desiredExporterDeployment, existingExporterDeployment) {
			Info(r.Log, "ExporterDeployment Spec equal", el.Redis)
			currentExporterStatus.Status = roav1.Desired
		} else {
			Info(r.Log, "ExporterDeployment Spec not equal", el.Redis)
			currentExporterStatus.Status = roav1.Pending
		}
	} else {
		currentExporterStatus.Status = ""
	}

	if currentExporterStatus.Status == roav1.Desired {
		return el, nil
	} else if currentExporterStatus.Status == roav1.Pending {
		Info(r.Log, "start update ExporterDeployment...", el.Redis)
		// ...and Update it on the cluster
		if err := r.K8SService.Update(context.Background(), desiredExporterDeployment); err != nil {
			return el, err
		}
	} else {
		service := util.CreateExporterDeployment(el.Redis, el.OwnerRefs, password)

		PrintOBJ("create exporterDeployment object", el.Redis, service)

		// ...and create it on the cluster
		if err := r.K8SService.Create(context.Background(), service); err != nil {
			return el, err
		}
	}

	return el, nil
}

// DealResource 将pod的资源量的单位统一
func DealResource(requirements *v1.PodSpec) {
	for _, container := range requirements.Containers {
		oldResource := container.Resources
		for key, value := range oldResource.Limits {
			oldResource.Limits[key] = resource.MustParse(value.String())
		}
		for key, value := range oldResource.Requests {
			oldResource.Requests[key] = resource.MustParse(value.String())
		}
	}
}

// --- EnsureSentinelHeadlessService ---
func (r *RedisEnsurer) EnsureSentinelHeadlessService(el element.Element) (element.Element, error) {
	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	err := util.NilError()
	for i := 0; i < int(el.Redis.Spec.Sentinel.Replicas); i++ {
		el, err = r.ensureSentinelHeadlessService(el, i)
		if err != nil {
			return el, err
		}
	}

	return el, nil

}

func (r *RedisEnsurer) ensureSentinelHeadlessService(el element.Element, index int) (element.Element, error) {
	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	headlessServiceName := util.GetSentinelHeadlessServiceNameByIndex(el.Redis, index)

	currentSentinelHeadlessStatus := roav1.RedisStatusItem{}

	exists := true
	service, err := r.K8SService.GetService(el.Redis.Namespace, headlessServiceName)
	if err != nil {
		if errors.IsNotFound(err) {
			exists = false
		} else {
			return el, err
		}
	}

	PrintOBJ("get SentinelHeadlessService", el.Redis, service)

	if exists {
		currentSentinelHeadlessStatus.Status = roav1.Desired
	} else {
		currentSentinelHeadlessStatus.Status = roav1.Pending
	}

	if currentSentinelHeadlessStatus.Status == roav1.Desired {
		return el, nil
	} else {

		service := util.CreateSentinelHeadlessServiceByIndex(el.Redis, el.OwnerRefs, index)

		PrintOBJ("create SentinelHeadlessService object", el.Redis, service)

		// ...and create it on the cluster
		if err := r.K8SService.Create(context.Background(), service); err != nil {
			return el, err
		}
	}

	return el, nil
}

// --- EnsureRedisHeadlessService ---
func (r *RedisEnsurer) EnsureRedisHeadlessService(el element.Element) (element.Element, error) {
	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	err := util.NilError()
	for i := 0; i < int(el.Redis.Spec.Redis.Replicas); i++ {
		el, err = r.ensureRedisHeadlessService(el, i)
		if err != nil {
			return el, err
		}
	}

	return el, nil
}

func (r *RedisEnsurer) ensureRedisHeadlessService(el element.Element, index int) (element.Element, error) {
	if el.NeedReLoad {
		redisNew, err := r.K8SService.Get(el.Req)
		if err != nil {
			return el, err
		}
		el.Redis = redisNew
	}
	el.NeedReLoad = false

	headlessServiceName := util.GetRedisHeadlessServiceNameByIndex(el.Redis, index)

	currentRedisHeadlessStatus := roav1.RedisStatusItem{}

	exists := true
	service, err := r.K8SService.GetService(el.Redis.Namespace, headlessServiceName)
	if err != nil {
		if errors.IsNotFound(err) {
			exists = false
		} else {
			return el, err
		}
	}

	PrintOBJ("get RedisHeadlessService", el.Redis, service)

	if exists {
		currentRedisHeadlessStatus.Status = roav1.Desired
	} else {
		currentRedisHeadlessStatus.Status = roav1.Pending
	}

	if currentRedisHeadlessStatus.Status == roav1.Desired {
		return el, nil
	} else {

		service := util.CreateRedisHeadlessServiceByIndex(el.Redis, el.OwnerRefs, index)

		PrintOBJ("create RedisHeadlessService object", el.Redis, service)

		// ...and create it on the cluster
		if err := r.K8SService.Create(context.Background(), service); err != nil {
			return el, err
		}
	}

	return el, nil
}
