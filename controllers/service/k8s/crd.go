package k8s

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	roav1 "github.com/zhizuqiu/redis-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CRD interface {
	GetOnly(req ctrl.Request) (*roav1.Redis, error)
	Get(req ctrl.Request) (*roav1.Redis, error)
	UpdateRedisConfigStatus(redis *roav1.Redis, currentStatus roav1.RedisConfig) error
	UpdateSentinelConfigStatus(redis *roav1.Redis, currentStatus roav1.SentinelConfig) error
	UpdateRedisPasswordStatus(redis *roav1.Redis, currentStatus roav1.RedisPassword) error
	UpdateSentinelPasswordStatus(redis *roav1.Redis, currentStatus roav1.RedisPassword) error
}

type CRDService struct {
	KubeClient client.Client
	Log        logr.Logger
}

func NewCRDService(kubeClient client.Client, log logr.Logger) *CRDService {
	log = log.WithValues("service", "k8s.CRDService")
	return &CRDService{
		KubeClient: kubeClient,
		Log:        log,
	}
}

func (r *CRDService) GetOnly(req ctrl.Request) (*roav1.Redis, error) {
	Info2(r.Log, "start fetch Redis yaml...", req)

	ctx := context.Background()
	// Load the Redis by name
	redis := &roav1.Redis{}
	err := r.KubeClient.Get(ctx, req.NamespacedName, redis)

	PrintOBJ("Redis yaml", redis, redis)

	return redis, err
}

func (r *CRDService) Get(req ctrl.Request) (*roav1.Redis, error) {

	redis, err := r.GetOnly(req)
	if err != nil {
		Error2(r.Log, err, "unable to fetch Redis", req)
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		// return nil, client.IgnoreNotFound(err)
		return nil, err
	}
	if redis.GetObjectMeta().GetDeletionTimestamp() != nil {
		return nil, errors.New("this crd has been deleted")
	}

	return redis, nil
}

func (r *CRDService) UpdateRedisConfigStatus(redis *roav1.Redis, currentStatus roav1.RedisConfig) error {
	redis.Status.Redis.RedisCustomConfig = currentStatus
	if err := r.KubeClient.Status().Update(context.Background(), redis); err != nil {
		return err
	}
	return nil
}

func (r *CRDService) UpdateSentinelConfigStatus(redis *roav1.Redis, currentStatus roav1.SentinelConfig) error {
	redis.Status.Sentinel.SentinelCustomConfig = currentStatus
	if err := r.KubeClient.Status().Update(context.Background(), redis); err != nil {
		return err
	}
	return nil
}

func (r *CRDService) UpdateRedisPasswordStatus(redis *roav1.Redis, currentStatus roav1.RedisPassword) error {
	redis.Status.Redis.RedisPassword = currentStatus
	if err := r.KubeClient.Status().Update(context.Background(), redis); err != nil {
		return err
	}
	return nil
}

func (r *CRDService) UpdateSentinelPasswordStatus(redis *roav1.Redis, currentStatus roav1.RedisPassword) error {
	redis.Status.Sentinel.SentinelPassword = currentStatus
	if err := r.KubeClient.Status().Update(context.Background(), redis); err != nil {
		return err
	}
	return nil
}
