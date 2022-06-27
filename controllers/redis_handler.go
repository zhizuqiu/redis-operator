package controllers

import (
	"github.com/go-logr/logr"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	"github.com/zhizuqiu/redis-operator/controllers/service/check"
	"github.com/zhizuqiu/redis-operator/controllers/service/ensure"
	"github.com/zhizuqiu/redis-operator/controllers/service/k8s"
	"github.com/zhizuqiu/redis-operator/controllers/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

const (
	rfLabelManagedByKey = "app.kubernetes.io/managed-by"
)

const (
	resync       = 30 * time.Second
	operatorName = "redis-operator"
)

var (
	defaultLabels = map[string]string{
		rfLabelManagedByKey: operatorName,
	}
)

func Info(log logr.Logger, msg string, rf *roav1.Redis) {
	log.Info(msg, "nameSpace", rf.Namespace, "name", rf.Name)
}

func Info2(log logr.Logger, msg string, req ctrl.Request) {
	log.Info(msg, "nameSpace", req.Namespace, "name", req.Name)
}

func Error(log logr.Logger, err error, msg string, rf *roav1.Redis) {
	log.Error(err, msg, "nameSpace", rf.Namespace, "name", rf.Name)
}

type RedisHandler struct {
	Ensurer       ensure.RedisEnsure
	DeleteEnsurer ensure.RedisDeleteEnsure
	Checker       check.RedisCheck
	DeleteChecker check.RedisDeleteChecke
	Healer        check.RedisHeal
	K8sServices   k8s.Services
	Log           logr.Logger
}

func NewRedisHandler(ensurer ensure.RedisEnsure, deleteEnsurer ensure.RedisDeleteEnsure, checker check.RedisCheck, deleteChecker check.RedisDeleteChecke, healer check.RedisHeal, k8sServices k8s.Services, log logr.Logger) *RedisHandler {
	return &RedisHandler{
		Ensurer:       ensurer,
		DeleteEnsurer: deleteEnsurer,
		Checker:       checker,
		DeleteChecker: deleteChecker,
		Healer:        healer,
		K8sServices:   k8sServices,
		Log:           log,
	}
}

func (r *RedisHandler) createOwnerReferences(rf *roav1.Redis) []metav1.OwnerReference {
	rfvk := schema.GroupVersionKind{
		Group:   roav1.GroupVersion.Group,
		Version: roav1.GroupVersion.Version,
		Kind:    "Redis",
	}
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(rf, rfvk),
	}
}

func (r *RedisHandler) getPodStates(podList *v1.PodList) map[string]roav1.PodState {
	podStates := make(map[string]roav1.PodState)

	if podList == nil {
		return podStates
	}
	for _, item := range podList.Items {
		podStates[item.Name] = roav1.PodState{
			Name:          item.Name,
			Role:          util.GetRoleFromLabel(item),
			Phase:         item.Status.Phase,
			HostIP:        item.Status.HostIP,
			PodIP:         item.Status.PodIP,
			ContainerPort: util.GetPort(item),
			PodIPs:        item.Status.PodIPs,
			StartTime:     item.Status.StartTime,
		}
	}
	return podStates
}
