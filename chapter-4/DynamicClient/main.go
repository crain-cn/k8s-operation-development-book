package main
import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/zhaoyu/.kube/config")
	if err != nil {
		panic(err.Error())
	}

	// 创建 DynamicClient
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 查询 kube-system 的 pod
	unstructuredList, err := dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}).Namespace("kube-system").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// 将 unstructuredList 的数据转成对应的结构体类型 corev1.PodList
	pods := &corev1.PodList{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(
		unstructuredList.UnstructuredContent(),
		pods,
	); err != nil {
		panic(err.Error())
	}

	for _, pod := range pods.Items {
		fmt.Printf("Pod =>  %s, namespace => %s \n\r", pod.Name, pod.Namespace)
	}
}

