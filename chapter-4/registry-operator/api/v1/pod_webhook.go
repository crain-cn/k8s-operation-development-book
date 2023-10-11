/*
Copyright 2023.

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
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"setting.zhihu.com/generated/clientset/versioned"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"strings"
)

// log is for logging in this package.
var podlog = logf.Log.WithName("pod-resource")

// 定义核心组件pod的webhook的主struct
type PodWebhookMutate struct {
	ClientSet *versioned.Clientset
	decoder   *admission.Decoder
}

// +kubebuilder:webhook:path=/mutate-core-v1-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups=core,resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io,admissionReviewVersions=v1
func (a *PodWebhookMutate) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}
	err := a.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// 获得CRD资源列表
	list, err := a.ClientSet.RepoV1beta1().Repos().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// 更新镜像地址
	for _, item := range list.Items {
		for i := 0; i < len(pod.Spec.Containers); i++ {
			if strings.Contains(pod.Spec.Containers[i].Image, item.Spec.Source) {
				pod.Spec.Containers[i].Image = strings.Replace(pod.Spec.Containers[i].Image, item.Spec.Source, item.Spec.Destination, -1)
			}
		}
	}

	// TODO: 变量marshaledPod是一个Map，可以直接修改pod的一些属性
	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// 打印
	fmt.Println("======================================================")
	fmt.Println(string(marshaledPod))
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

//
//func (r *Pod) SetupWebhookWithManager(mgr ctrl.Manager) error {
//	return ctrl.NewWebhookManagedBy(mgr).
//		For(r).
//		Complete()
//}
//
//// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
//
//var _ webhook.Defaulter = &Pod{}
//
//// Default implements webhook.Defaulter so a webhook will be registered for the type
//func (r *Pod) Default() {
//	podlog.Info("default", "name", r.Name)
//
//	// TODO(user): fill in your defaulting logic.
//}
