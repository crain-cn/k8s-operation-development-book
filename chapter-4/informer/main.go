package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"time"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main()  {
	// 通过 kubeconfig 生成配置
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/zhaoyu/.kube/config")
	if err != nil {
		panic(err)
	}

	// 创建 Clientset 对象
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 初始化 informer factory，每30s重新 List 一次
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second * 30)

	// 对 Deployment 监听
	deployInformer := informerFactory.Apps().V1().Deployments()

	// 创建 Informer, 相当于注册到工厂，这样下面启动的时候就会去 List & Watch 对应的资源
	informer := deployInformer.Informer()

	// 创建 Lister
	deployLister := deployInformer.Lister()

	// 注册事件处理程序
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deploy := obj.(*v1.Deployment)
			fmt.Println("add a deployment:", deploy.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldDeploy := oldObj.(*v1.Deployment)
			newDeploy := newObj.(*v1.Deployment)
			fmt.Println("update deployment:", oldDeploy.Name, newDeploy.Name)
		},
		DeleteFunc: func(obj interface{}) {
			deploy := obj.(*v1.Deployment)
			fmt.Println("delete a deployment:", deploy.Name)
		},
	})

	stopper := make(chan struct{})
	defer close(stopper)

	// 启动 informer，List & Watch
	informerFactory.Start(stopper)

	// 等待所有启动的 Informer 的缓存被同步
	informerFactory.WaitForCacheSync(stopper)

	// 从本地缓存中获取 default 中的所有 deployment 列表
	deployments, err := deployLister.Deployments("default").List(labels.Everything())
	if err != nil {
		panic(err)
	}
	for idx, deploy := range deployments {
		fmt.Printf("%d -> %s \n\r", idx, deploy.Name)
	}
	<-stopper
}
