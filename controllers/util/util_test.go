package util

import (
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	"gopkg.in/yaml.v2"
	"testing"
)

var (
	redisInStr = `
spec:
  sentinel:
  redis:
  exporter:
  auth:
`
	redisInAPIVersion = "component.zhizuqiu/v1alpha1"
	redisInKind       = "Redis"

	redisInName      = "redis-sample"
	redisInNamespace = "redis-system"
	redisInLabel     = map[string]string{
		"instance_id":   "6dc60b614c954d369ae7458e98442529",
		"instance_name": "redis-sample",
		"product_id":    "redis",
		"region_id":     "beijing",
	}

	redisIn *roav1.Redis
)

func init() {
	_ = yaml.Unmarshal([]byte(redisInStr), &redisIn)

	redisIn.APIVersion = redisInAPIVersion
	redisIn.Kind = redisInKind

	redisIn.Name = redisInName
	redisIn.Namespace = redisInNamespace
	redisIn.Labels = redisInLabel
}

func TestGetRedisNameByIndex(t *testing.T) {
	expected := "redis-redis-sample-0"
	actual := GetRedisNameByIndex(redisIn, 0)
	if actual != expected {
		t.Fatalf("actual = %s; expected = %s", actual, expected)
	}
}

func TestGetRedisReadinessConfigMapName(t *testing.T) {
	expected := "redis-readiness-redis-sample"
	actual := GetRedisReadinessConfigMapName(redisIn)
	if actual != expected {
		t.Fatalf("actual = %s; expected = %s", actual, expected)
	}
}

func TestGetSentinelNameByIndex(t *testing.T) {
	expected := "sentinel-redis-sample-0"
	actual := GetSentinelNameByIndex(redisIn, 0)
	if actual != expected {
		t.Fatalf("actual = %s; expected = %s", actual, expected)
	}
}

func TestGetRedisConfigMapName(t *testing.T) {
	expected := "redis-redis-sample"
	actual := GetRedisConfigMapName(redisIn)
	if actual != expected {
		t.Fatalf("actual = %s; expected = %s", actual, expected)
	}
}

func TestGetRedisSlaveConfigMapName(t *testing.T) {
	expected := "redis-slave-redis-sample-0"
	actual := GetRedisConfigMapNameByIndex(redisIn, 0)
	if actual != expected {
		t.Fatalf("actual = %s; expected = %s", actual, expected)
	}
}

func TestGetSentinelSlaveConfigMapName(t *testing.T) {
	expected := "sentinel-slave-redis-sample-0"
	actual := GetSentinelConfigMapNameByIndex(redisIn, 0)
	if actual != expected {
		t.Fatalf("actual = %s; expected = %s", actual, expected)
	}
}

func TestGetExporterRootName(t *testing.T) {
	expected := "exporter-redis-sample"
	actual := GetExporterRootName(redisIn)
	if actual != expected {
		t.Fatalf("actual = %s; expected = %s", actual, expected)
	}
}

func TestGetSentinelServiceName(t *testing.T) {
	expected := "sentinel-redis-sample"
	actual := GetSentinelServiceName(redisIn)
	if actual != expected {
		t.Fatalf("actual = %s; expected = %s", actual, expected)
	}
}

func TestGetRedisConfigWritablePath(t *testing.T) {
	expected := "/data/conf/redis.conf"
	actual := GetRedisConfigWritablePath()
	if actual != expected {
		t.Fatalf("actual = %s; expected = %s", actual, expected)
	}
}

func TestGetRedisConfigPath(t *testing.T) {
	expected := "/redis/redis.conf"
	actual := GetRedisConfigPath()
	if actual != expected {
		t.Fatalf("actual = %s; expected = %s", actual, expected)
	}
}

func TestGetSentinelConfigWritablePath(t *testing.T) {
	expected := "/data/conf/sentinel.conf"
	actual := GetSentinelConfigWritablePath()
	if actual != expected {
		t.Fatalf("actual = %s; expected = %s", actual, expected)
	}
}

func TestGetSentinelConfigPath(t *testing.T) {
	expected := "/redis/sentinel.conf"
	actual := GetSentinelConfigPath()
	if actual != expected {
		t.Fatalf("actual = %s; expected = %s", actual, expected)
	}
}
