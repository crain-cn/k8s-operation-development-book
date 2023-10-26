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

package controllers

import (
	"context"
	"errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	clustersv1 "zhihu.com/cluster-mesh/api/v1"
	"zhihu.com/cluster-mesh/generated/clientset/versioned"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	sync.RWMutex
	client.Client
	Clientset *versioned.Clientset
	Scheme *runtime.Scheme
	Clientsets map[string]*versioned.Clientset
}

// 解析kubeconfig为*rest.Config对象
func LoadConfig(kubeconfig string) (*rest.Config, error) {
	cfg, err := clientcmd.Load([]byte(kubeconfig))
	if err != nil {
		return nil, err
	}
	for context := range cfg.Contexts {
		contextCfg, err := clientcmd.NewNonInteractiveClientConfig(*cfg, context, &clientcmd.ConfigOverrides{}, nil).ClientConfig()
		if err != nil {
			return nil, err
		}
		contextCfg.QPS = 100
		contextCfg.Burst = 1000
		return contextCfg, nil
	}
	return nil, errors.New("invalid kubeconfig")
}

//+kubebuilder:rbac:groups=clusters.zhihu.com,resources=clusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.zhihu.com,resources=clusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.zhihu.com,resources=clusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//获取CRD资源信息
	cluster ,err :=  r.Clientset.ClustersV1().Clusters().Get(ctx, req.NamespacedName.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// 获得各个集群资源
	clusterList , err :=  r.Clientset.ClustersV1().Clusters().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 初始化Cluster对应CRD的Clientset集合
	for _, clusterInfo := range clusterList.Items {
		config, err := LoadConfig(clusterInfo.Spec.KubeConfig)
		if err != nil {
			//可以处理失败的集群
			continue
		}

		clusterClient, err  := versioned.NewForConfig(config)
		if err != nil {
			//可以处理失败的集群
			continue
		}
		r.Clientsets = make(map[string]*versioned.Clientset, 0)
		r.RWMutex.Lock()
		r.Clientsets[clusterInfo.Name] = clusterClient
		r.RWMutex.Unlock()
	}

	if cluster.DeletionTimestamp != nil{
		// 删除各个集群Cluster
		for _, clusterClient := range r.Clientsets {
			// 判断是否存在，存在则发起删除
			info, _ := clusterClient.ClustersV1().Clusters().Get(context.TODO(), cluster.Name, metav1.GetOptions{})
			if info != nil {
				clusterClient.ClustersV1().Clusters().Delete(context.TODO(), cluster.Name, metav1.DeleteOptions{})
			}
		}
		return ctrl.Result{}, nil
	}

	// 更新和增加
	for _, clusterClient := range r.Clientsets {
		// 判断是否存在，存在则走更新逻辑，不存在则走添加逻辑
		info, _ := clusterClient.ClustersV1().Clusters().Get(context.TODO(), cluster.Name, metav1.GetOptions{})
		if info != nil {
			if info.Spec.KubeConfig != cluster.Spec.KubeConfig ||
				info.Spec.MainCluster != cluster.Spec.MainCluster ||
				info.Spec.CiCluster != cluster.Spec.CiCluster ||
				info.Spec.DataSyncEnable != cluster.Spec.DataSyncEnable ||
				info.Spec.ApiServer != cluster.Spec.ApiServer ||
				info.Spec.MainCluster != cluster.Spec.MainCluster ||
				info.Spec.Description != cluster.Spec.Description ||
				info.Spec.ClusterType != cluster.Spec.ClusterType {
				clusterClient.ClustersV1().Clusters().Update(context.TODO(), cluster, metav1.UpdateOptions{})
			}
		} else {
			clusterClient.ClustersV1().Clusters().Create(context.TODO(), cluster, metav1.CreateOptions{})
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1.Cluster{}).
		Complete(r)
}
