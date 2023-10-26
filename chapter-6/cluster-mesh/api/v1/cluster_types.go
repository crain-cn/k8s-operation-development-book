/*
Copyright 2023 zhaoyu.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// K8S集群类型
type ClusterType string

const (
	IDC        ClusterType = "std-idc"   	// IDC集群
	AliyunACK  ClusterType = "ali-ack"		// 阿里云ACK
	AliyunASK  ClusterType = "ali-ask"		// 阿里云ASK
	TencentTke ClusterType = "tencent-tke"  // 腾讯云TKE
	TencentEks ClusterType = "tencent-eks"	// 腾讯云EKS
	HuaweiCce  ClusterType = "huawei-cce"	// 华为CCE
	BaiduEks   ClusterType = "baidu-cce" 	// 百度CCE
	KsyunEks   ClusterType = "ksyun-kce" 	// 金山云KCE
)

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// 集群名
	Name string `json:"name"`

	// 集群描述
	Description string `json:"description"`

	// 集群apiserver
	ApiServer string `json:"apiServer"`

	// 集群kubeconfig
	KubeConfig string `json:"kubeConfig"`

	// 是否为CI集群(构建任务执行集群)
	CiCluster bool `json:"ciCluster"`

	// 集群类型
	ClusterType ClusterType `json:"clusterType"`

	// 是否开启从主集群同步数据
	DataSyncEnable bool `json:"dataSyncEnable"`

	// 是否为主集群(元数据集群)
	MainCluster bool `json:"mainCluster"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}
// +genclient
// +genclient:nonNamespaced
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
