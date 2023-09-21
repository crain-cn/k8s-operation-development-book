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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	appsv1 "kubebuilder/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	//新增引用包
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
)

// ZhAppReconciler reconciles a ZhApp object
type ZhAppReconciler struct {
	client.Client
	ClientSet *kubernetes.Clientset //新增ClientSet参数
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.app.zhihu.com,resources=zhapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.app.zhihu.com,resources=zhapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.app.zhihu.com,resources=zhapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ZhApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *ZhAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	//获取CRD资源信息
	zhapp := &appsv1.ZhApp{}
	err := r.Client.Get(ctx, req.NamespacedName, zhapp)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if zhapp.DeletionTimestamp != nil {
		//删除Deployment逻辑
		err = r.ClientSet.AppsV1().Deployments(zhapp.Namespace).Delete(ctx, zhapp.Name, metav1.DeleteOptions{})
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	//判断是否存在Deployment
	deployment, _ := r.ClientSet.AppsV1().Deployments(zhapp.Namespace).Get(ctx, zhapp.Name, metav1.GetOptions{})
	if deployment.Name != "" {
		return ctrl.Result{}, nil
	}


	// 创建Deployment逻辑
	deploy := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: zhapp.Name,
			Namespace: zhapp.Namespace,
		},
		Spec: v1.DeploymentSpec{
			Replicas: pointer.Int32(zhapp.Spec.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": zhapp.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": zhapp.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name: "main",
							Image: zhapp.Spec.Image,
						},
					},
				},
			},
		},
	}

	_, err = r.ClientSet.AppsV1().Deployments(zhapp.Namespace).Create(ctx, deploy, metav1.CreateOptions{})
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZhAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.ZhApp{}).
		Complete(r)
}
