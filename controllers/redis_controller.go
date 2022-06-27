/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/zhizuqiu/redis-operator/controllers/service/element"
	"github.com/zhizuqiu/redis-operator/controllers/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	runt "runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	componentv1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
)

// RedisReconciler reconciles a Redis object
type RedisReconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	RedisHandler *RedisHandler
}

var (
	ErrorRequeueAfter  = 10 * time.Second
	NormalRequeueAfter = 30 * time.Second
)

// +kubebuilder:rbac:groups=component.zhizuqiu,resources=redis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=component.zhizuqiu,resources=redis/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=statefulsets/status,verbs=get
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=deployments/status,verbs=get
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods/status,verbs=get
// +kubebuilder:rbac:groups="",resources=pods/exec,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps/status,verbs=get
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets/status,verbs=get
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims/status,verbs=get

func (r *RedisReconciler) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {

	Info2(r.Log, "----------------------", req)

	var m runt.MemStats
	runt.ReadMemStats(&m)
	mb := 1024 * 1024.0
	logstr := fmt.Sprintf("\nAlloc = %vMb\tTotalAlloc = %vMb\tSys = %vMb\t NumGC = %v\n", float64(m.Alloc)/mb, float64(m.TotalAlloc)/mb, float64(m.Sys)/mb, m.NumGC)
	fmt.Println(logstr)

	redis, err := r.RedisHandler.K8sServices.GetOnly(req)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			r.Log.Info("Redis resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		r.Log.Error(err, "Failed to get Redis.")
		return ctrl.Result{}, err
	}

	// Create owner refs so the objects manager by this handler have ownership to the
	// received RF.
	oRefs := r.RedisHandler.createOwnerReferences(redis)

	el := element.Element{
		NeedReLoad: false,
		Req:        req,
		Redis:      redis,
		OwnerRefs:  oRefs,
	}

	isRedisMarkedToBeDeleted := el.Redis.GetDeletionTimestamp() != nil
	if isRedisMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(el.Redis, util.RedisFinalizer) {
			// Run finalization logic for RedisFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeRedis(r.Log, el); err != nil {
				Error(r.Log, err, "finalizeRedis error!", redis)
				return ctrl.Result{RequeueAfter: ErrorRequeueAfter}, nil
			}

			// Remove RedisFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(el.Redis, util.RedisFinalizer)
			err := r.Update(context.Background(), el.Redis)
			if err != nil {
				Error(r.Log, err, "Remove RedisFinalizer, Update CR error!", redis)
				return ctrl.Result{RequeueAfter: ErrorRequeueAfter}, nil
			}
		}
		return ctrl.Result{RequeueAfter: NormalRequeueAfter}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(el.Redis, util.RedisFinalizer) {
		controllerutil.AddFinalizer(el.Redis, util.RedisFinalizer)
		err = r.Update(context.Background(), el.Redis)
		if err != nil {
			Error(r.Log, err, "Add finalizer for this CR, Update CR error!", redis)
			return ctrl.Result{RequeueAfter: ErrorRequeueAfter}, nil
		}
	}

	err = el.Redis.Check()
	if err != nil {
		Error(r.Log, err, "Redis.Check error!", redis)
		return ctrl.Result{RequeueAfter: ErrorRequeueAfter}, nil
	}

	el, err = r.Ensure(el)
	if err != nil {
		Error(r.Log, err, "Ensure error!", redis)
		return ctrl.Result{RequeueAfter: ErrorRequeueAfter}, nil
	}

	fmt.Println("after ensure needReLoad:")
	fmt.Println(el.NeedReLoad)

	el, err = r.CheckAndHeal(el)
	if err != nil {
		Error(r.Log, err, "CheckAndHeal error!", redis)
		return ctrl.Result{RequeueAfter: ErrorRequeueAfter}, nil
	}
	if len(el.NeedReCheckError) > 0 {
		Error(r.Log, err, "len(el.NeedReCheckError) > 0, wait next reconcile", redis)
		return ctrl.Result{RequeueAfter: ErrorRequeueAfter}, nil
	}

	_, err = r.CheckCluster(el, true)
	if err != nil {
		Error(r.Log, err, "CheckCluster error!", redis)
		return ctrl.Result{RequeueAfter: ErrorRequeueAfter}, nil
	}

	fmt.Println("Reconcile over")

	return ctrl.Result{RequeueAfter: NormalRequeueAfter}, nil
}

func (r *RedisReconciler) finalizeRedis(reqLogger logr.Logger, el element.Element) error {
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.

	el, err := r.DeleteEnsure(el)
	if err != nil {
		return err
	}

	reqLogger.Info("Successfully finalized redis")
	return nil
}

func (r *RedisReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&componentv1.Redis{}).
		Complete(r)
}
