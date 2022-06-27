package k8s

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var K8sDebug = true

func Info(log logr.Logger, msg string, rf *roav1.Redis) {
	log.Info(msg, "nameSpace", rf.Namespace, "name", rf.Name)
}

func Info2(log logr.Logger, msg string, req ctrl.Request) {
	log.Info(msg, "nameSpace", req.Namespace, "name", req.Name)
}

func Error(log logr.Logger, err error, msg string, rf *roav1.Redis) {
	log.Error(err, msg, "nameSpace", rf.Namespace, "name", rf.Name)
}

func Error3(log logr.Logger, err error, msg string) {
	log.Error(err, msg)
}

func Error2(log logr.Logger, err error, msg string, req ctrl.Request) {
	log.Error(err, msg, "nameSpace", req.Namespace, "name", req.Name)
}

func PrintOBJ(name string, rf *roav1.Redis, obj interface{}) {
	if K8sDebug {
		fmt.Println("nameSpace:" + rf.Namespace + ",name:" + rf.Name + "," + name + ":")
		b, _ := json.Marshal(obj)
		fmt.Println(string(b))
	}
}
