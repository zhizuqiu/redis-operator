package util

import (
	"bytes"
	"fmt"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"text/template"
)

func CreateSentinelConfigMapByIndex(rf *roav1.Redis, ownerRefs []metav1.OwnerReference, password string, index int) *corev1.ConfigMap {
	name := GetSentinelConfigMapNameByIndex(rf, index)
	namespace := rf.Namespace

	labels := GetSentinelSlaveConfigMapLabels(rf)

	quorum := strconv.Itoa(int(GetQuorum(rf)))
	masterIp, masterPort := GetMasterIpAndPortFromSpec(rf)
	realSentinelConfigFileContent := fmt.Sprintf("sentinel monitor %s %s %s %s\n%s", redisGroupName, masterIp, masterPort, quorum, sentinelConfigFile)

	port := GetSentinelPortFromSpecByIndex(rf, index)
	realSentinelConfigFileContent = fmt.Sprintf("port %s\n%s", port, realSentinelConfigFileContent)

	if password != "" {
		realSentinelConfigFileContent = fmt.Sprintf("%s\nsentinel auth-pass %s \"%s\"", realSentinelConfigFileContent, redisGroupName, password)
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Data: map[string]string{
			sentinelConfigFileName: realSentinelConfigFileContent,
		},
	}
}

func CreateSentinelConfigMapObjByExistingObjByIndex(rf *roav1.Redis, password string, oldConfigMap *corev1.ConfigMap, index int) *corev1.ConfigMap {

	quorum := strconv.Itoa(int(GetQuorum(rf)))
	masterIp, masterPort := GetMasterIpAndPortFromSpec(rf)
	realSentinelConfigFileContent := fmt.Sprintf("sentinel monitor mymaster %s %s %s\n%s", masterIp, masterPort, quorum, sentinelConfigFile)

	port := GetSentinelPortFromSpecByIndex(rf, index)
	realSentinelConfigFileContent = fmt.Sprintf("port %s\n%s", port, realSentinelConfigFileContent)

	if password != "" {
		realSentinelConfigFileContent = fmt.Sprintf("%s\nsentinel auth-pass %s \"%s\"", realSentinelConfigFileContent, redisGroupName, password)
	}

	oldConfigMap.Data = map[string]string{
		sentinelConfigFileName: realSentinelConfigFileContent,
	}

	return oldConfigMap
}

func CreateReadinessConfigMap(rf *roav1.Redis, ownerRefs []metav1.OwnerReference) *corev1.ConfigMap {

	name := GetRedisReadinessConfigMapName(rf)
	namespace := rf.Namespace

	labels := GetReadinessConfigMapLabels(rf)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Data: map[string]string{
			"ready.sh": readinessContent,
		},
	}
}

func GetReadinessConfigMapLabels(rf *roav1.Redis) map[string]string {
	return GenerateSelectorLabels(redisRootName, rf)
}

func CreateRedisMasterConfigMap(rf *roav1.Redis, ownerRefs []metav1.OwnerReference, password string) *corev1.ConfigMap {
	name := GetRedisConfigMapNameByIndex(rf, 0)
	labels := GetRedisMasterConfigMapLabels(rf)

	tmpl, err := template.New("redis").Parse(redisConfigTemplate)
	if err != nil {
		panic(err)
	}

	var tplOutput bytes.Buffer
	if err := tmpl.Execute(&tplOutput, rf); err != nil {
		panic(err)
	}

	redisConfigFileContent := tplOutput.String()

	_, port := GetMasterIpAndPortFromSpec(rf)
	redisConfigFileContent = fmt.Sprintf("port %s\n%s", port, redisConfigFileContent)

	if password != "" {
		redisConfigFileContent = fmt.Sprintf("%s\nmasterauth \"%s\"\nrequirepass \"%s\"", redisConfigFileContent, password, password)
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       rf.Namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Data: map[string]string{
			redisConfigFileName: redisConfigFileContent,
		},
	}

}

func CreateRedisSlaveConfigMapByIndex(rf *roav1.Redis, ownerRefs []metav1.OwnerReference, password string, index int) *corev1.ConfigMap {
	name := GetRedisConfigMapNameByIndex(rf, index)
	labels := GetRedisSlaveConfigMapLabels(rf)

	tmpl, err := template.New("redis").Parse(redisConfigTemplate)
	if err != nil {
		panic(err)
	}

	var tplOutput bytes.Buffer
	if err := tmpl.Execute(&tplOutput, rf); err != nil {
		panic(err)
	}

	redisConfigFileContent := tplOutput.String()

	masterIp, masterPort := GetMasterIpAndPortFromSpec(rf)
	redisConfigFileContent = fmt.Sprintf("replicaof %s %s\n%s", masterIp, masterPort, redisConfigFileContent)

	port := GetRedisPortFromSpecByIndex(rf, index)
	redisConfigFileContent = fmt.Sprintf("port %s\n%s", port, redisConfigFileContent)

	if password != "" {
		redisConfigFileContent = fmt.Sprintf("%s\nmasterauth \"%s\"\nrequirepass \"%s\"", redisConfigFileContent, password, password)
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       rf.Namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Data: map[string]string{
			redisConfigFileName: redisConfigFileContent,
		},
	}

}

func GetRedisMasterConfigMapLabels(rf *roav1.Redis) map[string]string {
	return GenerateSelectorLabels(redisRoleName, rf)
}

func GetRedisSlaveConfigMapLabels(rf *roav1.Redis) map[string]string {
	return GenerateSelectorLabels(redisRoleName, rf)
}

func CreateRedisMasterConfigMapObjByExistingObj(rf *roav1.Redis, password string, oldConfigMap *corev1.ConfigMap) *corev1.ConfigMap {
	tmpl, err := template.New("redis").Parse(redisConfigTemplate)
	if err != nil {
		panic(err)
	}

	var tplOutput bytes.Buffer
	if err := tmpl.Execute(&tplOutput, rf); err != nil {
		panic(err)
	}

	redisConfigFileContent := tplOutput.String()

	_, port := GetMasterIpAndPortFromSpec(rf)
	redisConfigFileContent = fmt.Sprintf("port %s\n%s", port, redisConfigFileContent)

	if password != "" {
		redisConfigFileContent = fmt.Sprintf("%s\nmasterauth \"%s\"\nrequirepass \"%s\"", redisConfigFileContent, password, password)
	}

	oldConfigMap.Data = map[string]string{
		redisConfigFileName: redisConfigFileContent,
	}

	return oldConfigMap
}

func CreateRedisSlaveConfigMapObjByExistingObjByIndex(rf *roav1.Redis, password string, oldConfigMap *corev1.ConfigMap, index int) *corev1.ConfigMap {
	tmpl, err := template.New("redis").Parse(redisConfigTemplate)
	if err != nil {
		panic(err)
	}

	var tplOutput bytes.Buffer
	if err := tmpl.Execute(&tplOutput, rf); err != nil {
		panic(err)
	}

	redisConfigFileContent := tplOutput.String()

	masterIp, masterPort := GetMasterIpAndPortFromSpec(rf)
	redisConfigFileContent = fmt.Sprintf("replicaof %s %s\n%s", masterIp, masterPort, redisConfigFileContent)

	port := GetRedisPortFromSpecByIndex(rf, index)
	redisConfigFileContent = fmt.Sprintf("port %s\n%s", port, redisConfigFileContent)

	if password != "" {
		redisConfigFileContent = fmt.Sprintf("%s\nmasterauth \"%s\"\nrequirepass \"%s\"", redisConfigFileContent, password, password)
	}

	oldConfigMap.Data = map[string]string{
		redisConfigFileName: redisConfigFileContent,
	}

	return oldConfigMap
}

func GetRedisConfigMapName(rf *roav1.Redis) string {
	return generateName(redisRootName, rf.Name)
}

// GetRedisReadinessConfigMapName returns the name for redis resources
func GetRedisReadinessConfigMapName(rf *roav1.Redis) string {
	return generateName(redisReadinessName, rf.Name)
}

func GetRedisConfigMapNameByIndex(rf *roav1.Redis, index int) string {
	return generateName(redisRoleName, rf.Name) + "-" + strconv.Itoa(index)
}

func GetSentinelConfigMapNameByIndex(rf *roav1.Redis, index int) string {
	return generateName(sentinelRoleName, rf.Name) + "-" + strconv.Itoa(index)
}

func GetSentinelSlaveConfigMapLabels(rf *roav1.Redis) map[string]string {
	return GenerateSelectorLabels(sentinelRoleName, rf)
}
