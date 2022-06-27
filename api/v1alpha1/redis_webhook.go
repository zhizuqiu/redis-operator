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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var redislog = logf.Log.WithName("redis-resource")

func (r *Redis) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-component-zhizuqiu-v1alpha1-redis,mutating=true,failurePolicy=fail,groups=component.zhizuqiu,resources=redis,verbs=create;update,versions=v1alpha1,name=mredis.kb.io

var _ webhook.Defaulter = &Redis{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Redis) Default() {
	redislog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-component-zhizuqiu-v1alpha1-redis,mutating=false,failurePolicy=fail,groups=component.zhizuqiu,resources=redis,versions=v1alpha1,name=vredis.kb.io

var _ webhook.Validator = &Redis{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Redis) ValidateCreate() error {
	redislog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Redis) ValidateUpdate(old runtime.Object) error {
	redislog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Redis) ValidateDelete() error {
	redislog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
