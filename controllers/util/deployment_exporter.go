package util

import (
	"fmt"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	"github.com/zhizuqiu/redis-operator/controllers/util/sm4"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"strconv"
)

func CreateExporterDeployment(rf *roav1.Redis, ownerRefs []metav1.OwnerReference, password string) *v1.Deployment {
	name := GetExporterRootName(rf)
	namespace := rf.Namespace

	labels := GetExporterDeploymentLabels(rf)
	redisAddr := GetRedisAddr(rf)
	sentinelAddr := GetSentinelAddr(rf)
	productId := GetProductId(rf)
	regionId := GetRegionId(rf)
	instanceId := GetInstanceId(rf)
	annotations := getExporterAnnotations(rf)

	nodeAffinity := getExporterNodeAffinity(rf)
	affinity := getNodeAndPodAffinity(rf.Spec.Sentinel.Affinity, rf.Spec.Sentinel.EnabledPodAntiAffinity, labels, nodeAffinity)
	port := GetExporterPortFromSpec(rf)
	portInt, _ := strconv.Atoi(port)

	sm4Password, err := sm4.EncryptSm4([]byte(password), sm4.Sm4Key)
	if err != nil {
		fmt.Println(err)
	}

	dd := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Spec: v1.DeploymentSpec{
			Replicas: Int32P(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Affinity:    affinity,
					HostNetwork: rf.Spec.Exporter.HostNetwork,
					Containers: []corev1.Container{
						{
							Name:            exporterRoleName,
							Image:           rf.Spec.Exporter.Image,
							ImagePullPolicy: pullPolicy(rf.Spec.Exporter.ImagePullPolicy),
							Ports: []corev1.ContainerPort{
								{
									Name:          exporterContainerName,
									ContainerPort: int32(portInt),
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "REDIS_EXPORTER_WEB_LISTEN_ADDRESS",
									Value: ":" + port,
								},
								{
									Name:  "REDIS_ADDR",
									Value: redisAddr,
								},
								{
									Name:  "REDIS_SENTINEL_ADDR",
									Value: sentinelAddr,
								},
								{
									Name:  "REDIS_EXPORTER_REGION_ID",
									Value: regionId,
								},
								{
									Name:  "REDIS_EXPORTER_PRODUCT_ID",
									Value: productId,
								},
								{
									Name:  "REDIS_EXPORTER_INSTANCE_ID",
									Value: instanceId,
								},
								{
									Name:  "REDIS_EXPORTER_INSTANCE_NAME",
									Value: rf.Name,
								},
								{
									Name:  "REDIS_SM4_PASSWORD",
									Value: sm4Password,
								},
								{
									Name:  "TZ",
									Value: "Asia/Shanghai",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(exporterDefaultLimitCPU),
									corev1.ResourceMemory: resource.MustParse(exporterDefaultLimitMemory),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(exporterDefaultRequestCPU),
									corev1.ResourceMemory: resource.MustParse(exporterDefaultRequestMemory),
								},
							},
							LivenessProbe: &corev1.Probe{
								InitialDelaySeconds: graceTime,
								TimeoutSeconds:      5,
								PeriodSeconds:       defaultPeriodSeconds,
								SuccessThreshold:    defaultSuccessThreshold,
								FailureThreshold:    defaultFailureThreshold,
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/health",
										Port:   intstr.FromInt(portInt),
										Scheme: corev1.URISchemeHTTP,
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: graceTime,
								TimeoutSeconds:      5,
								PeriodSeconds:       defaultPeriodSeconds,
								SuccessThreshold:    defaultSuccessThreshold,
								FailureThreshold:    defaultFailureThreshold,
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/metrics",
										Port:   intstr.FromInt(portInt),
										Scheme: corev1.URISchemeHTTP,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return dd
}

func CreateExporterDeploymentObjByExistingObj(rf *roav1.Redis, password string, oldDeplyment *v1.Deployment) *v1.Deployment {

	redisAddr := GetRedisAddr(rf)
	sentinelAddr := GetSentinelAddr(rf)

	sm4Password, err := sm4.EncryptSm4([]byte(password), sm4.Sm4Key)
	if err != nil {
		fmt.Println(err)
	}

	oldDeplyment.Spec.Template.Spec.Containers = SetEnvByContainerName(exporterRoleName, oldDeplyment.Spec.Template.Spec.Containers, "REDIS_ADDR", redisAddr)
	oldDeplyment.Spec.Template.Spec.Containers = SetEnvByContainerName(exporterRoleName, oldDeplyment.Spec.Template.Spec.Containers, "REDIS_SENTINEL_ADDR", sentinelAddr)
	oldDeplyment.Spec.Template.Spec.Containers = SetEnvByContainerName(exporterRoleName, oldDeplyment.Spec.Template.Spec.Containers, "REDIS_SM4_PASSWORD", sm4Password)

	return oldDeplyment
}

func ExporterDeploymentEqual(a *v1.Deployment, b *v1.Deployment) bool {

	redisAddrOk := reflect.DeepEqual(getEnvByContainerName(redisName, a.Spec.Template.Spec.Containers, "REDIS_ADDR"), getEnvByContainerName(redisName, b.Spec.Template.Spec.Containers, "REDIS_ADDR"))
	sentinelAddrOk := reflect.DeepEqual(getEnvByContainerName(redisName, a.Spec.Template.Spec.Containers, "REDIS_SENTINEL_ADDR"), getEnvByContainerName(redisName, b.Spec.Template.Spec.Containers, "REDIS_SENTINEL_ADDR"))
	passOk := reflect.DeepEqual(getEnvByContainerName(redisName, a.Spec.Template.Spec.Containers, "REDIS_SM4_PASSWORD"), getEnvByContainerName(redisName, b.Spec.Template.Spec.Containers, "REDIS_SM4_PASSWORD"))

	if redisAddrOk && sentinelAddrOk && passOk {
		return true
	}
	return false
}

func getEnvByContainerName(name string, container []corev1.Container, key string) string {
	for _, c := range container {
		if c.Name == name {
			for _, envVar := range c.Env {
				if envVar.Name == key {
					return envVar.Value
				}
			}
		}
	}
	return ""
}

func SetEnvByContainerName(name string, oldContainers []corev1.Container, key, value string) []corev1.Container {
	containers := make([]corev1.Container, 0)
	for _, c := range oldContainers {
		if c.Name == name {
			envs := make([]corev1.EnvVar, 0)
			for _, envVar := range c.Env {
				if envVar.Name == key {
					envs = append(envs, corev1.EnvVar{
						Name:  key,
						Value: value,
					})
				} else {
					envs = append(envs, envVar)
				}
			}
			c.Env = envs
		}
		containers = append(containers, c)
	}
	return containers
}

func GetExporterRootName(rf *roav1.Redis) string {
	return generateName(exporterName, rf.Name)
}

func GetRedisAddr(rf *roav1.Redis) string {
	addr := ""
	for i := 0; i < int(rf.Spec.Redis.Replicas); i++ {
		host := GetRedisHostByIndex(rf, i)
		port := GetRedisPortFromSpecByIndex(rf, i)
		if 0 == i {
			addr = addr + "redis://" + host + ":" + port
		} else {
			addr = addr + ",redis://" + host + ":" + port
		}
	}
	return addr
}

func GetSentinelAddr(rf *roav1.Redis) string {
	addr := ""
	for i := 0; i < int(rf.Spec.Sentinel.Replicas); i++ {
		host := GetSentinelHostByIndex(rf, i)
		port := GetSentinelPortFromSpecByIndex(rf, i)
		if 0 == i {
			addr = addr + "redis://" + host + ":" + port
		} else {
			addr = addr + ",redis://" + host + ":" + port
		}
	}
	return addr
}

func GetProductId(rf *roav1.Redis) string {
	productId := "unknown"
	if rf.Labels != nil {
		if rf.Labels["product_id"] != "" {
			productId = rf.Labels["product_id"]
		}
	}
	return productId
}

func GetRegionId(rf *roav1.Redis) string {
	regionId := "unknown"
	if rf.Labels != nil {
		if rf.Labels["region_id"] != "" {
			regionId = rf.Labels["region_id"]
		}
	}
	return regionId
}

func GetInstanceId(rf *roav1.Redis) string {
	instanceId := "unknown"
	if rf.Labels != nil {
		if rf.Labels["instance_id"] != "" {
			instanceId = rf.Labels["instance_id"]
		}
	}
	return instanceId
}

func getExporterAnnotations(rf *roav1.Redis) map[string]string {
	annotations := make(map[string]string)
	return annotations
}

func GetExporterDeploymentLabels(rf *roav1.Redis) map[string]string {
	return GenerateSelectorLabels(exporterRoleName, rf)
}

func getExporterNodeAffinity(rf *roav1.Redis) *corev1.NodeAffinity {
	if rf.Spec.Exporter.StaticResource.Host == "" {
		return nil
	}

	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "kubernetes.io/hostname",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{rf.Spec.Exporter.StaticResource.Host},
						},
					},
				},
			},
		},
	}
}
