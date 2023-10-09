package main

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 定义Handler
type MutatingHandler struct {
	Decoder *admission.Decoder
}

// webhook 回调函数
func (h *MutatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {

	reqJson, err := json.Marshal(req.AdmissionRequest)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	fmt.Println(string(reqJson))
	// 解析拦截的Pod内容
	pod := &corev1.Pod{}
	err = h.Decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	//获得原Pod内容
	originObj, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	//设置挂载Pod对应configmap
	pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
		Name: "hello-volume",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "hello-configmap", //挂载hello-configmap
				},
			},
		},
	})
	for i := 0; i < len(pod.Spec.Containers); i++ {
		pod.Spec.Containers[i].VolumeMounts = append(pod.Spec.Containers[i].VolumeMounts, corev1.VolumeMount{
			Name:      "hello-volume",
			MountPath: "/etc/config",
		})
	}
	//获得更新后的Pod内容
	newobj := pod.DeepCopy()
	currentObj, err := json.Marshal(newobj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	// 返回patch内容
	resp := admission.PatchResponseFromRaw(originObj, currentObj)
	respJson, err := json.Marshal(resp.AdmissionResponse)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	fmt.Println(string(respJson))
	return resp
}

