package main

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"strings"
	"sync"
	"time"
	"zhihu.com/cluster-mesh/controllers"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	clusterV1 "zhihu.com/cluster-mesh/api/clusters/v1"
	clusterVersionedV1 "zhihu.com/cluster-mesh/generated/generated/clientset/versioned"
	clusterInfor "zhihu.com/cluster-mesh/generated/generated/informers/externalversions"
)

var clusterMap map[string]*clusterV1.Cluster = make(map[string]*clusterV1.Cluster, 0)

// 事件存储结构
type EventHistory struct {
	Reason          string    `json:"reason" example:"NotFound"`
	Severity        string    `json:"severity" example:"warning"`
	Cluster         string    `json:"cluster" example:"k8s-dev"`
	Namespace       string    `json:"namespace" example:"kube-system"`
	ObjKind         string    `json:"obj_kind" example:"deployment"`
	ObjName         string    `json:"obj_name" example:"demo"`
	Datetime        time.Time `json:"datetime" example:"2023-11-12 22:00:00"`
	Message         string    `json:"message" example:"demo Not found"`
	SourceComponent string    `json:"source_component"`
	SourceHost      string    `json:"source_host"`
}

// 解析各个集群
func informerEvent(name string, cluser *clusterV1.Cluster) error {
	config, err := controllers.LoadConfig(cluser.Spec.KubeConfig)
	if err != nil {
		return err
	}
	// 创建 Clientset 对象
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second*30)
	eventInformer := informerFactory.Events().V1().Events()
	informer := eventInformer.Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			event := obj.(*v1.Event)
			fmt.Println(fmt.Sprintf("集群： %s, 事件：%v", name, obj))
			// 事件存储结构，下面内容可以存储到列式数据库中，这样可以做成功能，也可以存储到Prometheus中
			_ = EventHistory{
				Severity:        strings.ToLower(event.Type), // 事件类型
				Message:         event.Message,				  // 事件信息
				Reason:          event.Reason,				  // 事件原因
				Datetime:        event.CreationTimestamp.Time, // 事件创建时间
				Namespace:       event.InvolvedObject.Namespace, // 空间
				Cluster:         name,							 // 集群名称
				ObjKind:         event.InvolvedObject.Kind,		 // Kind
				ObjName:         event.InvolvedObject.Name,      // 名称
				SourceComponent: event.Source.Component,		 // 组件名称
				SourceHost:      event.Source.Host,				 // Host
			}

		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// oldEvent := oldObj.(*v1.Event)
			// newEvent := newObj.(*v1.Event)
			// 事件存储到库中
		},
	})

	stopper := make(chan struct{})
	defer close(stopper)

	// 启动 informer，List & Watch
	informerFactory.Start(stopper)

	// 等待所有启动的 Informer 的缓存被同步
	informerFactory.WaitForCacheSync(stopper)
	return nil
}

// 增加到集群map中处理
func addCluster(cluser *clusterV1.Cluster) {
	clusterMap[cluser.Name] = cluser
	go informerEvent(cluser.Name, clusterMap[cluser.Name])
}

// 更新集群map
func updateCluster(oldCluser, newCluster *clusterV1.Cluster) {
	delete(clusterMap, oldCluser.Name)
	clusterMap[newCluster.Name] = newCluster
	go informerEvent(newCluster.Name, clusterMap[newCluster.Name])
}

// 转换
func ObjToCluster(obj interface{}) *clusterV1.Cluster {
	cluster, ok := obj.(*clusterV1.Cluster)
	if ok {
		return cluster
	}
	deletedObj, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		cluster, ok := deletedObj.Obj.(*clusterV1.Cluster)
		if ok {
			return cluster
		}
	}
	return nil
}

func main() {
	var lock sync.RWMutex
	// 通过 kubeconfig 生成配置
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/zhaoyu/.kube/config")
	if err != nil {
		panic(err)
	}

	// 获取cluster ClientSet
	clusterClientSet, err := clusterVersionedV1.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	clusterInformerFactory := clusterInfor.NewSharedInformerFactory(clusterClientSet, time.Second*30)
	clusterInformer := clusterInformerFactory.Clusters().V1().Clusters()
	informer := clusterInformer.Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{ // 增加集群
		AddFunc: func(obj interface{}) {
			lock.Lock()
			cluster := ObjToCluster(obj)
			if cluster != nil {
				addCluster(cluster)
			}

			lock.Unlock()
		},
		UpdateFunc: func(oldObj, newObj interface{}) { // 更新集群
			lock.Lock()
			oldCluster := ObjToCluster(oldObj)
			newCluster := ObjToCluster(newObj)
			if oldCluster != nil && newCluster != nil {
				updateCluster(oldCluster, newCluster)
			}
			lock.Unlock()
		},
		DeleteFunc: func(obj interface{}) { // 删除集群
			lock.Lock()
			cluster, ok := obj.(*clusterV1.Cluster)
			if !ok {
				delete(clusterMap, cluster.Name)
			}
			lock.Unlock()
		},
	})

	stopper := make(chan struct{})
	defer close(stopper)

	// 启动 informer，List & Watch
	informer.Run(stopper)
}
