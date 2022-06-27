package check

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	"github.com/zhizuqiu/redis-operator/controllers/service/redis_client"
)

var CheckDebug = true

func Info(log logr.Logger, msg string, rf *roav1.Redis) {
	log.Info(msg, "nameSpace", rf.Namespace, "name", rf.Name)
}

func Info2(log logr.Logger, msg string, redisParam redis_client.RedisParam) {
	log.Info(msg, "nameSpace", redisParam.NameSpace, "name", redisParam.Name)
}

func Error(log logr.Logger, err error, msg string, rf *roav1.Redis) {
	log.Error(err, msg, "nameSpace", rf.Namespace, "name", rf.Name)
}

func PrintOBJ(name string, rf *roav1.Redis, obj interface{}) {
	if CheckDebug {
		fmt.Println("nameSpace:" + rf.Namespace + ",name:" + rf.Name + "," + name + ":")
		b, _ := json.Marshal(obj)
		fmt.Println(string(b))
	}
}
