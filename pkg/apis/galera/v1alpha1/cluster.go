// Copyright 2016 The galera-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// TODO: This is where we put the galera image
	DefaultBaseImage = "quay.io/beekhof/centos"
	DefaultVersion   = "0.0.1"
)

var (
	// TODO: move validation code into separate package.
	ErrBackupUnsetRestoreSet = errors.New("spec: backup policy must be set if restore policy is set")
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ReplicatedStatefulSetList is a list of galera clusters.
type ReplicatedStatefulSetList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReplicatedStatefulSet `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReplicatedStatefulSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ClusterSpec   `json:"spec"`
	Status            ClusterStatus `json:"status"`
}

func (c *ReplicatedStatefulSet) AsOwner() metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       ReplicatedStatefulSetResourceKind,
		Name:       c.Name,
		UID:        c.UID,
		Controller: &trueVar,
	}
}

type ClusterSpec struct {
	// Size is the expected size of the galera cluster.
	// The galera-operator will eventually make the size of the running
	// cluster equal to the expected size.
	// The vaild range of the size is from 1 to 7.
	Size int `json:"size"`

	MaxPrimaries int `json:"maxPrimaries"`
	// Number of instances to deploy for a Prometheus deployment.
	//Replicas *int32 `json:"replicas,omitempty"`

	// BaseImage is the base galera image name that will be used to launch
	// galera clusters. This is useful for private registries, etc.
	//
	// If image is not set, default is gcr.io/galera-development/etcd
	BaseImage string `json:"baseImage"`

	// Version is the expected version of the galera cluster.
	// The galera-operator will eventually make the galera cluster version
	// equal to the expected version.
	//
	// The version must follow the [semver]( http://semver.org) format, for example "3.2.10".
	// Only galera released versions are supported: https://github.com/coreos/etcd/releases
	//
	// If version is not set, default is "3.2.10".
	Version string `json:"version,omitempty"`

	StatusCommand         []string `json:"statusCommand"`
	SequenceCommand       []string `json:"sequenceCommand"`
	StartSeedCommand      []string `json:"startSeedCommand,omitempty"`
	StartPrimaryCommand   []string `json:"startPrimaryCommand"`
	StartSecondaryCommand []string `json:"startSecondaryCommand,omitempty"`
	StopCommand           []string `json:"stopCommand"`

	// An optional list of references to secrets in the same namespace
	// to use for pulling prometheus and alertmanager images from registries
	// see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// Paused is to pause the control of the operator for the galera cluster.
	Paused bool `json:"paused,omitempty"`

	// Pod defines the policy to create pod for the galera pod.
	//
	// Updating Pod does not take effect on any existing galera pods.
	Pod *PodPolicy `json:"pod,omitempty"`

	// Pod defines the policy to create pod for the galera pod.
	//
	// Updating Pod does not take effect on any existing galera pods.
	Service *ServicePolicy `json:"pod,omitempty"`

	// galera cluster TLS configuration
	TLS *TLSPolicy `json:"TLS,omitempty"`

	// Storage spec to specify how storage shall be used.
	//Storage *StorageSpec `json:"storage,omitempty"`
	VolumeClaimTemplate *v1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`

	// A selector to select which ConfigMaps to mount for loading rule files from.
	RuleSelector *metav1.LabelSelector `json:"ruleSelector,omitempty"`

	// Define resources requests and limits for single Pods.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount to use to run the
	// Prometheus Pods.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Secrets is a list of Secrets in the same namespace as the Prometheus
	// object, which shall be mounted into the Prometheus Pods.
	// The Secrets are mounted into /etc/prometheus/secrets/<secret-name>.
	// Secrets changes after initial creation of a Prometheus object are not
	// reflected in the running Pods. To change the secrets mounted into the
	// Prometheus Pods, the object must be deleted and recreated with the new list
	// of secrets.
	Secrets []string `json:"secrets,omitempty"`

	// If specified, the pod's scheduling constraints.
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// If specified, the pod's tolerations.
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
}

// ServicePolicy defines the policy to create service for the galera container.
type ServicePolicy struct {
	Name            string             `json:"name,omitempty"`
	SessionAffinity v1.ServiceAffinity `json:"sessionAffinity,omitempty"`
	Ports           []v1.ServicePort   `json:"ports,omitempty"`
}

// PodPolicy defines the policy to create pod for the galera container.
type PodPolicy struct {
	// Labels specifies the labels to attach to pods the operator creates for the
	// galera cluster.
	// "app" and "galera_*" labels are reserved for the internal use of the galera operator.
	// Do not overwrite them.
	Labels map[string]string `json:"labels,omitempty"`

	// NodeSelector specifies a map of key-value pairs. For the pod to be eligible
	// to run on a node, the node must have each of the indicated key-value pairs as
	// labels.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// AntiAffinity determines if the galera-operator tries to avoid putting
	// the galera members in the same cluster onto the same node.
	AntiAffinity bool `json:"antiAffinity,omitempty"`

	// Resources is the resource requirements for the galera container.
	// This field cannot be updated once the cluster is created.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// Tolerations specifies the pod's tolerations.
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`

	// List of environment variables to set in the galera container.
	// This is used to configure galera process. galera cluster cannot be created, when
	// bad environement variables are provided. Do not overwrite any flags used to
	// bootstrap the cluster (for example `--initial-cluster` flag).
	// This field cannot be updated.
	GaleraEnv []v1.EnvVar `json:"galeraEnv,omitempty"`

	// By default, kubernetes will mount a service account token into the galera pods.
	// AutomountServiceAccountToken indicates whether pods running with the service account should have an API token automatically mounted.
	AutomountServiceAccountToken *bool `json:"automountServiceAccountToken,omitempty"`

	Ports []v1.ContainerPort `json:"ports,omitempty"`

	LivenessProbe  v1.Probe `json:"livenessProbe,omitempty"`
	ReadinessProbe v1.Probe `json:"readinessProbe,omitempty"`
}

func (c *ClusterSpec) PodLabels() map[string]string {
	if c.Pod != nil {
		return c.Pod.Labels
	}
	return map[string]string{}
}

func (c *ClusterSpec) ServiceName(cname string) string {
	if c.Service != nil && c.Service.Name != "" {
		return c.Service.Name
	}
	return fmt.Sprintf("%s-svc", cname)
}

func (c *ClusterSpec) ServicePorts() []v1.ServicePort {
	if c.Service != nil && c.Service.Ports != nil {
		return c.Service.Ports
	}
	return []v1.ServicePort{
		{
			Name:       "web",
			Port:       9090,
			TargetPort: intstr.FromString("web"),
		},
	}
}

func (c *ClusterSpec) ContainerPorts() []v1.ContainerPort {
	if c.Pod != nil && c.Pod.Ports != nil {
		return c.Pod.Ports
	}
	return []v1.ContainerPort{
		{
			Name:          "web",
			ContainerPort: 9090,
			Protocol:      v1.ProtocolTCP},
	}
}

func (c *ClusterSpec) Validate() error {
	if c.TLS != nil {
		if err := c.TLS.Validate(); err != nil {
			return err
		}
	}

	if c.Pod != nil {
		for k := range c.Pod.Labels {
			if k == "app" || strings.HasPrefix(k, "galera_") {
				return errors.New("spec: pod labels contains reserved label")
			}
		}
	}
	return nil
}

// Cleanup cleans up user passed spec, e.g. defaulting, transforming fields.
// TODO: move this to admission controller
func (c *ClusterSpec) Cleanup() {
	if len(c.BaseImage) == 0 {
		c.BaseImage = DefaultBaseImage
	}

	if len(c.Version) == 0 {
		c.Version = DefaultVersion
	}

	c.Version = strings.TrimLeft(c.Version, "v")
}
