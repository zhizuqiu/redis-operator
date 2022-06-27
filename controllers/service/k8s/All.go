package k8s

import (
	"context"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type All interface {
	Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error
	Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error
	Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error
	UpdateStatus(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error
}

type AllService struct {
	KubeClient client.Client
	Log        logr.Logger
}

func NewAllService(kubeClient client.Client, log logr.Logger) *AllService {
	log = log.WithValues("service", "k8s.All")
	return &AllService{
		KubeClient: kubeClient,
		Log:        log,
	}
}

func (a *AllService) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if err := a.KubeClient.Create(ctx, obj, opts...); err != nil {
		return err
	}
	return nil
}

func (a *AllService) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if err := a.KubeClient.Delete(ctx, obj, opts...); err != nil {
		return err
	}
	return nil
}

func (a *AllService) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if err := a.KubeClient.Update(ctx, obj, opts...); err != nil {
		return err
	}
	return nil
}

func (a *AllService) UpdateStatus(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if err := a.KubeClient.Status().Update(ctx, obj, opts...); err != nil {
		return err
	}
	return nil
}
