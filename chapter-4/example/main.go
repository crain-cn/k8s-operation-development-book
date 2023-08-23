package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var err error
	var config *rest.Config
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	// 使用 ServiceAccount 创建集群配置（InCluster模式）
	if config, err = rest.InClusterConfig(); err != nil {
		// 使用 KubeConfig 文件创建集群配置
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig);
		if err != nil {
			panic(err.Error())
		}
	}

	// 创建 clientset
	_, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}