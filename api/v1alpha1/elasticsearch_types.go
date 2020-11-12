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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ElasticsearchStatus defines the observed state of Elasticsearch
type ElasticsearchStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Elasticsearch is the Schema for the elasticsearches API
type Elasticsearch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticsearchSpec   `json:"spec,omitempty"`
	Status ElasticsearchStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ElasticsearchList contains a list of Elasticsearch
type ElasticsearchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Elasticsearch `json:"items"`
}

// ElasticsearchSpec defines the desired state of Elasticsearch
type ElasticsearchSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	//Master Node Replicas
	MasterReplicas int32 `json:"master-replicas"`

	//Client Node Replicas
	ClientReplicas int32 `json:"client-replicas"`

	//Hot Data Node Replica
	HotDataReplicas int32 `json:"hot-data-replicas"`

	//Warm Data Node Replica
	WarmDataReplicas int32 `json:"warm-data-replicas"`

	//NodeSelector for the pod to be eligible to run on a node (ex. Hot-Warm Architecture)
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	//Annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	//Disk Size of Hot Data Node
	HotDataDiskSize string `json:"hot-data-volume"`

	//Disk Size of Warm Data Node
	WarmDataDiskSize string `json:"warm-data-volume"`

	//Elasticsearch Image
	ElasticsearchImage string `json:"elasticsearch-image"`

	//Elasticsearch Cluster Name
	ElasticsearchClusterName string `json:"elasticsearch-cluster-name"`
	//MasterJavaOpt
	MasterJavaOpts string `json:"master-javaOpts"`

	//ClientJavaOpt
	ClientJavaOpts string `json:"client-javaOpts"`

	//HotDataJavaOpt
	HotDataJavaOpts string `json:"hot-data-javaOpts"`

	//WarmDataJavaOpt
	WarmDataJavaOpts string `json:"warm-data-javaOpts"`

	//Cerebro
	Cerebro Cerebro `json:"cerebro"`

	//Kibana
	Kibana Kibana `json:"kibana"`

	//Curator
	Curator Curator `json:"curator"`
}

// Kibana properties (Optional)
type Kibana struct {
	// Defines the image to use for deploying kibana
	Image string `json:"image"`
}

// Cerebro properties (Optional)
type Cerebro struct {
	// Defines the image to use for deploying Cerebro
	Image string `json:"image"`
}

// Curator properties (Optional)
type Curator struct {
	// Defines the image to use for deploying Curator
	Image string `json:"image"`
}

func init() {
	SchemeBuilder.Register(&Elasticsearch{}, &ElasticsearchList{})
}
