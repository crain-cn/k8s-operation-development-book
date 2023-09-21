package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 通过 kubeconfig 生成配置
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/zhaoyu/.kube/config")
	if err != nil {
		panic(err)
	}

	// 创建 DiscoveryClient
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	_, apiResources, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		panic(err.Error())
	}

	for _, apiResource := range apiResources {
		gv, err := schema.ParseGroupVersion(apiResource.GroupVersion)
		if err != nil {
			panic(err.Error())
		}
		for _, resource := range apiResource.APIResources {
			fmt.Printf("[Group]%s [Version]%s [Resource]%s \n\r", gv.Group, gv.Version, resource.Name)
		}
	}
}