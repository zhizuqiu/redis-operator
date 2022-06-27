package util

import (
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
)

func CreateSentinelService(rf *roav1.Redis, ownerRefs []metav1.OwnerReference) *corev1.Service {
	name := GetSentinelServiceName(rf)
	namespace := rf.Namespace

	sentinelTargetPort := intstr.FromInt(26379)
	labels := GetSentinelServiceLabels(rf)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
			Annotations:     rf.Spec.Sentinel.Service.ServiceAnnotations,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       sentinelName,
					Port:       sentinelContainerPort,
					TargetPort: sentinelTargetPort,
					Protocol:   "TCP",
				},
			},
		},
	}
}

func GetSentinelServiceLabels(rf *roav1.Redis) map[string]string {
	return GenerateSelectorLabels(sentinelRootName, rf)
}

func GetRedisServiceLabels(rf *roav1.Redis) map[string]string {
	return GenerateSelectorLabels(redisRootName, rf)
}

func GetSentinelHeadlessServiceSelectorByIndex(rf *roav1.Redis, index int) map[string]string {
	return map[string]string{
		statefulSetPodLabelKey: GetSentinelNameByIndex(rf, index) + "-0",
	}
}

func GetRedisHeadlessServiceSelectorByIndex(rf *roav1.Redis, index int) map[string]string {
	return map[string]string{
		statefulSetPodLabelKey: GetRedisNameByIndex(rf, index) + "-0",
	}
}

func GetSentinelServiceName(rf *roav1.Redis) string {
	return GetSentinelRootName(rf)
}

func GetSentinelHeadlessServiceNameByIndex(rf *roav1.Redis, index int) string {
	return headlessServiceBaseName + "-" + GetSentinelNameByIndex(rf, index)
}

func GetRedisHeadlessServiceNameByIndex(rf *roav1.Redis, index int) string {
	return headlessServiceBaseName + "-" + GetRedisNameByIndex(rf, index)
}

func GetRedisHostByIndex(rf *roav1.Redis, index int) string {
	host := GetRedisHeadlessServiceNameByIndex(rf, index)
	if len(rf.Spec.Redis.StaticResources) > index {
		host = rf.Spec.Redis.StaticResources[index].Host
	}
	return host
}

func GetSentinelHostByIndex(rf *roav1.Redis, index int) string {
	host := GetSentinelHeadlessServiceNameByIndex(rf, index)
	if len(rf.Spec.Sentinel.StaticResources) > index {
		host = rf.Spec.Sentinel.StaticResources[index].Host
	}
	return host
}

func CreateSentinelHeadlessServiceByIndex(rf *roav1.Redis, ownerRefs []metav1.OwnerReference, index int) *corev1.Service {
	name := GetSentinelHeadlessServiceNameByIndex(rf, index)
	namespace := rf.Namespace

	port, _ := strconv.Atoi(GetSentinelPortFromSpecByIndex(rf, index))
	sentinelTargetPort := intstr.FromInt(port)
	labels := GetSentinelServiceLabels(rf)
	selector := GetSentinelHeadlessServiceSelectorByIndex(rf, index)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Spec: corev1.ServiceSpec{
			Selector:  selector,
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
			Ports: []corev1.ServicePort{
				{
					Name:       sentinelName,
					Port:       int32(port),
					TargetPort: sentinelTargetPort,
					Protocol:   "TCP",
				},
			},
		},
	}
}

func CreateRedisHeadlessServiceByIndex(rf *roav1.Redis, ownerRefs []metav1.OwnerReference, index int) *corev1.Service {
	name := GetRedisHeadlessServiceNameByIndex(rf, index)
	namespace := rf.Namespace

	port, _ := strconv.Atoi(GetRedisPortFromSpecByIndex(rf, index))
	redisTargetPort := intstr.FromInt(port)
	labels := GetRedisServiceLabels(rf)
	selector := GetRedisHeadlessServiceSelectorByIndex(rf, index)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Spec: corev1.ServiceSpec{
			Selector:  selector,
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
			Ports: []corev1.ServicePort{
				{
					Name:       redisName,
					Port:       int32(port),
					TargetPort: redisTargetPort,
					Protocol:   "TCP",
				},
			},
		},
	}
}
