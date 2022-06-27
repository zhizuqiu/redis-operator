package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"strconv"
	"time"
)

func Floadt64ToString(f float64) string {
	return strconv.FormatFloat(f, 'E', -1, 64)
}

func Int32P(i int32) *int32 {
	return &i
}

func Int64P(i int64) *int64 {
	return &i
}

func PrintNowTime() {
	fmt.Println(time.Now().Format("2006-01-02 15:04"))
}

func CreateNowTime() string {
	return time.Now().Format("2006-01-02 15:04")
}

// MergeLabels merges all the label maps received as argument into a single new label map.
func MergeLabels(allLabels ...map[string]string) map[string]string {
	res := map[string]string{}

	for _, labels := range allLabels {
		if labels != nil {
			for k, v := range labels {
				res[k] = v
			}
		}
	}
	return res
}

func NilError() error {
	return nil
}

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func SearchStatefulSetByName(name string, ssl *appsv1.StatefulSetList) *appsv1.StatefulSet {
	if ssl == nil {
		return nil
	}
	for _, item := range ssl.Items {
		if item.Name == name {
			return &item
		}
	}
	return nil
}

func GetMasterIpAndPortFromSpec(rf *roav1.Redis) (string, string) {
	masterIp := "127.0.0.1"
	masterPort := strconv.Itoa(redisContainerPort)

	if len(rf.Spec.Redis.StaticResources) > 0 {
		masterIp = rf.Spec.Redis.StaticResources[0].Host
		masterPort = strconv.Itoa(rf.Spec.Redis.StaticResources[0].Port)
	}

	return masterIp, masterPort
}

func GetSentinelPortFromSpecByIndex(rf *roav1.Redis, index int) string {
	port := strconv.Itoa(sentinelContainerPort)
	if len(rf.Spec.Sentinel.StaticResources) > index {
		port = strconv.Itoa(rf.Spec.Sentinel.StaticResources[index].Port)
	}
	return port
}

func GetRedisPortFromSpecByIndex(rf *roav1.Redis, index int) string {
	port := strconv.Itoa(redisContainerPort)
	if len(rf.Spec.Redis.StaticResources) > index {
		port = strconv.Itoa(rf.Spec.Redis.StaticResources[index].Port)
	}
	return port
}

func GetExporterPortFromSpec(rf *roav1.Redis) string {
	port := strconv.Itoa(exporterContainerPort)
	if rf.Spec.Exporter.StaticResource.Port != 0 {
		port = strconv.Itoa(rf.Spec.Exporter.StaticResource.Port)
	}
	return port
}

func GetQuorum(rf *roav1.Redis) int32 {
	return rf.Spec.Sentinel.Replicas/2 + 1
}

func IsNeedAutoFailover(rf *roav1.Redis) bool {
	if HasNoHostNetwork(rf) {
		return true
	}
	return false
}

func HasNoHostNetwork(rf *roav1.Redis) bool {
	if !rf.Spec.Redis.HostNetwork {
		return true
	}
	if !rf.Spec.Sentinel.HostNetwork {
		return true
	}
	return false
}
