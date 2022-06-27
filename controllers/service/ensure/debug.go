package ensure

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
)

var EnsureDebug = false

func Info(log logr.Logger, msg string, rf *roav1.Redis) {
	log.Info(msg, "nameSpace", rf.Namespace, "name", rf.Name)
}

func Error(log logr.Logger, err error, msg string, rf *roav1.Redis) {
	log.Error(err, msg, "nameSpace", rf.Namespace, "name", rf.Name)
}

func PrintOBJ(name string, rf *roav1.Redis, obj interface{}) {
	if EnsureDebug {
		fmt.Println("---")
		fmt.Println("nameSpace:" + rf.Namespace + ",name:" + rf.Name + "," + name + ":")
		b, _ := json.Marshal(obj)
		fmt.Println(string(b))
		fmt.Println("---")
	}
}

func PrintOBJDebug(name string, rf *roav1.Redis, obj interface{}) {
	fmt.Println("---")
	fmt.Println("nameSpace:" + rf.Namespace + ",name:" + rf.Name + "," + name + ":")
	b, _ := json.Marshal(obj)
	fmt.Println(string(b))
	fmt.Println("---")
}
