package redis_client

import (
	"fmt"
	"github.com/go-logr/logr"
)

func Info(log logr.Logger, msg string, redisParam RedisParam) {
	log.Info(msg, "nameSpace", redisParam.NameSpace, "name", redisParam.Name)
}

func Error(log logr.Logger, err error, msg string, redisParam RedisParam) {
	log.Error(err, msg, "nameSpace", redisParam.NameSpace, "name", redisParam.Name)
}

func Error2(msg string, namespace, name string) {
	fmt.Println(msg, "nameSpace", namespace, "name", name)
}
