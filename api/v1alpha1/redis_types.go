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
	"errors"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisSpec defines the desired state of Redis
type RedisSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Redis    RedisSettings    `json:"redis,omitempty"`
	Sentinel SentinelSettings `json:"sentinel,omitempty"`
	Exporter Exporter         `json:"exporter,omitempty"`
	Auth     AuthSettings     `json:"auth,omitempty"`
}

// RedisSettings defines the specification of the redis cluster
type RedisSettings struct {
	Resources    corev1.ResourceRequirements `json:"resources,omitempty"`
	CustomConfig []string                    `json:"customConfig,omitempty"`

	Image                  string                        `json:"image"`
	UpdateStrategy         v1.StatefulSetUpdateStrategy  `json:"updateStrategy,omitempty"`
	ImagePullPolicy        corev1.PullPolicy             `json:"imagePullPolicy,omitempty"`
	Replicas               int32                         `json:"replicas"`
	CustomCommandRenames   []RedisCommandRename          `json:"customCommandRenames,omitempty"`
	Command                []string                      `json:"command,omitempty"`
	Storage                RedisStorage                  `json:"storage,omitempty"`
	StorageLog             RedisStorage                  `json:"storageLog,omitempty"`
	Affinity               *corev1.Affinity              `json:"affinity,omitempty"`
	SecurityContext        *corev1.PodSecurityContext    `json:"securityContext,omitempty"`
	ImagePullSecrets       []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Tolerations            []corev1.Toleration           `json:"tolerations,omitempty"`
	NodeSelector           map[string]string             `json:"nodeSelector,omitempty"`
	PodAnnotations         map[string]string             `json:"podAnnotations,omitempty"`
	HostNetwork            bool                          `json:"hostNetwork,omitempty"`
	DNSPolicy              corev1.DNSPolicy              `json:"dnsPolicy,omitempty"`
	PriorityClassName      string                        `json:"priorityClassName,omitempty"`
	EnabledPodAntiAffinity bool                          `json:"enabledPodAntiAffinity,omitempty"`
	StaticResources        []StaticResource              `json:"staticResources,omitempty"`
}

type StaticResource struct {
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
}

type Exporter struct {
	Enabled         bool              `json:"enabled,omitempty"`
	Image           string            `json:"image,omitempty"`
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	HostNetwork     bool              `json:"hostNetwork,omitempty"`
	StaticResource  StaticResource    `json:"staticResource,omitempty"`
	Affinity        *corev1.Affinity  `json:"affinity,omitempty"`
}

// RedisCommandRename defines the specification of a "rename-command" configuration option
type RedisCommandRename struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

// RedisStorage defines the structure used to store the Redis Data
type RedisStorage struct {
	KeepAfterDeletion     bool                          `json:"keepAfterDeletion,omitempty"`
	EmptyDir              *corev1.EmptyDirVolumeSource  `json:"emptyDir,omitempty"`
	PersistentVolumeClaim *corev1.PersistentVolumeClaim `json:"persistentVolumeClaim,omitempty"`
}

// SentinelSettings defines the specification of the sentinel cluster
type SentinelSettings struct {
	Resources    corev1.ResourceRequirements `json:"resources,omitempty"`
	CustomConfig []string                    `json:"customConfig,omitempty"`
	Service      SentinelService             `json:"service,omitempty"`

	Image                  string                        `json:"image"`
	UpdateStrategy         v1.StatefulSetUpdateStrategy  `json:"updateStrategy,omitempty"`
	ImagePullPolicy        corev1.PullPolicy             `json:"imagePullPolicy,omitempty"`
	Replicas               int32                         `json:"replicas"`
	Command                []string                      `json:"command,omitempty"`
	Storage                RedisStorage                  `json:"storage,omitempty"`
	StorageLog             RedisStorage                  `json:"storageLog,omitempty"`
	Affinity               *corev1.Affinity              `json:"affinity,omitempty"`
	SecurityContext        *corev1.PodSecurityContext    `json:"securityContext,omitempty"`
	ImagePullSecrets       []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Tolerations            []corev1.Toleration           `json:"tolerations,omitempty"`
	NodeSelector           map[string]string             `json:"nodeSelector,omitempty"`
	PodAnnotations         map[string]string             `json:"podAnnotations,omitempty"`
	HostNetwork            bool                          `json:"hostNetwork,omitempty"`
	DNSPolicy              corev1.DNSPolicy              `json:"dnsPolicy,omitempty"`
	PriorityClassName      string                        `json:"priorityClassName,omitempty"`
	EnabledPodAntiAffinity bool                          `json:"enabledPodAntiAffinity,omitempty"`
	StaticResources        []StaticResource              `json:"staticResources,omitempty"`
}

type SentinelService struct {
	Enabled            bool              `json:"enabled,omitempty"`
	ServiceAnnotations map[string]string `json:"serviceAnnotations,omitempty"`
}

// AuthSettings contains settings about auth
type AuthSettings struct {
	SecretPath string   `json:"secretPath,omitempty"`
	Password   Password `json:"password,omitempty"`
}

type Password struct {
	EncodeType PasswordEncodeType `json:"encodeType,omitempty"`
	Value      string             `json:"value,omitempty"`
}

type PasswordEncodeType string

var (
	BASE64 PasswordEncodeType = "base64"
	SM4    PasswordEncodeType = "sm4"
)

func GetDefaultPasswordEncodeType() PasswordEncodeType {
	return BASE64
}

// RedisStatus defines the observed state of Redis
type RedisStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Redis    RedisState    `json:"redis,omitempty"`
	Sentinel SentinelState `json:"sentinel,omitempty"`
	Exporter ExporterState `json:"exporter,omitempty"`
	State    State         `json:"state,omitempty"`
}

type State struct {
	Pods    map[string]PodState `json:"pods,omitempty"`
	Phase   corev1.PodPhase     `json:"phase,omitempty"`
	Ready   bool                `json:"ready,omitempty"`
	Cluster bool                `json:"cluster,omitempty"`
}

type PodState struct {
	Name          string          `json:"name,omitempty"`
	Role          string          `json:"role,omitempty"`
	Phase         corev1.PodPhase `json:"phase,omitempty"`
	HostIP        string          `json:"hostIP,omitempty"`
	PodIP         string          `json:"podIP,omitempty"`
	ContainerPort int32           `json:"containerPort,omitempty"`
	PodIPs        []corev1.PodIP  `json:"podIPs,omitempty"`
	StartTime     *metav1.Time    `json:"startTime,omitempty"`
}

type RedisState struct {
	RedisCustomConfig RedisConfig   `json:"redisCustomConfig,omitempty"`
	RedisPassword     RedisPassword `json:"redisPassword,omitempty"`
}

type SentinelState struct {
	SentinelCustomConfig SentinelConfig `json:"sentinelCustomConfig,omitempty"`
	SentinelPassword     RedisPassword  `json:"sentinelPassword,omitempty"`
}

type RedisPassword struct {
	Md5 string `json:"md5,omitempty"`
}

type RedisConfig struct {
	Md5 string `json:"md5,omitempty"`
}

type SentinelConfig struct {
	Md5 string `json:"md5,omitempty"`
}

type RedisStatusItem struct {
	// +optional
	Status RedisStatusEnum `json:"status,omitempty"`
}

type RedisStatusEnum string

var (
	Desired RedisStatusEnum = "Desired"
	Pending RedisStatusEnum = "Pending"
)

type ExporterState struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.state.phase",description="Phase of instances in Redis"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.state.ready",description="Ready status of instances in Redis"
// +kubebuilder:printcolumn:name="Cluster",type="boolean",JSONPath=".status.state.cluster",description="cluster status of instances in Redis"
// +kubebuilder:printcolumn:name="Redis_Replicas",type="integer",JSONPath=".spec.redis.replicas",description="Redis Replicas of instances in Redis"
// +kubebuilder:printcolumn:name="Sentinel_Replicas",type="integer",JSONPath=".spec.sentinel.replicas",description="Sentinel Replicas of instances in Redis"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status

// Redis is the Schema for the redis API
type Redis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisSpec   `json:"spec,omitempty"`
	Status RedisStatus `json:"status,omitempty"`
}

func (r *Redis) Check() error {
	if r.Spec.Redis.HostNetwork {
		if r.Spec.Redis.Replicas > 0 {
			if r.Spec.Redis.StaticResources == nil {
				return errors.New("Spec.Redis.StaticResources==nil")
			}
			if len(r.Spec.Redis.StaticResources) < int(r.Spec.Redis.Replicas) {
				return errors.New("len(Spec.Redis.StaticResources) < Spec.Redis.Replicas")
			}
		}
	}
	if r.Spec.Sentinel.HostNetwork {
		if r.Spec.Sentinel.Replicas > 0 {
			if r.Spec.Sentinel.StaticResources == nil {
				return errors.New("Spec.Sentinel.StaticResources==nil")
			}
			if len(r.Spec.Sentinel.StaticResources) < int(r.Spec.Sentinel.Replicas) {
				return errors.New("len(Spec.Sentinel.StaticResources) < Spec.Sentinel.Replicas")
			}
		}
	}
	if !r.Spec.Redis.HostNetwork || !r.Spec.Sentinel.HostNetwork {
		if r.Spec.Exporter.HostNetwork {
			return errors.New("(!Spec.Redis.HostNetwork || !Spec.Sentinel.HostNetwork) when Spec.Exporter.HostNetwork=true")
		}
	}

	return nil
}

// +kubebuilder:object:root=true

// RedisList contains a list of Redis
type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redis `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Redis{}, &RedisList{})
}
