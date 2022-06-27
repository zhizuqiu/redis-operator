package util

import (
	"fmt"
)

const (
	redisConfigTemplate = `protected-mode no
pidfile /redis/redis.pid
dir /data/
loglevel notice
logfile /redislog/redis.log
appendonly yes
appendfilename "appendonly.aof"
client-output-buffer-limit normal 0 0 0
client-output-buffer-limit slave 0 0 0   
client-output-buffer-limit pubsub 33554432 8388608 60
repl-backlog-size 1048576000
tcp-keepalive 60
repl-timeout 300
slave-priority 50
timeout 600
{{- range .Spec.Redis.CustomCommandRenames}}
rename-command "{{.From}}" "{{.To}}"
{{- end}}
`

	sentinelConfigFile = `sentinel down-after-milliseconds mymaster 1000
sentinel failover-timeout mymaster 3000
sentinel parallel-syncs mymaster 2
protected-mode no
loglevel notice
logfile /redislog/redis.log
timeout 600`

	readinessContent = `ROLE="role"
   ROLE_MASTER="role:master"
   ROLE_SLAVE="role:slave"
   IN_SYNC="master_sync_in_progress:1"
   NO_MASTER="master_host:127.0.0.1"

	function getPass(){
		local password=$(cat /data/conf/redis.conf | grep requirepass | awk -F\" '{print $2}')
		echo "$password"
	}

	function getPort(){
		local port=$(cat /data/conf/redis.conf | grep port | awk '{print $2}')
		echo "$port"
	}

	REDIS_PASSWORD="$(getPass)"
	REDIS_PORT="$(getPort)"

   check_master(){
           exit 0
   }

   check_slave(){
           in_sync=$(redis-cli -p "${REDIS_PORT}" --no-auth-warning -a "${REDIS_PASSWORD}" info replication | grep $IN_SYNC | tr -d "\r" | tr -d "\n")
           no_master=$(redis-cli -p "${REDIS_PORT}" --no-auth-warning -a "${REDIS_PASSWORD}" info replication | grep $NO_MASTER | tr -d "\r" | tr -d "\n")

           if [ -z "$in_sync" ] && [ -z "$no_master" ]; then
                   exit 0
           fi

           exit 1
   }

   role=$(redis-cli -p "${REDIS_PORT}" --no-auth-warning -a "${REDIS_PASSWORD}" info replication | grep $ROLE | tr -d "\r" | tr -d "\n")

   case $role in
           $ROLE_MASTER)
                   check_master
                   ;;
           $ROLE_SLAVE)
                   check_slave
                   ;;
           *)
                   echo "unespected"
                   exit 1
   esac`

	redisReadinessVolumeName  = "redis-readiness-config"
	redisStorageVolumeName    = "redis-data"
	redisLogStorageVolumeName = "redis-log"
	redisConfigCopy           = "redis-config-copy"
	redisConfig               = "redis-config"
	sentinelConfigCopy        = "sentinel-config-copy"
	sentinelConfig            = "sentinel-config"
	graceTime                 = 30
)

const (
	// baseName               = "rc-"
	headlessServiceBaseName = "headless"

	baseName = ""
	// 用于区分不同实例，用于拼接name、label
	sentinelRootName = "sentinel"
	// container、port name
	sentinelName = "sentinel"
	// status.state.pods.{pod name}.role
	sentinelRoleName       = "sentinel"
	sentinelConfigFileName = "sentinel.conf"
	sentinelContainerPort  = 26379
	redisConfigFileName    = "redis.conf"
	// 用于区分不同实例，用于拼接name、label
	redisRootName = "redis"
	// container、port name
	redisName = "redis"
	// status.state.pods.{pod name}.role
	redisRoleName      = "redis"
	redisReadinessName = "redis-readiness"
	redisGroupName     = "mymaster"
	redisContainerPort = 6379
	// 用于区分不同实例，用于拼接name、label
	exporterRootName = "exporter"
	// container、port name
	exporterName = "exporter"
	// status.state.pods.{pod name}.role
	exporterRoleName             = "exporter"
	exporterContainerName        = "e-metrics"
	exporterContainerPort        = 9121
	exporterDefaultRequestCPU    = "25m"
	exporterDefaultLimitCPU      = "50m"
	exporterDefaultRequestMemory = "50Mi"
	exporterDefaultLimitMemory   = "100Mi"

	appLabel            = "redis-sentinel"
	hostnameTopologyKey = "kubernetes.io/hostname"

	appNameLabelKey         = "app.kubernetes.io/name"
	statefulSetNameLabelKey = "app.kubernetes.io/statefulset"
	appComponentLabelKey    = "app.kubernetes.io/component"
	appPartOfLabelKey       = "app.kubernetes.io/part-of"

	statefulSetPodLabelKey = "statefulset.kubernetes.io/pod-name"

	RedisFinalizer = "redis.component.zhizuqiu/finalizer"
)

const (
	redisConfWritableMountPath    = "/data/conf"
	redisConfMountPath            = "/redis"
	sentinelConfWritableMountPath = "/data/conf"
	sentinelConfMountPath         = "/redis"
)

const (
	defaultPeriodSeconds    = 10
	defaultSuccessThreshold = 1
	defaultFailureThreshold = 3
)

func GetRedisConfigWritablePath() string {
	return fmt.Sprintf(redisConfWritableMountPath+"/%s", redisConfigFileName)
}

func GetRedisConfigPath() string {
	return fmt.Sprintf(redisConfMountPath+"/%s", redisConfigFileName)
}

func GetSentinelConfigWritablePath() string {
	return fmt.Sprintf(sentinelConfWritableMountPath+"/%s", sentinelConfigFileName)
}

func GetSentinelConfigPath() string {
	return fmt.Sprintf(sentinelConfMountPath+"/%s", sentinelConfigFileName)
}
