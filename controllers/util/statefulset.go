package util

import (
	"fmt"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

func generateName(typeName, metaName string) string {
	return fmt.Sprintf("%s%s-%s", baseName, typeName, metaName)
}

func GenerateSelectorLabels(component string, rf *roav1.Redis) map[string]string {
	return MergeLabels(map[string]string{
		appNameLabelKey:      rf.Name,
		appComponentLabelKey: component,
		appPartOfLabelKey:    appLabel,
	}, rf.Labels)
}

func getDnsPolicy(dnspolicy corev1.DNSPolicy) corev1.DNSPolicy {
	if dnspolicy == "" {
		return corev1.DNSClusterFirst
	}
	return dnspolicy
}

func getSecurityContext(secctx *corev1.PodSecurityContext) *corev1.PodSecurityContext {
	if secctx != nil {
		return secctx
	}

	defaultUserAndGroup := int64(1000)
	runAsNonRoot := true

	return &corev1.PodSecurityContext{
		RunAsUser:    &defaultUserAndGroup,
		RunAsGroup:   &defaultUserAndGroup,
		RunAsNonRoot: &runAsNonRoot,
		FSGroup:      &defaultUserAndGroup,
	}
}

func getNodeAndPodAffinity(affinity *corev1.Affinity, enabledPodAntiAffinity bool, labels map[string]string, nodeAffinity *corev1.NodeAffinity) *corev1.Affinity {
	var aff *corev1.Affinity
	if affinity != nil {
		aff = affinity
	}

	if enabledPodAntiAffinity {
		hostnamePodAffinityTerm := getHostnamePodAffinityTerm(labels)
		if aff == nil {
			aff = &corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						hostnamePodAffinityTerm,
					},
				},
			}
		} else {
			if aff.PodAntiAffinity == nil {
				aff.PodAntiAffinity = &corev1.PodAntiAffinity{}
			}
			if !hasPodAffinityTerm(aff.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, hostnamePodAffinityTerm) {
				aff.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
					aff.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
					hostnamePodAffinityTerm,
				)
			}
		}
	}

	if nodeAffinity != nil {
		if aff == nil {
			aff = &corev1.Affinity{
				NodeAffinity: nodeAffinity,
			}
		} else {
			// todo no overwrite
			aff.NodeAffinity = nodeAffinity
		}
	}

	return aff
}

func getAffinity(affinity *corev1.Affinity, labels map[string]string) *corev1.Affinity {
	var aff *corev1.Affinity
	if affinity != nil {
		aff = affinity
	}

	hostnamePodAffinityTerm := getHostnamePodAffinityTerm(labels)
	if aff == nil {
		aff = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					hostnamePodAffinityTerm,
				},
			},
		}
	} else {
		if aff.PodAntiAffinity == nil {
			aff.PodAntiAffinity = &corev1.PodAntiAffinity{}
		}
		if !hasPodAffinityTerm(aff.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, hostnamePodAffinityTerm) {
			aff.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
				aff.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
				hostnamePodAffinityTerm,
			)
		}
	}

	return aff
}

func hasPodAffinityTerm(podAffinityTerms []corev1.PodAffinityTerm, podAffinityTerm corev1.PodAffinityTerm) bool {
	for _, term := range podAffinityTerms {
		if reflect.DeepEqual(term, podAffinityTerm) {
			return true
		}
	}
	return false
}

func getHostnamePodAffinityTerm(labels map[string]string) corev1.PodAffinityTerm {
	return corev1.PodAffinityTerm{
		TopologyKey: hostnameTopologyKey,
		LabelSelector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
	}
}

func pullPolicy(specPolicy corev1.PullPolicy) corev1.PullPolicy {
	if specPolicy == "" {
		return corev1.PullIfNotPresent
	}
	return specPolicy
}
