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

func CreateSentinelStatefulSetObjByIndex(rf *roav1.Redis, ownerRefs []metav1.OwnerReference, index int) *v1.StatefulSet {
	name := GetSentinelNameByIndex(rf, index)
	namespace := rf.Namespace

	sentinelCommand := getSentinelCommand(rf)
	selector := GetSentinelLabels(rf)
	labels := GetSentinelLabelsWithName(rf, name)
	volumeMounts := getSentinelVolumeMounts(rf)
	volumes := getSentinelVolumes(rf, index)
	annotations := getSentinelAnnotations(rf)
	nodeAffinity := getSentinelNodeAffinityByIndex(rf, index)
	affinity := getNodeAndPodAffinity(rf.Spec.Sentinel.Affinity, rf.Spec.Sentinel.EnabledPodAntiAffinity, selector, nodeAffinity)
	port := GetSentinelPortFromSpecByIndex(rf, index)
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
			UpdateStrategy:      getSentinelUpdateStrategy(rf),
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
					Tolerations:       rf.Spec.Sentinel.Tolerations,
					NodeSelector:      rf.Spec.Sentinel.NodeSelector,
					SecurityContext:   getSecurityContext(rf.Spec.Sentinel.SecurityContext),
					HostNetwork:       rf.Spec.Sentinel.HostNetwork,
					DNSPolicy:         getDnsPolicy(rf.Spec.Sentinel.DNSPolicy),
					ImagePullSecrets:  rf.Spec.Sentinel.ImagePullSecrets,
					PriorityClassName: rf.Spec.Sentinel.PriorityClassName,
					InitContainers: []corev1.Container{
						{
							Name:            sentinelConfigCopy,
							Image:           rf.Spec.Sentinel.Image,
							ImagePullPolicy: pullPolicy(rf.Spec.Sentinel.ImagePullPolicy),
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      sentinelConfig,
									MountPath: sentinelConfMountPath,
								},
								{
									Name:      getSentinelDataVolumeName(rf),
									MountPath: "/data",
								},
							},
							Command: []string{
								"sh",
								"-c",
								"if test ! -f \"" + GetSentinelConfigWritablePath() + "\"; then echo \"not exists\" && mkdir -p " + sentinelConfWritableMountPath + " && cp " + GetSentinelConfigPath() + " " + GetSentinelConfigWritablePath() + "; else echo \"exists\"; fi",
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
							Name:            sentinelName,
							Image:           rf.Spec.Sentinel.Image,
							ImagePullPolicy: pullPolicy(rf.Spec.Sentinel.ImagePullPolicy),
							Ports: []corev1.ContainerPort{
								{
									Name:          sentinelName,
									ContainerPort: int32(portInt),
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: volumeMounts,
							Command:      sentinelCommand,
							ReadinessProbe: &corev1.Probe{
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
							Resources: rf.Spec.Sentinel.Resources,
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

	if rf.Spec.Sentinel.Storage.PersistentVolumeClaim != nil {
		if !rf.Spec.Sentinel.Storage.KeepAfterDeletion {
			// Set an owner reference so the persistent volumes are deleted when the RF is
			rf.Spec.Sentinel.Storage.PersistentVolumeClaim.OwnerReferences = ownerRefs
		}
		ss.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			*rf.Spec.Sentinel.Storage.PersistentVolumeClaim,
		}
	}

	if rf.Spec.Sentinel.StorageLog.PersistentVolumeClaim != nil {
		if !rf.Spec.Sentinel.StorageLog.KeepAfterDeletion {
			// Set an owner reference so the persistent volumes are deleted when the RF is
			rf.Spec.Sentinel.StorageLog.PersistentVolumeClaim.OwnerReferences = ownerRefs
		}
		ss.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			*rf.Spec.Sentinel.StorageLog.PersistentVolumeClaim,
		}
	}

	return ss
}

func CreateSentinelStatefulSetObjByExistingObjByIndex(rf *roav1.Redis, ownerRefs []metav1.OwnerReference, oldStatefulSet *v1.StatefulSet, index int) *v1.StatefulSet {
	oldStatefulSet.Spec.Template.Spec.Containers = SetResourcesByContainerName(sentinelName, oldStatefulSet.Spec.Template.Spec.Containers, rf.Spec.Sentinel.Resources)
	return oldStatefulSet
}

func SentinelStatefulSetEqual(a *v1.StatefulSet, b *v1.StatefulSet) bool {
	return reflect.DeepEqual(getResourcesByContainerName(sentinelName, a.Spec.Template.Spec.Containers), getResourcesByContainerName(sentinelName, b.Spec.Template.Spec.Containers))
}

func getSentinelUpdateStrategy(rf *roav1.Redis) v1.StatefulSetUpdateStrategy {
	if rf.Spec.Sentinel.UpdateStrategy.Type == "" {
		return v1.StatefulSetUpdateStrategy{
			Type: v1.RollingUpdateStatefulSetStrategyType,
		}
	}
	return rf.Spec.Sentinel.UpdateStrategy
}

func getSentinelVolumeMounts(rf *roav1.Redis) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      getSentinelDataVolumeName(rf),
			MountPath: "/data",
		},
		{
			Name:      getSentinelLogVolumeName(rf),
			MountPath: "/redislog",
		},
	}
}

func getSentinelVolumes(rf *roav1.Redis, index int) []corev1.Volume {
	configMapName := GetSentinelConfigMapNameByIndex(rf, index)

	defaultMode := corev1.ConfigMapVolumeSourceDefaultMode
	volumes := []corev1.Volume{
		{
			Name: sentinelConfig,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
					DefaultMode: &defaultMode,
				},
			},
		},
	}

	dataVolume := getSentinelDataVolume(rf)
	if dataVolume != nil {
		volumes = append(volumes, *dataVolume)
	}

	logVolume := getSentinelLogVolume(rf)
	if logVolume != nil {
		volumes = append(volumes, *logVolume)
	}

	return volumes
}

func getSentinelDataVolumeName(rf *roav1.Redis) string {
	switch {
	case rf.Spec.Sentinel.Storage.PersistentVolumeClaim != nil:
		return rf.Spec.Sentinel.Storage.PersistentVolumeClaim.Name
	case rf.Spec.Sentinel.Storage.EmptyDir != nil:
		return redisStorageVolumeName
	default:
		return redisStorageVolumeName
	}
}

func getSentinelDataVolume(rf *roav1.Redis) *corev1.Volume {

	// This will find the volumed desired by the user. If no volume defined
	// an EmptyDir will be used by default
	switch {
	case rf.Spec.Sentinel.Storage.PersistentVolumeClaim != nil:
		return nil
	case rf.Spec.Sentinel.Storage.EmptyDir != nil:
		return &corev1.Volume{
			Name: redisStorageVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: rf.Spec.Sentinel.Storage.EmptyDir,
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

func getSentinelLogVolumeName(rf *roav1.Redis) string {

	switch {
	case rf.Spec.Sentinel.StorageLog.PersistentVolumeClaim != nil:
		return rf.Spec.Sentinel.StorageLog.PersistentVolumeClaim.Name
	case rf.Spec.Sentinel.StorageLog.EmptyDir != nil:
		return redisLogStorageVolumeName
	default:
		return redisLogStorageVolumeName
	}
}

func getSentinelLogVolume(rf *roav1.Redis) *corev1.Volume {

	// This will find the volumed desired by the user. If no volume defined
	// an EmptyDir will be used by default
	switch {
	case rf.Spec.Sentinel.StorageLog.PersistentVolumeClaim != nil:
		return nil
	case rf.Spec.Sentinel.StorageLog.EmptyDir != nil:
		return &corev1.Volume{
			Name: redisLogStorageVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: rf.Spec.Sentinel.StorageLog.EmptyDir,
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

func getSentinelCommand(rf *roav1.Redis) []string {
	if len(rf.Spec.Sentinel.Command) > 0 {
		return rf.Spec.Sentinel.Command
	}
	return []string{
		"redis-server",
		GetSentinelConfigWritablePath(),
		"--sentinel",
	}
}

// GetSentinelRootName returns the name for sentinel resources
func GetSentinelRootName(rf *roav1.Redis) string {
	return generateName(sentinelRootName, rf.Name)
}

func GetSentinelNameByIndex(rf *roav1.Redis, index int) string {
	return GetSentinelRootName(rf) + "-" + strconv.Itoa(index)
}

func GetSentinelLabels(rf *roav1.Redis) map[string]string {
	return GenerateSelectorLabels(sentinelRootName, rf)
}

func GetSentinelLabelsWithName(rf *roav1.Redis, name string) map[string]string {
	return MergeLabels(
		GetSentinelLabels(rf),
		map[string]string{
			statefulSetNameLabelKey: name,
		},
	)
}

func getSentinelAnnotations(rf *roav1.Redis) map[string]string {
	annotations := rf.Spec.Sentinel.PodAnnotations
	return annotations
}

func getSentinelNodeAffinityByIndex(rf *roav1.Redis, index int) *corev1.NodeAffinity {
	if rf.Spec.Sentinel.StaticResources == nil || len(rf.Spec.Sentinel.StaticResources) < 1 {
		return nil
	}

	staticResource := rf.Spec.Sentinel.StaticResources[index]

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
