package util

import (
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"strconv"
)

func CreateRedisStatefulSetObjByIndex(rf *roav1.Redis, ownerRefs []metav1.OwnerReference, index int) *v1.StatefulSet {
	name := GetRedisNameByIndex(rf, index)
	namespace := rf.Namespace

	redisCommand := getRedisCommand(rf)
	selector := GetRedisLabels(rf)
	labels := GetRedisLabelsWithName(rf, name)
	volumeMounts := getRedisVolumeMounts(rf)
	volumes := getRedisVolumesByIndex(rf, index)
	annotations := getRedisAnnotations(rf)
	nodeAffinity := getRedisNodeAffinityByIndex(rf, index)
	affinity := getNodeAndPodAffinity(rf.Spec.Redis.Affinity, rf.Spec.Redis.EnabledPodAntiAffinity, selector, nodeAffinity)
	port := GetRedisPortFromSpecByIndex(rf, index)
	portInt, _ := strconv.Atoi(port)
	livenessProbeCommand := "redis-cli -p " + port + " -h $(hostname) ping"

	ss := &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Spec: v1.StatefulSetSpec{
			Replicas:            Int32P(1),
			UpdateStrategy:      getRedisUpdateStrategy(rf),
			PodManagementPolicy: v1.ParallelPodManagement,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Affinity:          affinity,
					Tolerations:       rf.Spec.Redis.Tolerations,
					NodeSelector:      rf.Spec.Redis.NodeSelector,
					SecurityContext:   getSecurityContext(rf.Spec.Redis.SecurityContext),
					HostNetwork:       rf.Spec.Redis.HostNetwork,
					DNSPolicy:         getDnsPolicy(rf.Spec.Redis.DNSPolicy),
					ImagePullSecrets:  rf.Spec.Redis.ImagePullSecrets,
					PriorityClassName: rf.Spec.Redis.PriorityClassName,
					InitContainers: []corev1.Container{
						{
							Name:            redisConfigCopy,
							Image:           rf.Spec.Redis.Image,
							ImagePullPolicy: pullPolicy(rf.Spec.Redis.ImagePullPolicy),
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      redisConfig,
									MountPath: redisConfMountPath,
								},
								{
									Name:      getRedisDataVolumeName(rf),
									MountPath: "/data",
								},
							},
							Command: []string{
								"sh",
								"-c",
								"if test ! -f \"" + GetRedisConfigWritablePath() + "\"; then echo \"not exists\" && mkdir -p " + redisConfWritableMountPath + " && cp " + GetRedisConfigPath() + " " + GetRedisConfigWritablePath() + "; else echo \"exists\"; fi",
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("32Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("32Mi"),
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "TZ",
									Value: "Asia/Shanghai",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            redisName,
							Image:           rf.Spec.Redis.Image,
							ImagePullPolicy: pullPolicy(rf.Spec.Redis.ImagePullPolicy),
							Ports: []corev1.ContainerPort{
								{
									Name:          redisName,
									ContainerPort: int32(portInt),
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: volumeMounts,
							Command:      redisCommand,
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: graceTime,
								TimeoutSeconds:      5,
								PeriodSeconds:       defaultPeriodSeconds,
								SuccessThreshold:    defaultSuccessThreshold,
								FailureThreshold:    defaultFailureThreshold,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/sh", "/redis-readiness/ready.sh"},
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								InitialDelaySeconds: graceTime,
								TimeoutSeconds:      5,
								PeriodSeconds:       defaultPeriodSeconds,
								SuccessThreshold:    defaultSuccessThreshold,
								FailureThreshold:    defaultFailureThreshold,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											livenessProbeCommand,
										},
									},
								},
							},
							Resources: rf.Spec.Redis.Resources,
							Env: []corev1.EnvVar{
								{
									Name:  "TZ",
									Value: "Asia/Shanghai",
								},
							},
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	if rf.Spec.Redis.Storage.PersistentVolumeClaim != nil {
		if !rf.Spec.Redis.Storage.KeepAfterDeletion {
			// Set an owner reference so the persistent volumes are deleted when the RF is
			rf.Spec.Redis.Storage.PersistentVolumeClaim.OwnerReferences = ownerRefs
		}
		ss.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			*rf.Spec.Redis.Storage.PersistentVolumeClaim,
		}
	}

	if rf.Spec.Redis.StorageLog.PersistentVolumeClaim != nil {
		if !rf.Spec.Redis.StorageLog.KeepAfterDeletion {
			// Set an owner reference so the persistent volumes are deleted when the RF is
			rf.Spec.Redis.StorageLog.PersistentVolumeClaim.OwnerReferences = ownerRefs
		}
		ss.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			*rf.Spec.Redis.StorageLog.PersistentVolumeClaim,
		}
	}

	return ss
}

func CreateRedisStatefulSetObjByExistingObjByIndex(rf *roav1.Redis, ownerRefs []metav1.OwnerReference, oldStatefulSet *v1.StatefulSet, index int) *v1.StatefulSet {
	oldStatefulSet.Spec.Template.Spec.Containers = SetResourcesByContainerName(redisName, oldStatefulSet.Spec.Template.Spec.Containers, rf.Spec.Redis.Resources)
	return oldStatefulSet
}

func RedisStatefulSetEqual(a *v1.StatefulSet, b *v1.StatefulSet) bool {
	return reflect.DeepEqual(getResourcesByContainerName(redisName, a.Spec.Template.Spec.Containers), getResourcesByContainerName(redisName, b.Spec.Template.Spec.Containers))
}

func getResourcesByContainerName(name string, container []corev1.Container) *corev1.ResourceRequirements {
	for _, c := range container {
		if c.Name == name {
			return &c.Resources
		}
	}
	return nil
}

func SetResourcesByContainerName(name string, oldContainers []corev1.Container, resources corev1.ResourceRequirements) []corev1.Container {
	containers := make([]corev1.Container, 0)
	for _, c := range oldContainers {
		if c.Name == name {
			c.Resources = resources
		}
		containers = append(containers, c)
	}
	return containers
}

func getRedisUpdateStrategy(rf *roav1.Redis) v1.StatefulSetUpdateStrategy {
	if rf.Spec.Redis.UpdateStrategy.Type == "" {
		return v1.StatefulSetUpdateStrategy{
			Type: v1.RollingUpdateStatefulSetStrategyType,
		}
	}
	return rf.Spec.Redis.UpdateStrategy
}

func getRedisVolumeMounts(rf *roav1.Redis) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      redisReadinessVolumeName,
			MountPath: "/redis-readiness",
		},
		{
			Name:      getRedisDataVolumeName(rf),
			MountPath: "/data",
		},
		{
			Name:      getRedisLogVolumeName(rf),
			MountPath: "/redislog",
		},
	}

	return volumeMounts
}

func getRedisVolumesByIndex(rf *roav1.Redis, index int) []corev1.Volume {
	configMapName := GetRedisConfigMapNameByIndex(rf, index)

	readinessConfigMapName := GetRedisReadinessConfigMapName(rf)

	defaultMode := corev1.ConfigMapVolumeSourceDefaultMode

	executeMode := int32(0744)
	volumes := []corev1.Volume{
		{
			Name: redisConfig,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
					DefaultMode: &defaultMode,
				},
			},
		},
		{
			Name: redisReadinessVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: readinessConfigMapName,
					},
					DefaultMode: &executeMode,
				},
			},
		},
	}

	dataVolume := getRedisDataVolume(rf)
	if dataVolume != nil {
		volumes = append(volumes, *dataVolume)
	}

	logVolume := getRedisLogVolume(rf)
	if logVolume != nil {
		volumes = append(volumes, *logVolume)
	}

	return volumes
}

func getRedisDataVolumeName(rf *roav1.Redis) string {
	switch {
	case rf.Spec.Redis.Storage.PersistentVolumeClaim != nil:
		return rf.Spec.Redis.Storage.PersistentVolumeClaim.Name
	case rf.Spec.Redis.Storage.EmptyDir != nil:
		return redisStorageVolumeName
	default:
		return redisStorageVolumeName
	}
}

func getRedisDataVolume(rf *roav1.Redis) *corev1.Volume {

	// This will find the volumed desired by the user. If no volume defined
	// an EmptyDir will be used by default
	switch {
	case rf.Spec.Redis.Storage.PersistentVolumeClaim != nil:
		return nil
	case rf.Spec.Redis.Storage.EmptyDir != nil:
		return &corev1.Volume{
			Name: redisStorageVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: rf.Spec.Redis.Storage.EmptyDir,
			},
		}
	default:
		return &corev1.Volume{
			Name: redisStorageVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}
	}

}

func getRedisLogVolumeName(rf *roav1.Redis) string {
	switch {
	case rf.Spec.Redis.StorageLog.PersistentVolumeClaim != nil:
		return rf.Spec.Redis.StorageLog.PersistentVolumeClaim.Name
	case rf.Spec.Redis.StorageLog.EmptyDir != nil:
		return redisLogStorageVolumeName
	default:
		return redisLogStorageVolumeName
	}
}

func getRedisLogVolume(rf *roav1.Redis) *corev1.Volume {

	// This will find the volumed desired by the user. If no volume defined
	// an EmptyDir will be used by default
	switch {
	case rf.Spec.Redis.StorageLog.PersistentVolumeClaim != nil:
		return nil
	case rf.Spec.Redis.StorageLog.EmptyDir != nil:
		return &corev1.Volume{
			Name: redisLogStorageVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: rf.Spec.Redis.StorageLog.EmptyDir,
			},
		}
	default:
		return &corev1.Volume{
			Name: redisLogStorageVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}
	}

}

func getRedisCommand(rf *roav1.Redis) []string {
	if len(rf.Spec.Redis.Command) > 0 {
		return rf.Spec.Redis.Command
	}
	return []string{
		"redis-server",
		GetRedisConfigWritablePath(),
	}
}

// GetRedisRootName returns the name for redis resources
func GetRedisRootName(rf *roav1.Redis) string {
	return generateName(redisRootName, rf.Name)
}

func GetRedisNameByIndex(rf *roav1.Redis, index int) string {
	return GetRedisRootName(rf) + "-" + strconv.Itoa(index)
}

func GetRedisLabels(rf *roav1.Redis) map[string]string {
	return GenerateSelectorLabels(redisRootName, rf)
}

func GetRedisLabelsWithName(rf *roav1.Redis, name string) map[string]string {
	return MergeLabels(
		GetRedisLabels(rf),
		map[string]string{
			statefulSetNameLabelKey: name,
		},
	)
}

func getRedisAnnotations(rf *roav1.Redis) map[string]string {
	annotations := rf.Spec.Redis.PodAnnotations
	return annotations
}

func getRedisNodeAffinityByIndex(rf *roav1.Redis, index int) *corev1.NodeAffinity {
	if rf.Spec.Redis.StaticResources == nil || len(rf.Spec.Redis.StaticResources) < 1 {
		return nil
	}

	staticResource := rf.Spec.Redis.StaticResources[index]

	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "kubernetes.io/hostname",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{staticResource.Host},
						},
					},
				},
			},
		},
	}
}
