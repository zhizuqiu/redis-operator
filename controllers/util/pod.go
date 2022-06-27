package util

import (
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func GetRoleFromLabel(pod v1.Pod) string {
	if pod.Labels == nil {
		return "unknown"
	}
	if rootName, ok := pod.Labels[appComponentLabelKey]; ok {
		if rootName == redisRootName {
			return redisRoleName
		} else if rootName == sentinelRootName {
			return sentinelRoleName
		} else if rootName == exporterRootName {
			return exporterRoleName
		}
		return rootName
	}
	return "unknown"
}

func GetContainerNameFromLabel(pod v1.Pod) string {
	if pod.Labels == nil {
		return "unknown"
	}
	if rootName, ok := pod.Labels[appComponentLabelKey]; ok {
		if rootName == redisRootName {
			return redisName
		} else if rootName == sentinelRootName {
			return sentinelName
		} else if rootName == exporterRootName {
			return exporterName
		}
		return rootName
	}
	return "unknown"
}

func getContainerPort(containerName, portName string, containers []v1.Container) int32 {
	for _, c := range containers {
		if c.Name == containerName {
			ports := c.Ports
			if ports != nil {
				for _, p := range ports {
					if p.Name == portName {
						return p.ContainerPort
					}
				}
			}
		}
	}
	return 0
}

func GetPort(pod v1.Pod) int32 {
	role := GetRoleFromLabel(pod)
	if role == sentinelRoleName {
		return getContainerPort(sentinelName, sentinelName, pod.Spec.Containers)
	} else if role == redisRoleName {
		return getContainerPort(redisName, redisName, pod.Spec.Containers)
	} else if role == exporterRoleName {
		return getContainerPort(exporterName, exporterName, pod.Spec.Containers)
	} else {
		return 0
	}
}

func GetGlobalPhase(redis *roav1.Redis, podList *v1.PodList) v1.PodPhase {
	if podList == nil {
		return v1.PodPending
	}
	if redis.Spec.Exporter.Enabled {
		if len(podList.Items) != int(redis.Spec.Redis.Replicas+redis.Spec.Sentinel.Replicas)+1 {
			return v1.PodPending
		}
	} else {
		if len(podList.Items) != int(redis.Spec.Redis.Replicas+redis.Spec.Sentinel.Replicas) {
			return v1.PodPending
		}
	}
	for _, pod := range podList.Items {
		if pod.Status.Phase != v1.PodRunning {
			return v1.PodPending
		}
	}
	return v1.PodRunning
}

func getContainerStatus(containerName string, containerStatuses []v1.ContainerStatus) *v1.ContainerStatus {
	if containerStatuses == nil {
		return nil
	}
	for _, c := range containerStatuses {
		if c.Name == containerName {
			return &c
		}
	}
	return nil
}

func GetGlobalReady(redis *roav1.Redis, podList *v1.PodList) bool {
	if podList == nil {
		return false
	}
	if redis.Spec.Exporter.Enabled {
		if len(podList.Items) != int(redis.Spec.Redis.Replicas+redis.Spec.Sentinel.Replicas)+1 {
			return false
		}
	} else {
		if len(podList.Items) != int(redis.Spec.Redis.Replicas+redis.Spec.Sentinel.Replicas) {
			return false
		}
	}
	for _, pod := range podList.Items {
		containerName := GetContainerNameFromLabel(pod)
		containerStatus := getContainerStatus(containerName, pod.Status.ContainerStatuses)
		if containerStatus == nil {
			return false
		}
		if containerStatus.Ready != true {
			return false
		}
	}
	return true
}

func GetInstanceLabels(name string) map[string]string {
	return map[string]string{
		appNameLabelKey:   name,
		appPartOfLabelKey: appLabel,
	}
}
