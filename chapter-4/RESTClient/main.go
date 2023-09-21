package main
import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	config, err := clientcmd.BuildConfigFromFlags("", "/Users/zhaoyu/.kube/config")
	if err != nil {
		panic(err.Error())
	}
	config.APIPath = "apis"
	// 手动指定版本
	config.GroupVersion = &appsv1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs
	// 创建 clientset
	restclient, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err.Error())
	}

	result := &appsv1.DaemonSetList{}
	if err := restclient.
		Get().
		Namespace(metav1.NamespaceAll).
		Resource("daemonsets").
		Do(context.TODO()).
		Into(result); err != nil {
		panic(err)
	}


	// 遍历DaemonSet
	for _, item := range result.Items {
		fmt.Printf("DaemonSet =>  %s, namespace => %s \n\r", item.Name, item.Namespace)
	}
}
