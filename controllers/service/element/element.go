package element

import (
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Element struct {
	NeedReLoad       bool
	NeedReCheckError []error
	Req              ctrl.Request
	Redis            *roav1.Redis
	OwnerRefs        []metav1.OwnerReference
}
